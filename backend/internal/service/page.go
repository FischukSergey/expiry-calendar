package service

import (
	"duekeep/internal/model"
)

const (
	fieldPage    = "page"
	fieldPerPage = "per_page"
)

// NormalizePage приводит page и per_page к диапазону 1 / 1..100 (дефолт 20).
func NormalizePage(page, perPage int) (model.Page, error) {
	if page < 0 {
		return model.Page{}, model.Validation("invalid page", map[string]any{fieldPage: ">= 1"})
	}
	if perPage < 0 {
		return model.Page{}, model.Validation("invalid per_page", map[string]any{fieldPerPage: ">= 1"})
	}
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = model.DefaultPerPage
	}
	if perPage > model.MaxPerPage {
		perPage = model.MaxPerPage
	}
	return model.Page{Page: page, PerPage: perPage}, nil
}
