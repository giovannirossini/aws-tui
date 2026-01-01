package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

func (m Model) getCategoryColumn(categoryIdx int) int {
	switch categoryIdx {
	case 0, 1:
		return 0
	case 2, 3, 5:
		return 1
	case 4, 6:
		return 2
	default:
		return 0
	}
}

func (m Model) getCategoriesInColumn(col int) []int {
	switch col {
	case 0:
		return []int{0, 1}
	case 1:
		return []int{2, 3, 5}
	case 2:
		return []int{4, 6}
	default:
		return []int{}
	}
}

func (m *Model) moveToColumn(newCol int) {
	oldCol := m.getCategoryColumn(m.selectedCategory)
	categoriesInOldCol := m.getCategoriesInColumn(oldCol)

	oldRank := 0
	for i, catIdx := range categoriesInOldCol {
		if catIdx == m.selectedCategory {
			oldRank = i
			break
		}
	}

	categoriesInNewCol := m.getCategoriesInColumn(newCol)
	if len(categoriesInNewCol) > 0 {
		newRank := oldRank
		if newRank >= len(categoriesInNewCol) {
			newRank = len(categoriesInNewCol) - 1
		}
		m.selectedCategory = categoriesInNewCol[newRank]
		if m.selectedService >= len(m.categories[m.selectedCategory].Services) {
			m.selectedService = len(m.categories[m.selectedCategory].Services) - 1
		}
	}
}

func (m Model) isInputFocused() bool {
	if m.searching {
		return true
	}
	if m.view == viewS3 && m.s3Model.state == S3StateInput {
		return true
	}
	if m.view == viewIAM && m.iamModel.state == IAMStateInput {
		return true
	}
	return false
}

func (m *Model) updateFilter() {
	query := m.searchInput.Value()
	allServices := []string{}
	for _, cat := range m.categories {
		allServices = append(allServices, cat.Services...)
	}

	if query == "" {
		m.filteredServices = allServices
	} else {
		matches := fuzzy.Find(query, allServices)
		m.filteredServices = make([]string, len(matches))
		for i, match := range matches {
			m.filteredServices[i] = match.Str
		}
	}

	// Limit to max 10 items
	maxItems := 10
	if len(m.filteredServices) > maxItems {
		m.filteredServices = m.filteredServices[:maxItems]
	}

	if m.selectedFiltered >= len(m.filteredServices) {
		m.selectedFiltered = len(m.filteredServices) - 1
		if m.selectedFiltered < 0 {
			m.selectedFiltered = 0
		}
	}
	if m.selectedFiltered < 0 && len(m.filteredServices) > 0 {
		m.selectedFiltered = 0
	}
}

func (m *Model) handleServiceSelection(selectedService string) (tea.Model, tea.Cmd) {
	handlers := getServiceHandlers()
	if handler, ok := handlers[selectedService]; ok {
		return handler(m)
	}
	return *m, nil
}
