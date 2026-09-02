package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"

	"duekeep/internal/model"
)

const (
	csvFieldTitle     = "title"
	csvFieldKindSlug  = "kind_slug"
	csvFieldExpiresAt = "expires_at"
	csvFieldCost      = "cost_amount"
	csvFieldCurrency  = "currency"
	csvFieldVendor    = "vendor"
	csvFieldBilling   = "billing_period"
	csvFieldCategory  = "category_name"
	csvFieldTags      = "tags"
	csvFieldID        = "id"
	csvFieldStatus    = "status"
	csvFieldNotify    = "notify_before_days"
	csvAttrsPrefix    = "attrs."
	csvPreviewMax     = 20
	fieldMapping      = "mapping"
	fieldFile         = "file"
)

var csvMappedFields = []string{
	csvFieldTitle, csvFieldKindSlug, csvFieldExpiresAt, csvFieldCost,
	csvFieldCurrency, csvFieldVendor, csvFieldBilling, csvFieldCategory, csvFieldTags,
	csvFieldStatus, csvFieldNotify,
}

// NormalizeCSVMapping проверяет ключи: поля записи и attrs.*. Значения — имена колонок CSV.
func NormalizeCSVMapping(in map[string]string) (map[string]string, error) {
	if len(in) == 0 {
		return nil, model.Validation("mapping required", map[string]any{fieldMapping: detailRequired})
	}
	out := make(map[string]string, len(in))
	for field, col := range in {
		field = strings.TrimSpace(field)
		col = strings.TrimSpace(col)
		if field == "" || col == "" {
			return nil, model.Validation("invalid mapping", map[string]any{fieldMapping: field})
		}
		if strings.HasPrefix(field, csvAttrsPrefix) {
			key := strings.TrimSpace(field[len(csvAttrsPrefix):])
			if key == "" {
				return nil, model.Validation("invalid mapping", map[string]any{fieldMapping: field})
			}
			out[csvAttrsPrefix+key] = col
			continue
		}
		if !slices.Contains(csvMappedFields, field) {
			return nil, model.Validation("unknown mapping field", map[string]any{fieldMapping: field})
		}
		out[field] = col
	}
	return out, nil
}

// ParseCSVAttr приводит непустую ячейку к типу схемы.
func ParseCSVAttr(typ, raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	switch typ {
	case model.AttrString:
		return strings.Clone(raw), nil
	case model.AttrNumber:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, model.Validation("invalid attr type", map[string]any{fieldKey: typ, fieldWant: model.AttrNumber})
		}
		return n, nil
	case model.AttrBoolean:
		switch strings.ToLower(raw) {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return nil, model.Validation("invalid attr type", map[string]any{fieldKey: typ, fieldWant: model.AttrBoolean})
		}
	default:
		return nil, model.Validation("invalid attr type", map[string]any{fieldKey: typ})
	}
}

// Export пишет CSV текущего фильтра. Только свой owner_id; потолок MaxCSVExport.
func (s *Item) Export(ctx context.Context, f model.ItemFilter, actorID string) ([]byte, error) {
	f, err := s.scopedFilter(ctx, f, actorID)
	if err != nil {
		return nil, err
	}
	rows, _, err := s.items.List(ctx, f, model.Page{Page: 1, PerPage: model.MaxCSVExport})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []model.Item{}
	}
	kinds, err := s.kinds.List(ctx)
	if err != nil {
		return nil, err
	}
	kindByID := make(map[string]model.Kind, len(kinds))
	for _, k := range kinds {
		kindByID[k.ID] = k
	}
	cats, err := s.cats.List(ctx, actorID)
	if err != nil {
		return nil, err
	}
	catByID := make(map[string]model.Category, len(cats))
	for _, c := range cats {
		catByID[c.ID] = c
	}
	attrKeys := exportAttrKeys(rows, kindByID)
	return writeItemsCSV(rows, kindByID, catByID, attrKeys)
}

