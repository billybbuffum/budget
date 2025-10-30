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
			Name:         "Housing",
			Description:  "Home and living expenses",
			DisplayOrder: 1,
			Categories: []DefaultCategory{
				{Name: "Rent/Mortgage", Description: "Monthly rent or mortgage payment", Color: "#3B82F6"},
				{Name: "Electric", Description: "Electricity bill", Color: "#EAB308"},
				{Name: "Water", Description: "Water and sewer", Color: "#06B6D4"},
				{Name: "Gas/Heating", Description: "Natural gas or heating oil", Color: "#F97316"},
				{Name: "Internet", Description: "Internet service", Color: "#8B5CF6"},
				{Name: "Home Maintenance", Description: "Repairs and maintenance", Color: "#6366F1"},
			},
		},
		{
			Name:         "Transportation",
			Description:  "Vehicle and travel expenses",
			DisplayOrder: 2,
			Categories: []DefaultCategory{
				{Name: "Car Payment", Description: "Auto loan payment", Color: "#EF4444"},
				{Name: "Gas/Fuel", Description: "Gasoline or fuel", Color: "#F59E0B"},
				{Name: "Car Insurance", Description: "Auto insurance premium", Color: "#10B981"},
				{Name: "Public Transit", Description: "Bus, train, subway passes", Color: "#3B82F6"},
				{Name: "Car Maintenance", Description: "Oil changes, repairs, tires", Color: "#6B7280"},
			},
		},
		{
			Name:         "Food",
			Description:  "Food and dining expenses",
			DisplayOrder: 3,
			Categories: []DefaultCategory{
				{Name: "Groceries", Description: "Supermarket and grocery stores", Color: "#10B981"},
				{Name: "Restaurants", Description: "Dining out", Color: "#F59E0B"},
				{Name: "Coffee Shops", Description: "Coffee and cafes", Color: "#92400E"},
			},
		},
		{
			Name:         "Personal",
			Description:  "Personal care and lifestyle",
			DisplayOrder: 4,
			Categories: []DefaultCategory{
				{Name: "Clothing", Description: "Clothes and shoes", Color: "#EC4899"},
				{Name: "Personal Care", Description: "Haircuts, toiletries, etc.", Color: "#8B5CF6"},
				{Name: "Entertainment", Description: "Movies, concerts, events", Color: "#F59E0B"},
				{Name: "Hobbies", Description: "Hobby supplies and activities", Color: "#06B6D4"},
				{Name: "Gifts", Description: "Gifts for others", Color: "#EF4444"},
			},
		},
		{
			Name:         "Bills & Utilities",
			Description:  "Recurring bills and services",
			DisplayOrder: 5,
			Categories: []DefaultCategory{
				{Name: "Phone", Description: "Cell phone bill", Color: "#3B82F6"},
				{Name: "Streaming Services", Description: "Netflix, Spotify, etc.", Color: "#EC4899"},
				{Name: "Subscriptions", Description: "Other recurring subscriptions", Color: "#8B5CF6"},
			},
		},
		{
			Name:         "Savings & Debt",
			Description:  "Savings goals and debt payments",
			DisplayOrder: 6,
			Categories: []DefaultCategory{
				{Name: "Emergency Fund", Description: "Emergency savings", Color: "#10B981"},
				{Name: "Debt Payments", Description: "Credit cards, loans", Color: "#EF4444"},
				{Name: "Savings Goals", Description: "Vacation, car, house, etc.", Color: "#3B82F6"},
			},
		},
		{
			Name:         "General",
			Description:  "Other miscellaneous expenses",
			DisplayOrder: 7,
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
