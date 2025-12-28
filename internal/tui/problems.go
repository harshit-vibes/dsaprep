package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
)

// ProblemsModel is the model for the problems browser view
type ProblemsModel struct {
	keys       KeyMap
	width      int
	height     int
	problems   []codeforces.Problem
	filtered   []codeforces.Problem
	statistics map[string]int // problemID -> solved count
	cursor     int
	offset     int
	pageSize   int

	// Filtering
	minRating  int
	maxRating  int
	searchMode bool
	search     textinput.Model
	searchTerm string
	tags       []string
}

// NewProblemsModel creates a new problems model
func NewProblemsModel(keys KeyMap) *ProblemsModel {
	ti := textinput.New()
	ti.Placeholder = "Search problems..."
	ti.CharLimit = 50
	ti.Width = 40

	return &ProblemsModel{
		keys:       keys,
		statistics: make(map[string]int),
		search:     ti,
		pageSize:   15,
		minRating:  800,
		maxRating:  1600,
	}
}

// SetSize sets the view dimensions
func (m *ProblemsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.pageSize = height - 10
	if m.pageSize < 5 {
		m.pageSize = 5
	}
}

// SetProblems sets the problems data
func (m *ProblemsModel) SetProblems(problems []codeforces.Problem) {
	m.problems = problems
	m.applyFilters()
}

// SetStatistics sets the problem statistics
func (m *ProblemsModel) SetStatistics(stats []codeforces.ProblemStatistic) {
	for _, s := range stats {
		key := fmt.Sprintf("%d%s", s.ContestID, s.Index)
		m.statistics[key] = s.SolvedCount
	}
}

// SetDifficultyRange sets the difficulty filter range
func (m *ProblemsModel) SetDifficultyRange(min, max int) {
	m.minRating = min
	m.maxRating = max
	m.applyFilters()
}

// Init implements tea.Model
func (m *ProblemsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ProblemsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle search input
	if m.searchMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEscape:
				m.searchMode = false
				m.search.Blur()
				return m, nil
			case tea.KeyEnter:
				m.searchMode = false
				m.search.Blur()
				m.searchTerm = m.search.Value()
				m.applyFilters()
				return m, nil
			}
		}

		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.pageSize {
					m.offset = m.cursor - m.pageSize + 1
				}
			}
		case key.Matches(msg, m.keys.PageUp):
			m.cursor -= m.pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.offset = m.cursor
		case key.Matches(msg, m.keys.PageDown):
			m.cursor += m.pageSize
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			if m.cursor >= m.offset+m.pageSize {
				m.offset = m.cursor - m.pageSize + 1
			}
		case key.Matches(msg, m.keys.Home):
			m.cursor = 0
			m.offset = 0
		case key.Matches(msg, m.keys.End):
			m.cursor = len(m.filtered) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.offset = m.cursor - m.pageSize + 1
			if m.offset < 0 {
				m.offset = 0
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				return m, func() tea.Msg {
					return ProblemSelectedMsg{Problem: m.filtered[m.cursor]}
				}
			}
		case key.Matches(msg, m.keys.Search):
			m.searchMode = true
			m.search.Focus()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Open):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				return m, func() tea.Msg {
					return OpenURLMsg{URL: m.filtered[m.cursor].URL()}
				}
			}
		case key.Matches(msg, m.keys.Random):
			return m, m.selectRandom()
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *ProblemsModel) View() string {
	var b strings.Builder

	// Header with search
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Table header
	b.WriteString(m.renderTableHeader())
	b.WriteString("\n")

	// Problem list
	b.WriteString(m.renderProblems())
	b.WriteString("\n")

	// Footer with pagination info
	b.WriteString(m.renderFooter())

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(1, 2).
		Render(b.String())
}

func (m *ProblemsModel) renderHeader() string {
	title := TitleStyle.Render("Problems Browser")

	var searchView string
	if m.searchMode {
		searchView = m.search.View()
	} else if m.searchTerm != "" {
		searchView = TextMuted.Render(fmt.Sprintf("Search: %s (/ to change, esc to clear)", m.searchTerm))
	} else {
		searchView = TextMuted.Render("Press / to search")
	}

	filters := TextMuted.Render(fmt.Sprintf(
		"Rating: %d-%d | Problems: %d",
		m.minRating, m.maxRating, len(m.filtered),
	))

	return lipgloss.JoinVertical(lipgloss.Left, title, searchView, filters)
}

