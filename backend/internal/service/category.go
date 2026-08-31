package service

import (
	"context"
	"strings"

	"duekeep/internal/model"
)

// CategoryStore — плоские строки categories.
type CategoryStore interface {
	List(ctx context.Context, ownerID string) ([]model.Category, error)
	ByID(ctx context.Context, id string) (model.Category, error)
	Create(ctx context.Context, c model.Category) (model.Category, error)
	Update(ctx context.Context, c model.Category) (model.Category, error)
	Delete(ctx context.Context, id string) error
	CountChildren(ctx context.Context, id string) (int, error)
	CountItems(ctx context.Context, id string) (int, error)
	DescendantIDs(ctx context.Context, id string) ([]string, error)
}

// Category — дерево, глубина ≤ 3, цикл и непустой DELETE.
type Category struct {
	store CategoryStore
}

// NewCategory собирает сервис категорий.
func NewCategory(store CategoryStore) *Category {
	return &Category{store: store}
}

// List дерево корней владельца с вложенными children.
func (s *Category) List(ctx context.Context, ownerID string) ([]model.Category, error) {
	rows, err := s.store.List(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return buildTree(rows), nil
}

// Create проверяет глубину нового узла (height=1).
func (s *Category) Create(
	ctx context.Context, parentID *string, name string, sortOrder int, ownerID string,
) (model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Category{}, model.Validation(msgInvalidName, map[string]any{fieldName: detailRequired})
	}
	rows, err := s.store.List(ctx, ownerID)
	if err != nil {
		return model.Category{}, err
	}
	if err := checkNewDepth(rows, parentID); err != nil {
		return model.Category{}, err
	}
	return s.store.Create(ctx, model.Category{OwnerID: ownerID, ParentID: parentID, Name: name, SortOrder: sortOrder})
}

// Patch имя/порядок/родитель. Смена родителя — те же инварианты + запрет цикла.
func (s *Category) Patch(ctx context.Context, id string, p model.CategoryPatch, actorID string) (model.Category, error) {
	cur, err := s.store.ByID(ctx, id)
	if err != nil {
		return model.Category{}, err
	}
	if err := requireOwner(cur.OwnerID, actorID); err != nil {
		return model.Category{}, err
	}
	if p.Name != nil {
		name := strings.TrimSpace(*p.Name)
		if name == "" {
			return model.Category{}, model.Validation(msgInvalidName, map[string]any{fieldName: detailRequired})
		}
		cur.Name = name
	}
	if p.SortOrder != nil {
		cur.SortOrder = *p.SortOrder
	}
	if p.SetParent {
		rows, err := s.store.List(ctx, actorID)
		if err != nil {
			return model.Category{}, err
		}
		if err := checkMove(rows, id, p.ParentID); err != nil {
			return model.Category{}, err
		}
		cur.ParentID = p.ParentID
	}
	return s.store.Update(ctx, cur)
}

// Delete → 409, если есть дети или items.
func (s *Category) Delete(ctx context.Context, id, actorID string) error {
	cur, err := s.store.ByID(ctx, id)
	if err != nil {
		return err
	}
	if err := requireOwner(cur.OwnerID, actorID); err != nil {
		return err
	}
	n, err := s.store.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return model.ErrConflict
	}
	items, err := s.store.CountItems(ctx, id)
	if err != nil {
		return err
	}
	if items > 0 {
		return model.ErrConflict
	}
	return s.store.Delete(ctx, id)
}

func checkNewDepth(rows []model.Category, parentID *string) error {
	if parentID == nil || *parentID == "" {
		return nil
	}
	d := CategoryDepth(rows, *parentID)
	if d < 1 {
		return model.ErrNotFound
	}
	if d+1 > model.MaxCategoryDepth {
		return model.Validation("category depth > 3", map[string]any{fieldParentID: *parentID})
	}
	return nil
}

func checkMove(rows []model.Category, id string, newParent *string) error {
	if newParent != nil && *newParent == id {
		return model.Validation("cycle", map[string]any{fieldParentID: id})
	}
	if newParent != nil && *newParent != "" {
		if CategoryDepth(rows, *newParent) < 1 {
			return model.ErrNotFound
		}
		if isAncestor(rows, id, *newParent) {
			return model.Validation("cycle", map[string]any{fieldParentID: *newParent})
		}
	}
	parentDepth := 0
	if newParent != nil && *newParent != "" {
		parentDepth = CategoryDepth(rows, *newParent)
	}
	if parentDepth+subtreeHeight(rows, id) > model.MaxCategoryDepth {
		return model.Validation("category depth > 3", map[string]any{"id": id})
	}
	return nil
}

// CategoryDepth — уровни от узла к корню (корень = 1). Цикл или нет узла → -1.
func CategoryDepth(rows []model.Category, id string) int {
	byID := indexCats(rows)
	cur, ok := byID[id]
	if !ok {
		return -1
	}
	depth := 1
	seen := map[string]struct{}{id: {}}
	for cur.ParentID != nil && *cur.ParentID != "" {
		if _, loop := seen[*cur.ParentID]; loop {
			return -1
		}
		seen[*cur.ParentID] = struct{}{}
		next, ok := byID[*cur.ParentID]
		if !ok {
			return -1
		}
		depth++
		cur = next
	}
	return depth
}

func subtreeHeight(rows []model.Category, id string) int {
	byParent := childrenOf(rows)
	var walk func(string) int
	walk = func(node string) int {
		maxC := 0
		for _, ch := range byParent[node] {
			maxC = max(maxC, walk(ch.ID))
		}
		return 1 + maxC
	}
	return walk(id)
}

func isAncestor(rows []model.Category, ancestor, node string) bool {
	byID := indexCats(rows)
	cur, ok := byID[node]
	if !ok {
		return false
	}
	seen := map[string]struct{}{}
	for cur.ParentID != nil && *cur.ParentID != "" {
		if *cur.ParentID == ancestor {
			return true
		}
		if _, loop := seen[*cur.ParentID]; loop {
			return false
		}
		seen[*cur.ParentID] = struct{}{}
		next, ok := byID[*cur.ParentID]
		if !ok {
			return false
		}
		cur = next
	}
	return false
}

func buildTree(rows []model.Category) []model.Category {
	byParent := childrenOf(rows)
	var attach func(model.Category) model.Category
	attach = func(n model.Category) model.Category {
		kids := byParent[n.ID]
		n.Children = make([]model.Category, 0, len(kids))
		for _, ch := range kids {
			n.Children = append(n.Children, attach(ch))
		}
		return n
	}
	roots := make([]model.Category, 0)
	for _, r := range rows {
		if r.ParentID == nil || *r.ParentID == "" {
			roots = append(roots, attach(r))
		}
	}
	return roots
}

func indexCats(rows []model.Category) map[string]model.Category {
	m := make(map[string]model.Category, len(rows))
	for _, r := range rows {
		m[r.ID] = r
	}
	return m
}

func childrenOf(rows []model.Category) map[string][]model.Category {
	m := make(map[string][]model.Category)
	for _, r := range rows {
		if r.ParentID != nil && *r.ParentID != "" {
			m[*r.ParentID] = append(m[*r.ParentID], r)
		}
	}
	return m
}
