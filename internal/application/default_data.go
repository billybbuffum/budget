package application

import (
	"context"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// DefaultCategoryGroup represents a default category group to be created
type DefaultCategoryGroup struct {
	Name         string
	Description  string
	DisplayOrder int
	Categories   []DefaultCategory
}

// DefaultCategory represents a default category to be created
type DefaultCategory struct {
	Name        string
	Description string
	Color       string
}

// GetDefaultCategoryGroups returns the default category groups and categories
func GetDefaultCategoryGroups() []DefaultCategoryGroup {
	return []DefaultCategoryGroup{
		{
			Name:         "Housing & Bills",
			Description:  "Home, utilities, and recurring bills",
			DisplayOrder: 1,
			Categories: []DefaultCategory{
				{Name: "Rent/Mortgage", Description: "Monthly rent or mortgage payment", Color: "#3B82F6"},
				{Name: "Utilities", Description: "Electric, water, gas, internet", Color: "#EAB308"},
				{Name: "Phone", Description: "Cell phone bill", Color: "#8B5CF6"},
			},
		},
		{
			Name:         "Transportation",
			Description:  "Vehicle and travel expenses",
			DisplayOrder: 2,
			Categories: []DefaultCategory{
				{Name: "Gas/Fuel", Description: "Gasoline, fuel, public transit", Color: "#F59E0B"},
				{Name: "Car Payment & Insurance", Description: "Auto loan and insurance", Color: "#EF4444"},
			},
		},
		{
			Name:         "Food & Dining",
			Description:  "Groceries and eating out",
			DisplayOrder: 3,
			Categories: []DefaultCategory{
				{Name: "Groceries", Description: "Supermarket and grocery stores", Color: "#10B981"},
				{Name: "Restaurants", Description: "Dining out, coffee, etc.", Color: "#F59E0B"},
			},
		},
		{
			Name:         "Personal & Lifestyle",
			Description:  "Personal care, entertainment, and subscriptions",
			DisplayOrder: 4,
			Categories: []DefaultCategory{
				{Name: "Shopping", Description: "Clothing, personal items", Color: "#EC4899"},
				{Name: "Entertainment", Description: "Movies, hobbies, streaming services", Color: "#06B6D4"},
			},
		},
		{
			Name:         "Other",
			Description:  "Everything else",
			DisplayOrder: 5,
			Categories: []DefaultCategory{
				{Name: "Miscellaneous", Description: "Uncategorized expenses", Color: "#6B7280"},
			},
		},
	}
}

// BootstrapService handles initialization of default data
type BootstrapService struct {
	categoryGroupRepo domain.CategoryGroupRepository
	categoryRepo      domain.CategoryRepository
}

// NewBootstrapService creates a new bootstrap service
func NewBootstrapService(
	categoryGroupRepo domain.CategoryGroupRepository,
	categoryRepo domain.CategoryRepository,
) *BootstrapService {
	return &BootstrapService{
		categoryGroupRepo: categoryGroupRepo,
		categoryRepo:      categoryRepo,
	}
}

// InitializeDefaultData creates default category groups and categories if they don't exist
func (s *BootstrapService) InitializeDefaultData(ctx context.Context) error {
	// Check if any category groups already exist
	existingGroups, err := s.categoryGroupRepo.List(ctx)
	if err != nil {
		return err
	}

	// If groups already exist, skip initialization
	if len(existingGroups) > 0 {
		return nil
	}

	// Create default groups and categories
	defaultGroups := GetDefaultCategoryGroups()
	now := time.Now()

	for _, defaultGroup := range defaultGroups {
		// Create the group
		groupID := uuid.New().String()
		group := &domain.CategoryGroup{
			ID:           groupID,
			Name:         defaultGroup.Name,
			Description:  defaultGroup.Description,
			DisplayOrder: defaultGroup.DisplayOrder,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := s.categoryGroupRepo.Create(ctx, group); err != nil {
			return err
		}

		// Create categories for this group
		for _, defaultCat := range defaultGroup.Categories {
			category := &domain.Category{
				ID:          uuid.New().String(),
				Name:        defaultCat.Name,
				Description: defaultCat.Description,
				Color:       defaultCat.Color,
				GroupID:     &groupID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := s.categoryRepo.Create(ctx, category); err != nil {
				return err
			}
		}
	}

	return nil
}
