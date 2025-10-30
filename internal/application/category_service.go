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
// Note: Categories no longer have types - only groups have types
// If a category with the same name was previously archived, returns an error suggesting restoration
func (s *CategoryService) CreateCategory(ctx context.Context, name, description, color string, groupID *string) (*domain.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	// Check if a category with this name exists (including archived)
	existing, err := s.categoryRepo.FindByNameIncludingArchived(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing category: %w", err)
	}

	// If archived category with same name exists, return error with archived category ID
	if existing != nil && existing.ArchivedAt != nil {
		return nil, fmt.Errorf("archived_category_exists:%s", existing.ID)
	}

	// If active category with same name exists, return error
	if existing != nil && existing.ArchivedAt == nil {
		return nil, fmt.Errorf("category with name '%s' already exists", name)
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
	// Allow explicit setting/unsetting of group_id
	if groupID != nil {
		category.GroupID = groupID
	}
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory archives a category (soft delete)
// Transaction history is preserved with category names intact
// Allocations are deleted via CASCADE to free up budget
func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.categoryRepo.Delete(ctx, id)
}

// RestoreCategory restores an archived category
func (s *CategoryService) RestoreCategory(ctx context.Context, id string) (*domain.Category, error) {
	if err := s.categoryRepo.RestoreCategory(ctx, id); err != nil {
		return nil, err
	}
	// Fetch and return the restored category
	return s.categoryRepo.GetByID(ctx, id)
}