// Import читает CSV с маппингом колонок. dry_run не пишет; иначе одна транзакция и audit import.
func (s *Item) Import(
	ctx context.Context,
	csvData []byte,
	mapping map[string]string,
	dryRun bool,
	actorID string,
) (model.CSVImportPreview, model.CSVImportResult, error) {
	mapping, err := NormalizeCSVMapping(mapping)
	if err != nil {
		return model.CSVImportPreview{}, model.CSVImportResult{}, err
	}
	records, err := readCSVRecords(csvData)
	if err != nil {
		return model.CSVImportPreview{}, model.CSVImportResult{}, err
	}
	if len(records) == 0 {
		return model.CSVImportPreview{}, model.CSVImportResult{},
			model.Validation("empty csv", map[string]any{fieldFile: detailRequired})
	}
	header := records[0]
	col := csvHeaderIndex(header)
	for _, csvCol := range mapping {
		if _, ok := col[csvCol]; !ok {
			return model.CSVImportPreview{}, model.CSVImportResult{},
				model.Validation("unknown csv column", map[string]any{fieldMapping: csvCol})
		}
	}
	dataRows := records[1:]
	if len(dataRows) > model.MaxCSVImport {
		return model.CSVImportPreview{}, model.CSVImportResult{}, model.Validation("too many rows", map[string]any{
			"rows": len(dataRows), "max": model.MaxCSVImport,
		})
	}

	kinds, err := s.kinds.List(ctx)
	if err != nil {
		return model.CSVImportPreview{}, model.CSVImportResult{}, err
	}
	kindBySlug := make(map[string]model.Kind, len(kinds))
	for _, k := range kinds {
		kindBySlug[k.Slug] = k
	}
	cats, err := s.cats.List(ctx, actorID)
	if err != nil {
		return model.CSVImportPreview{}, model.CSVImportResult{}, err
	}
	catByName := make(map[string]model.Category, len(cats))
	for _, c := range cats {
		catByName[c.Name] = c
	}

	preview := model.CSVImportPreview{
		Rows:    len(dataRows),
		Errors:  []model.CSVImportError{},
		Preview: []model.CSVPreviewRow{},
	}
	prepared := make([]model.Item, 0, len(dataRows))
	for i, rec := range dataRows {
		line := i + 2
		it, prev, rowErr := s.rowToItem(ctx, rec, col, mapping, kindBySlug, catByName, actorID)
		if rowErr != "" {
			preview.Errors = append(preview.Errors, model.CSVImportError{Line: line, Message: rowErr})
			continue
		}
		preview.Valid++
		if len(preview.Preview) < csvPreviewMax {
			preview.Preview = append(preview.Preview, prev)
		}
		prepared = append(prepared, it)
	}
	if dryRun {
		return preview, model.CSVImportResult{}, nil
	}
	if len(preview.Errors) > 0 {
		return preview, model.CSVImportResult{}, model.Validation("import invalid", map[string]any{
			"errors": preview.Errors,
		})
	}
	created, err := s.writeImport(ctx, prepared, actorID)
	if err != nil {
		return preview, model.CSVImportResult{}, err
	}
	return preview, model.CSVImportResult{Created: created}, nil
}

func (s *Item) writeImport(ctx context.Context, items []model.Item, actorID string) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}
	ids := make([]string, 0, len(items))
	err := s.tx(ctx, func(ctx context.Context) error {
		for _, it := range items {
			it.OwnerID = actorID
			created, err := s.items.Create(ctx, it)
			if err != nil {
				return err
			}
			ids = append(ids, created.ID)
		}
		after, merr := json.Marshal(struct {
			Created int      `json:"created"`
			IDs     []string `json:"ids"`
		}{Created: len(ids), IDs: ids})
		if merr != nil {
			return merr
		}
		return s.audit.Create(ctx, auditEntry(actorID, model.AuditImport, ids[0], nil, after))
	})
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

func (s *Item) rowToItem(
	ctx context.Context,
	rec []string,
	col map[string]int,
	mapping map[string]string,
	kindBySlug map[string]model.Kind,
	catByName map[string]model.Category,
	actorID string,
) (model.Item, model.CSVPreviewRow, string) {
	cell := func(field string) string {
		name, ok := mapping[field]
		if !ok {
			return ""
		}
		idx, ok := col[name]
		if !ok || idx >= len(rec) {
			return ""
		}
		return strings.Clone(strings.TrimSpace(rec[idx]))
	}

	title := cell(csvFieldTitle)
	kindSlug := cell(csvFieldKindSlug)
	expires := cell(csvFieldExpiresAt)
	prev := model.CSVPreviewRow{Title: title, KindSlug: kindSlug, ExpiresAt: expires}

	if title == "" {
		return model.Item{}, prev, "title required"
	}
	if kindSlug == "" {
		return model.Item{}, prev, "kind_slug required"
	}
	if expires == "" {
		return model.Item{}, prev, "expires_at required"
	}
	kind, ok := kindBySlug[kindSlug]
	if !ok {
		return model.Item{}, prev, "unknown kind_slug"
	}

	cost := 0
	if raw := cell(csvFieldCost); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return model.Item{}, prev, "invalid cost_amount"
		}
		cost = n
	}

	var categoryID *string
	if name := cell(csvFieldCategory); name != "" {
		c, ok := catByName[name]
		if !ok {
			return model.Item{}, prev, "unknown category_name"
		}
		id := c.ID
		categoryID = &id
	}

	attrs := map[string]any{}
	schemaByKey := make(map[string]model.AttrField, len(kind.AttrSchema))
	for _, f := range kind.AttrSchema {
		schemaByKey[f.Key] = f
	}
	for field := range mapping {
		key, ok := strings.CutPrefix(field, csvAttrsPrefix)
		if !ok {
			continue
		}
		raw := cell(field)
		if raw == "" {
			continue
		}
		spec, known := schemaByKey[key]
		if !known {
			return model.Item{}, prev, "unknown attr"
		}
		val, err := ParseCSVAttr(spec.Type, raw)
		if err != nil {
			return model.Item{}, prev, "invalid attr type"
		}
		attrs[key] = val
	}
	prev.Attrs = attrs

	tags := splitCSVTags(cell(csvFieldTags))
	days, notifyOff, err := ParseCSVNotify(cell(csvFieldNotify), mappingHas(mapping, csvFieldNotify))
	if err != nil {
		return model.Item{}, prev, "invalid notify_before_days"
	}
	var notify *int
	if !notifyOff {
		notify = model.Ptr(days)
	}
	it := model.Item{
		Title: title, KindID: kind.ID, CategoryID: categoryID,
		Vendor: cell(csvFieldVendor), Tags: tags, CostAmount: cost,
		Currency: cell(csvFieldCurrency), BillingPeriod: cell(csvFieldBilling),
		ExpiresAt: expires, NotifyBeforeDays: notify, Attrs: attrs,
		Status: cell(csvFieldStatus),
	}
	prepared, err := s.prepareWrite(ctx, it, actorID)
	if err != nil {
		return model.Item{}, prev, csvRowMessage(err)
	}
	return prepared, prev, ""
}

