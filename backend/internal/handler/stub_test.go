package handler_test

import (
	"context"

	"duekeep/internal/model"
)

type nopKinds struct{}

func (nopKinds) List(context.Context) ([]model.Kind, error) { return nil, nil }

func (nopKinds) Create(context.Context, model.Kind) (model.Kind, error) {
	return model.Kind{}, model.ErrForbidden
}

func (nopKinds) Patch(context.Context, string, model.KindPatch) (model.Kind, error) {
	return model.Kind{}, model.ErrForbidden
}

func (nopKinds) Delete(context.Context, string) error { return model.ErrForbidden }

type nopCategories struct{}

func (nopCategories) List(context.Context) ([]model.Category, error) { return nil, nil }

func (nopCategories) Create(context.Context, *string, string, int) (model.Category, error) {
	return model.Category{}, model.ErrForbidden
}

func (nopCategories) Patch(context.Context, string, model.CategoryPatch) (model.Category, error) {
	return model.Category{}, model.ErrForbidden
}

func (nopCategories) Delete(context.Context, string) error { return model.ErrForbidden }

type nopItems struct{}

func (nopItems) List(context.Context, model.ItemFilter, model.Page) (model.ItemList, error) {
	return model.ItemList{Items: []model.Item{}}, nil
}

func (nopItems) Create(context.Context, model.Item, string) (model.Item, error) {
	return model.Item{}, model.ErrForbidden
}

func (nopItems) Get(context.Context, string) (model.ItemCard, error) {
	return model.ItemCard{}, model.ErrNotFound
}

func (nopItems) Patch(context.Context, string, model.ItemPatch, string) (model.Item, error) {
	return model.Item{}, model.ErrForbidden
}

func (nopItems) Delete(context.Context, string, string) error { return model.ErrForbidden }

func (nopItems) Renew(context.Context, string, model.RenewInput, string) (model.Item, error) {
	return model.Item{}, model.ErrForbidden
}

func (nopItems) Bulk(context.Context, model.BulkInput, string) (model.BulkResult, error) {
	return model.BulkResult{}, model.ErrForbidden
}

func (nopItems) ListAudit(context.Context, model.Page) (model.AuditList, error) {
	return model.AuditList{Items: []model.AuditEntry{}}, nil
}

func (nopItems) Export(context.Context, model.ItemFilter) ([]byte, error) {
	return []byte("id,title\n"), nil
}

func (nopItems) Import(
	context.Context, []byte, map[string]string, bool, string,
) (model.CSVImportPreview, model.CSVImportResult, error) {
	return model.CSVImportPreview{
		Errors: []model.CSVImportError{}, Preview: []model.CSVPreviewRow{},
	}, model.CSVImportResult{}, nil
}

type nopNotifications struct{}

func (nopNotifications) List(context.Context, bool, model.Page) (model.NotificationList, error) {
	return model.NotificationList{Items: []model.Notification{}}, nil
}

func (nopNotifications) MarkRead(context.Context, string) error { return model.ErrNotFound }

func (nopNotifications) MarkAllRead(context.Context) error { return nil }
