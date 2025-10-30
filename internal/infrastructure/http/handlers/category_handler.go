package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
)

type CategoryHandler struct {
	categoryService *application.CategoryService
}

func NewCategoryHandler(categoryService *application.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	GroupID     *string `json:"group_id"`
}

type UpdateCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	GroupID     *string `json:"group_id"`
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.CreateCategory(r.Context(), req.Name, req.Description, req.Color, req.GroupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category id is required", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.GetCategory(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.ListCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category id is required", http.StatusBadRequest)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.UpdateCategory(r.Context(), id, req.Name, req.Description, req.Color, req.GroupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category id is required", http.StatusBadRequest)
		return
	}

	if err := h.categoryService.DeleteCategory(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) RestoreCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category id is required", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.RestoreCategory(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}