func (m *ProblemsModel) renderTableHeader() string {
	header := fmt.Sprintf(
		"%-8s  %-6s  %-45s  %-8s  %s",
		"ID", "Rating", "Name", "Solved", "Tags",
	)
	return TableHeaderStyle.Render(header)
}

func (m *ProblemsModel) renderProblems() string {
	if len(m.filtered) == 0 {
		return TextMuted.Render("\n  No problems found matching your criteria\n")
	}

	var rows []string

	end := m.offset + m.pageSize
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.offset; i < end; i++ {
		p := m.filtered[i]
		row := m.renderProblemRow(p, i == m.cursor)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func (m *ProblemsModel) renderProblemRow(p codeforces.Problem, selected bool) string {
	id := p.ID()
	rating := fmt.Sprintf("%d", p.Rating)
	if p.Rating == 0 {
		rating = "?"
	}

	name := p.Name
	if len(name) > 43 {
		name = name[:40] + "..."
	}

	solvedCount := m.statistics[id]
	solved := fmt.Sprintf("%d", solvedCount)

	tags := ""
	if len(p.Tags) > 0 {
		if len(p.Tags) > 3 {
			tags = strings.Join(p.Tags[:3], ", ") + "..."
		} else {
			tags = strings.Join(p.Tags, ", ")
		}
	}

	row := fmt.Sprintf(
		"%-8s  %-6s  %-45s  %-8s  %s",
		id, rating, name, solved, tags,
	)

	style := TableRowStyle
	if selected {
		style = TableSelectedRowStyle
		row = "▶ " + row
	} else {
		row = "  " + row
	}

	// Color the rating
	ratingStyle := RatingStyle(p.Rating)
	coloredRating := ratingStyle.Render(rating)
	row = strings.Replace(row, rating, coloredRating, 1)

	return style.Render(row)
}

func (m *ProblemsModel) renderFooter() string {
	if len(m.filtered) == 0 {
		return ""
	}

	current := m.cursor + 1
	total := len(m.filtered)
	page := (m.offset / m.pageSize) + 1
	totalPages := (total + m.pageSize - 1) / m.pageSize

	return TextMuted.Render(fmt.Sprintf(
		"Problem %d of %d  •  Page %d of %d  •  ↑↓ navigate  •  Enter to select  •  o to open in browser",
		current, total, page, totalPages,
	))
}

func (m *ProblemsModel) applyFilters() {
	m.filtered = nil
	m.cursor = 0
	m.offset = 0

	for _, p := range m.problems {
		// Rating filter
		if p.Rating > 0 && (p.Rating < m.minRating || p.Rating > m.maxRating) {
			continue
		}

		// Search filter
		if m.searchTerm != "" {
			searchLower := strings.ToLower(m.searchTerm)
			nameLower := strings.ToLower(p.Name)
			idLower := strings.ToLower(p.ID())

			if !strings.Contains(nameLower, searchLower) && !strings.Contains(idLower, searchLower) {
				// Also check tags
				found := false
				for _, tag := range p.Tags {
					if strings.Contains(strings.ToLower(tag), searchLower) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		// Tag filter
		if len(m.tags) > 0 {
			hasAllTags := true
			for _, requiredTag := range m.tags {
				found := false
				for _, problemTag := range p.Tags {
					if strings.EqualFold(problemTag, requiredTag) {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		m.filtered = append(m.filtered, p)
	}
}

func (m *ProblemsModel) selectRandom() tea.Cmd {
	return func() tea.Msg {
		if len(m.filtered) == 0 {
			return StatusMsg{Message: "No problems available", IsError: true}
		}

		// Simple random selection (use middle element for now)
		idx := len(m.filtered) / 2
		return ProblemSelectedMsg{Problem: m.filtered[idx]}
	}
}
