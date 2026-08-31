package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

const catOwner = "owner"

func TestValidateAttrSchema(t *testing.T) {
	t.Parallel()
	ok := []model.AttrField{{Key: "vin", Label: "VIN", Type: model.AttrString}}
	if err := service.ValidateAttrSchema(ok); err != nil {
		t.Fatal(err)
	}
	badType := []model.AttrField{{Key: "x", Label: "X", Type: "date"}}
	if err := service.ValidateAttrSchema(badType); !errors.Is(err, model.ErrValidation) {
		t.Fatalf("got %v", err)
	}
	dup := []model.AttrField{
		{Key: "a", Label: "A", Type: model.AttrString},
		{Key: "a", Label: "B", Type: model.AttrNumber},
	}
	if err := service.ValidateAttrSchema(dup); !errors.Is(err, model.ErrValidation) {
		t.Fatalf("got %v", err)
	}
}

func TestCategoryDepthAndCreateLimit(t *testing.T) {
	t.Parallel()
	root := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	child := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
	grand := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3"
	rows := []model.Category{
		{ID: root, OwnerID: catOwner, Name: "A", Children: []model.Category{}},
		{ID: child, OwnerID: catOwner, ParentID: &root, Name: "B", Children: []model.Category{}},
		{ID: grand, OwnerID: catOwner, ParentID: &child, Name: "C", Children: []model.Category{}},
	}
	if d := service.CategoryDepth(rows, root); d != 1 {
		t.Fatalf("root %d", d)
	}
	if d := service.CategoryDepth(rows, grand); d != 3 {
		t.Fatalf("grand %d", d)
	}
	if d := service.CategoryDepth(rows, "missing"); d != -1 {
		t.Fatalf("missing %d", d)
	}
	loopA, loopB := "loop-a", "loop-b"
	cycled := []model.Category{
		{ID: loopA, ParentID: &loopB, Name: "A"},
		{ID: loopB, ParentID: &loopA, Name: "B"},
	}
	if d := service.CategoryDepth(cycled, loopA); d != -1 {
		t.Fatalf("cycle %d", d)
	}

	store := newMemCats(rows)
	svc := service.NewCategory(store)
	if _, err := svc.Create(t.Context(), &grand, "too-deep", 0, catOwner); !errors.Is(err, model.ErrValidation) {
		t.Fatalf("want 422 depth, got %v", err)
	}
	if _, err := svc.Create(t.Context(), &child, "ok-leaf", 1, catOwner); err != nil {
		t.Fatal(err)
	}
}

func TestCategoryMoveCycle(t *testing.T) {
	t.Parallel()
	root := "cccccccc-cccc-cccc-cccc-ccccccccccc1"
	child := "cccccccc-cccc-cccc-cccc-ccccccccccc2"
	store := newMemCats([]model.Category{
		{ID: root, OwnerID: catOwner, Name: "R"},
		{ID: child, OwnerID: catOwner, ParentID: &root, Name: "C"},
	})
	svc := service.NewCategory(store)
	_, err := svc.Patch(t.Context(), root, model.CategoryPatch{SetParent: true, ParentID: &child}, catOwner)
	if !errors.Is(err, model.ErrValidation) {
		t.Fatalf("want cycle, got %v", err)
	}
	_, err = svc.Patch(t.Context(), child, model.CategoryPatch{SetParent: true, ParentID: &child}, catOwner)
	if !errors.Is(err, model.ErrValidation) {
		t.Fatalf("want self-parent, got %v", err)
	}
}

func TestDeleteCategoryWithChildren(t *testing.T) {
	t.Parallel()
	root := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1"
	child := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2"
	store := newMemCats([]model.Category{
		{ID: root, OwnerID: catOwner, Name: "R"},
		{ID: child, OwnerID: catOwner, ParentID: &root, Name: "C"},
	})
	svc := service.NewCategory(store)
	if err := svc.Delete(t.Context(), root, catOwner); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("got %v", err)
	}
	if err := svc.Delete(t.Context(), child, catOwner); err != nil {
		t.Fatal(err)
	}
}

