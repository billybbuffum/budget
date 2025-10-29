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
func (s *CategoryService) CreateCategory(ctx context.Context, name string, categoryType domain.CategoryType, description, color string) (*domain.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	if categoryType != domain.CategoryTypeIncome && categoryType != domain.CategoryTypeExpense {
		return nil, fmt.Errorf("invalid category type: must be 'income' or 'expense'")
	}

	category := &domain.Category{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        categoryType,
		Description: description,
		Color:       color,
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

// ListCategoriesByType retrieves categories by type
func (s *CategoryService) ListCategoriesByType(ctx context.Context, categoryType domain.CategoryType) ([]*domain.Category, error) {
	return s.categoryRepo.ListByType(ctx, categoryType)
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(ctx context.Context, id, name string, categoryType domain.CategoryType, description, color string) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		category.Name = name
	}
	if categoryType != "" {
		if categoryType != domain.CategoryTypeIncome && categoryType != domain.CategoryTypeExpense {
			return nil, fmt.Errorf("invalid category type: must be 'income' or 'expense'")
		}
		category.Type = categoryType
	}
	if description != "" {
		category.Description = description
	}
	if color != "" {
		category.Color = color
	}
	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.categoryRepo.Delete(ctx, id)
}
