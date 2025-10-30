package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// CategoryService handles category-related business logic
type CategoryService struct {
	categoryRepo domain.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo domain.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// CreateCategory creates a new category
// Note: groupID is required - all categories must belong to a group
func (s *CategoryService) CreateCategory(ctx context.Context, name, description, color string, groupID *string) (*domain.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	if groupID == nil || *groupID == "" {
		return nil, fmt.Errorf("group_id is required - all categories must belong to a group")
	}

	category := &domain.Category{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Color:       color,
		GroupID:     groupID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

// ListCategories retrieves all categories
func (s *CategoryService) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	return s.categoryRepo.List(ctx)
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(ctx context.Context, id, name, description, color string, groupID *string) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		category.Name = name
	}
	if description != "" {
		category.Description = description
	}
	if color != "" {
		category.Color = color
	}
	// Update group_id if provided, but ensure it's not nil
	if groupID != nil {
		if *groupID == "" {
			return nil, fmt.Errorf("group_id cannot be empty - all categories must belong to a group")
		}
		category.GroupID = groupID
	}
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
// NOTE: Consider implementing soft delete in the future to preserve history
// For now, foreign key constraints prevent deletion if transactions/allocations exist
func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.categoryRepo.Delete(ctx, id)
}
