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
