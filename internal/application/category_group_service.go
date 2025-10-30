package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// CategoryGroupService handles category group-related business logic
type CategoryGroupService struct {
	categoryGroupRepo domain.CategoryGroupRepository
	categoryRepo      domain.CategoryRepository
}

// NewCategoryGroupService creates a new category group service
func NewCategoryGroupService(categoryGroupRepo domain.CategoryGroupRepository, categoryRepo domain.CategoryRepository) *CategoryGroupService {
	return &CategoryGroupService{
		categoryGroupRepo: categoryGroupRepo,
		categoryRepo:      categoryRepo,
	}
}

// CreateCategoryGroup creates a new category group for budget organization
func (s *CategoryGroupService) CreateCategoryGroup(ctx context.Context, name, description string, displayOrder int) (*domain.CategoryGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("category group name is required")
	}

	group := &domain.CategoryGroup{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.categoryGroupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// GetCategoryGroup retrieves a category group by ID
func (s *CategoryGroupService) GetCategoryGroup(ctx context.Context, id string) (*domain.CategoryGroup, error) {
	return s.categoryGroupRepo.GetByID(ctx, id)
}

// ListCategoryGroups retrieves all category groups
func (s *CategoryGroupService) ListCategoryGroups(ctx context.Context) ([]*domain.CategoryGroup, error) {
	return s.categoryGroupRepo.List(ctx)
}

// UpdateCategoryGroup updates an existing category group
func (s *CategoryGroupService) UpdateCategoryGroup(ctx context.Context, id, name, description string, displayOrder *int) (*domain.CategoryGroup, error) {
	group, err := s.categoryGroupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		group.Name = name
	}
	if description != "" {
		group.Description = description
	}
	if displayOrder != nil {
		group.DisplayOrder = *displayOrder
	}
	group.UpdatedAt = time.Now()

	if err := s.categoryGroupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteCategoryGroup deletes a category group
// Returns an error if the group contains any categories
func (s *CategoryGroupService) DeleteCategoryGroup(ctx context.Context, id string) error {
	// Get all categories in this group
	categories, err := s.categoryRepo.ListByGroup(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get categories in group: %w", err)
	}

	// Prevent deletion if group contains categories
	if len(categories) > 0 {
		return fmt.Errorf("cannot delete category group: it contains %d categories. Please move or delete all categories first", len(categories))
	}

	// Delete the group
	return s.categoryGroupRepo.Delete(ctx, id)
}

// AssignCategoryToGroup assigns a category to a group
func (s *CategoryGroupService) AssignCategoryToGroup(ctx context.Context, categoryID, groupID string) error {
	// Get the category
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Verify the group exists
	_, err = s.categoryGroupRepo.GetByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("category group not found: %w", err)
	}

	// Note: We no longer validate category-group type matching since categories don't have types
	// All categories are budget categories (expenses), while groups can be income or expense

	// Assign the category to the group
	category.GroupID = &groupID
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return fmt.Errorf("failed to assign category to group: %w", err)
	}

	return nil
}

// UnassignCategoryFromGroup is deprecated - categories must always belong to a group
// This method now returns an error to maintain backward compatibility with existing API
func (s *CategoryGroupService) UnassignCategoryFromGroup(ctx context.Context, categoryID string) error {
	return fmt.Errorf("categories must belong to a group. Use AssignCategoryToGroup to move this category to a different group")
}
