package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/billybbuffum/budget/internal/application"
)

type CategoryGroupHandler struct {
	categoryGroupService *application.CategoryGroupService
}

func NewCategoryGroupHandler(categoryGroupService *application.CategoryGroupService) *CategoryGroupHandler {
	return &CategoryGroupHandler{categoryGroupService: categoryGroupService}
}

type CreateCategoryGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

type UpdateCategoryGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayOrder *int   `json:"display_order"`
}

type AssignCategoryRequest struct {
	CategoryID string `json:"category_id"`
	GroupID    string `json:"group_id"`
}

func (h *CategoryGroupHandler) CreateCategoryGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	group, err := h.categoryGroupService.CreateCategoryGroup(r.Context(), req.Name, req.Description, req.DisplayOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func (h *CategoryGroupHandler) GetCategoryGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category group id is required", http.StatusBadRequest)
		return
	}

	group, err := h.categoryGroupService.GetCategoryGroup(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *CategoryGroupHandler) ListCategoryGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.categoryGroupService.ListCategoryGroups(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func (h *CategoryGroupHandler) UpdateCategoryGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category group id is required", http.StatusBadRequest)
		return
	}

	var req UpdateCategoryGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	group, err := h.categoryGroupService.UpdateCategoryGroup(r.Context(), id, req.Name, req.Description, req.DisplayOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *CategoryGroupHandler) DeleteCategoryGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "category group id is required", http.StatusBadRequest)
		return
	}

	if err := h.categoryGroupService.DeleteCategoryGroup(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryGroupHandler) AssignCategoryToGroup(w http.ResponseWriter, r *http.Request) {
	var req AssignCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.CategoryID == "" || req.GroupID == "" {
		http.Error(w, "category_id and group_id are required", http.StatusBadRequest)
		return
	}

	if err := h.categoryGroupService.AssignCategoryToGroup(r.Context(), req.CategoryID, req.GroupID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryGroupHandler) UnassignCategoryFromGroup(w http.ResponseWriter, r *http.Request) {
	categoryID := r.PathValue("id")
	if categoryID == "" {
		http.Error(w, "category id is required", http.StatusBadRequest)
		return
	}

	if err := h.categoryGroupService.UnassignCategoryFromGroup(r.Context(), categoryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