func TestCategoryListOwnOnly(t *testing.T) {
	t.Parallel()
	mine := "cccccccc-cccc-cccc-cccc-ccccccccccc3"
	store := newMemCats([]model.Category{
		{ID: mine, OwnerID: catOwner, Name: "Mine"},
		{ID: "dddddddd-dddd-dddd-dddd-ddddddddddd1", OwnerID: otherOwner, Name: "Theirs"},
	})
	got, err := service.NewCategory(store).List(t.Context(), catOwner)
	if err != nil || len(got) != 1 || got[0].ID != mine {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestKindRejectsBadSchema(t *testing.T) {
	t.Parallel()
	svc := service.NewKind(newMemKinds())
	_, err := svc.Create(t.Context(), model.Kind{
		Slug: "visa", Name: "Виза", Color: "#111",
		AttrSchema: []model.AttrField{{Key: "n", Label: "N", Type: "date"}},
	})
	if !errors.Is(err, model.ErrValidation) {
		t.Fatalf("got %v", err)
	}
}

type memKinds struct {
	mu   sync.Mutex
	byID map[string]model.Kind
}

func newMemKinds() *memKinds {
	return &memKinds{byID: map[string]model.Kind{}}
}

func (m *memKinds) List(context.Context) ([]model.Kind, error) { return nil, nil }

func (m *memKinds) ByID(_ context.Context, id string) (model.Kind, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.byID[id]
	if !ok {
		return model.Kind{}, model.ErrNotFound
	}
	return k, nil
}

func (m *memKinds) Create(_ context.Context, k model.Kind) (model.Kind, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k.ID = "kind-1"
	m.byID[k.ID] = k
	return k, nil
}

func (m *memKinds) Update(_ context.Context, k model.Kind) (model.Kind, error) { return k, nil }

func (m *memKinds) Delete(context.Context, string) error { return nil }

func (m *memKinds) CountItems(context.Context, string) (int, error) { return 0, nil }

type memCats struct {
	mu   sync.Mutex
	rows []model.Category
}

func newMemCats(rows []model.Category) *memCats {
	cp := append([]model.Category(nil), rows...)
	return &memCats{rows: cp}
}

func (m *memCats) List(_ context.Context, ownerID string) ([]model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Category, 0, len(m.rows))
	for _, r := range m.rows {
		if r.OwnerID == ownerID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *memCats) ByID(_ context.Context, id string) (model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.ID == id {
			return r, nil
		}
	}
	return model.Category{}, model.ErrNotFound
}

func (m *memCats) Create(_ context.Context, c model.Category) (model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.ID = "new-cat"
	c.Children = []model.Category{}
	m.rows = append(m.rows, c)
	return c, nil
}

func (m *memCats) Update(_ context.Context, c model.Category) (model.Category, error) { return c, nil }

func (m *memCats) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.rows[:0]
	for _, r := range m.rows {
		if r.ID != id {
			out = append(out, r)
		}
	}
	m.rows = out
	return nil
}

func (m *memCats) CountChildren(_ context.Context, id string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, r := range m.rows {
		if r.ParentID != nil && *r.ParentID == id {
			n++
		}
	}
	return n, nil
}

func (m *memCats) CountItems(context.Context, string) (int, error) { return 0, nil }

func (m *memCats) DescendantIDs(_ context.Context, id string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append([]model.Category(nil), m.rows...)
	byParent := map[string][]string{}
	for _, r := range rows {
		if r.ParentID != nil && *r.ParentID != "" {
			byParent[*r.ParentID] = append(byParent[*r.ParentID], r.ID)
		}
	}
	out := []string{id}
	var walk func(string)
	walk = func(parent string) {
		for _, ch := range byParent[parent] {
			out = append(out, ch)
			walk(ch)
		}
	}
	walk(id)
	return out, nil
}
