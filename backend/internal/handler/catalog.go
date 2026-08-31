package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"duekeep/internal/middleware"
	"duekeep/internal/model"
)

const jsonNull = "null"

type kindWrite struct {
	Slug       string            `json:"slug"`
	Name       string            `json:"name"`
	Color      string            `json:"color"`
	AttrSchema []model.AttrField `json:"attr_schema"`
}

type kindPatchBody struct {
	Slug       *string            `json:"slug"`
	Name       *string            `json:"name"`
	Color      *string            `json:"color"`
	AttrSchema *[]model.AttrField `json:"attr_schema"`
}

type categoryWrite struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name"`
	SortOrder int     `json:"sort_order"`
}

type categoryPatchBody struct {
	Name      *string         `json:"name"`
	SortOrder *int            `json:"sort_order"`
	ParentID  json.RawMessage `json:"parent_id"`
}

func (a *API) listKinds(w http.ResponseWriter, r *http.Request) {
	items, err := a.kinds.List(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if items == nil {
		items = []model.Kind{}
	}
	writeBytes(w, http.StatusOK, model.KindList{Items: items})
}

func (a *API) createKind(w http.ResponseWriter, r *http.Request) {
	var body kindWrite
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	k, err := a.kinds.Create(r.Context(), model.Kind{
		Slug: body.Slug, Name: body.Name, Color: body.Color, AttrSchema: body.AttrSchema,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusCreated, k)
}

func (a *API) patchKind(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body kindPatchBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	k, err := a.kinds.Patch(r.Context(), id, model.KindPatch{
		Slug: body.Slug, Name: body.Name, Color: body.Color, AttrSchema: body.AttrSchema,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, k)
}

func (a *API) deleteKind(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.kinds.Delete(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := a.categories.List(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if items == nil {
		items = []model.Category{}
	}
	writeBytes(w, http.StatusOK, model.CategoryList{Items: items})
}

func (a *API) createCategory(w http.ResponseWriter, r *http.Request) {
	var body categoryWrite
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	c, err := a.categories.Create(r.Context(), body.ParentID, body.Name, body.SortOrder, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusCreated, c)
}

func (a *API) patchCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body categoryPatchBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	p := model.CategoryPatch{Name: body.Name, SortOrder: body.SortOrder}
	if body.ParentID != nil {
		p.SetParent = true
		if string(body.ParentID) != jsonNull {
			var parent string
			if err := json.Unmarshal(body.ParentID, &parent); err != nil {
				writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid parent_id")
				return
			}
			p.ParentID = &parent
		}
	}
	c, err := a.categories.Patch(r.Context(), id, p, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, c)
}

func (a *API) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.categories.Delete(r.Context(), id, middleware.UserID(r.Context())); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid id")
		return "", false
	}
	return id, true
}