func mappingHas(mapping map[string]string, field string) bool {
	_, ok := mapping[field]
	return ok
}

// ParseCSVNotify читает колонку дней. mapped=false — дефолт 30; пусто/off/- → off=true.
func ParseCSVNotify(raw string, mapped bool) (days int, off bool, err error) {
	if !mapped {
		return model.DefaultNotifyDays, false, nil
	}
	raw = strings.TrimSpace(raw)
	if csvNotifyOff(raw) {
		return 0, true, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, err
	}
	if n < 0 {
		return 0, false, strconv.ErrRange
	}
	return n, false, nil
}

func csvNotifyOff(raw string) bool {
	return raw == "" || strings.EqualFold(raw, "off") || raw == "-"
}

func csvRowMessage(err error) string {
	var val *model.ValidationError
	if errors.As(err, &val) {
		return val.Msg
	}
	if err != nil {
		return err.Error()
	}
	return "invalid row"
}

func readCSVRecords(data []byte) ([][]string, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, model.Validation("empty csv", map[string]any{fieldFile: detailRequired})
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.ReuseRecord = false
	r.LazyQuotes = true
	out := make([][]string, 0)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, model.Validation("invalid csv", map[string]any{fieldFile: err.Error()})
		}
		out = append(out, rec)
	}
}

func csvHeaderIndex(header []string) map[string]int {
	out := make(map[string]int, len(header))
	for i, h := range header {
		out[strings.TrimSpace(h)] = i
	}
	return out
}

func splitCSVTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.Clone(p))
		}
	}
	return out
}

func exportAttrKeys(rows []model.Item, kindByID map[string]model.Kind) []string {
	seen := map[string]struct{}{}
	for _, it := range rows {
		k, ok := kindByID[it.KindID]
		if !ok {
			continue
		}
		for _, f := range k.AttrSchema {
			seen[f.Key] = struct{}{}
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func writeItemsCSV(
	rows []model.Item,
	kindByID map[string]model.Kind,
	catByID map[string]model.Category,
	attrKeys []string,
) ([]byte, error) {
	header := make([]string, 0, 11+len(attrKeys))
	header = append(header,
		csvFieldID, csvFieldTitle, csvFieldKindSlug, csvFieldStatus, csvFieldExpiresAt,
		csvFieldNotify, csvFieldCost, csvFieldCurrency, csvFieldVendor, csvFieldBilling, csvFieldCategory, csvFieldTags,
	)
	for _, key := range attrKeys {
		header = append(header, csvAttrsPrefix+key)
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(header); err != nil {
		return nil, err
	}
	for _, it := range rows {
		kindSlug := ""
		if k, ok := kindByID[it.KindID]; ok {
			kindSlug = k.Slug
		}
		catName := ""
		if it.CategoryID != nil {
			if c, ok := catByID[*it.CategoryID]; ok {
				catName = c.Name
			}
		}
		rec := []string{
			it.ID, it.Title, kindSlug, it.Status, it.ExpiresAt,
			formatCSVNotify(it.NotifyBeforeDays),
			strconv.Itoa(it.CostAmount), it.Currency, it.Vendor, it.BillingPeriod,
			catName, strings.Join(it.Tags, ","),
		}
		for _, key := range attrKeys {
			rec = append(rec, formatCSVAttr(it.Attrs[key]))
		}
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatCSVNotify(days *int) string {
	if days == nil {
		return ""
	}
	return strconv.Itoa(*days)
}

func formatCSVAttr(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		if t == float64(int(t)) {
			return strconv.Itoa(int(t))
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
