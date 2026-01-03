package ui

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

func NewModel() (Model, error) {
	profiles, err := aws.GetProfiles()
	if err != nil {
		return Model{}, err
	}

	selected := ""
	// 1. Try to use AWS_PROFILE if set
	if p := os.Getenv("AWS_PROFILE"); p != "" {
		for _, profile := range profiles {
			if profile == p {
				selected = p
				break
			}
		}
	}

	// 2. If no AWS_PROFILE or not found, try "default"
	if selected == "" {
		for _, p := range profiles {
			if p == "default" {
				selected = "default"
				break
			}
		}
	}

	// 3. If "default" not found, use the first in the sorted list
	if selected == "" && len(profiles) > 0 {
		selected = profiles[0]
	}

	// Fallback if no profiles found at all
	if selected == "" {
		selected = "default"
	}

	styles := DefaultStyles()
	ps := NewProfileSelector(profiles, selected, styles)
	appCache := cache.New()

	// Start background cache cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			appCache.CleanExpired()
		}
	}()

	ti := textinput.New()
	ti.Placeholder = "Search services..."
	ti.Prompt = "/ "
	ti.CharLimit = 64
	ti.Width = 30

	return Model{
		profiles:         profiles,
		selectedProfile:  selected,
		profileSelector:  ps,
		styles:           styles,
		focus:            focusContent,
		view:             viewHome,
		categories:       getServiceCategories(),
		selectedCategory: 0,
		selectedService:  0,
		searchInput:      ti,
		cache:            appCache,
		cacheKeys:        cache.NewKeyBuilder(selected),
	}, nil
}

func (m Model) fetchIdentity() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		if cached, ok := m.cache.Get(m.cacheKeys.Identity()); ok {
			if id, ok := cached.(*aws.IdentityInfo); ok {
				return IdentityMsg(id)
			}
		}

		ctx := context.Background()
		stsClient, err := aws.NewSTSClient(ctx, m.selectedProfile)
		if err != nil {
			return nil
		}
		id, err := stsClient.GetCallerIdentity(ctx)
		if err != nil {
			return nil
		}

		// Try to fetch account alias
		iamClient, err := aws.NewIAMClient(ctx, m.selectedProfile)
		if err == nil {
			aliases, err := iamClient.ListAccountAliases(ctx)
			if err == nil && len(aliases) > 0 {
				id.Alias = aliases[0]
			}
		}

		// Cache the identity
		m.cache.Set(m.cacheKeys.Identity(), id, cache.TTLIdentity)

		return IdentityMsg(id)
	}
}
