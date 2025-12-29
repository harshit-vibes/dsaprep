package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/tui/styles"
)

// SettingsModel is the settings view model
type SettingsModel struct {
	width  int
	height int

	// State
	selectedIdx int
	items       []settingItem
}

type settingItem struct {
	key         string
	label       string
	value       string
	description string
	editable    bool
}

// NewSettingsModel creates a new settings model
func NewSettingsModel() SettingsModel {
	return SettingsModel{
		items: []settingItem{
			{
				key:         "cf_handle",
				label:       "CF Handle",
				description: "Your Codeforces username",
				editable:    true,
			},
			{
				key:         "difficulty.min",
				label:       "Min Difficulty",
				description: "Minimum problem rating to show",
				editable:    true,
			},
			{
				key:         "difficulty.max",
				label:       "Max Difficulty",
				description: "Maximum problem rating to show",
				editable:    true,
			},
			{
				key:         "daily_goal",
				label:       "Daily Goal",
				description: "Number of problems to solve per day",
				editable:    true,
			},
			{
				key:         "workspace_path",
				label:       "Workspace Path",
				description: "Path to your cf workspace",
				editable:    true,
			},
		},
	}
}

// SetSize sets the view dimensions
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(m.items)-1 {
				m.selectedIdx++
			}
		}
	}

	return m, nil
}

// View renders the settings view
func (m SettingsModel) View() string {
	// Load current config values
	cfg := config.Get()
	creds, _ := config.LoadCredentials()

	// Update item values
	if cfg != nil {
		for i := range m.items {
			switch m.items[i].key {
			case "cf_handle":
				if creds != nil && creds.CFHandle != "" {
					m.items[i].value = creds.CFHandle
				} else {
					m.items[i].value = cfg.CFHandle
				}
			case "difficulty.min":
				m.items[i].value = fmt.Sprintf("%d", cfg.Difficulty.Min)
			case "difficulty.max":
				m.items[i].value = fmt.Sprintf("%d", cfg.Difficulty.Max)
			case "daily_goal":
				m.items[i].value = fmt.Sprintf("%d", cfg.DailyGoal)
			case "workspace_path":
				if cfg.WorkspacePath != "" {
					m.items[i].value = cfg.WorkspacePath
				} else {
					m.items[i].value = "(not set)"
				}
			}
		}
	}

	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("⚙️  Settings"))
	b.WriteString("\n")
	b.WriteString(styles.SubtitleStyle.Render("  Configuration is stored in ~/.cf/config.yaml"))
	b.WriteString("\n\n")

	// Settings list
	for i, item := range m.items {
		b.WriteString(m.renderSettingItem(item, i == m.selectedIdx))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Credentials section
	b.WriteString(styles.TitleStyle.Render("🔑 Credentials"))
	b.WriteString("\n")
	b.WriteString(styles.SubtitleStyle.Render("  Stored in ~/.cf.env"))
	b.WriteString("\n\n")

	if creds != nil {
		b.WriteString(m.renderCredentialItem("API Key", maskValue(creds.APIKey)))
		b.WriteString("\n")
		b.WriteString(m.renderCredentialItem("API Secret", maskValue(creds.APISecret)))
		b.WriteString("\n")
		b.WriteString(m.renderCredentialItem("Session", func() string {
			if creds.HasSessionCookies() {
				return styles.SuccessStyle.Render("configured")
			}
			return styles.WarningStyle.Render("not configured")
		}()))
		b.WriteString("\n")
		b.WriteString(m.renderCredentialItem("CF Clearance", creds.GetCFClearanceStatus()))
	}

	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("  ↑/↓ navigate • Use 'cf config set <key> <value>' to modify settings"))

	return b.String()
}

func (m SettingsModel) renderSettingItem(item settingItem, selected bool) string {
	labelStyle := lipgloss.NewStyle().
		Width(18).
		Foreground(styles.ColorTextSecondary)

	valueStyle := lipgloss.NewStyle().
		Width(30).
		Foreground(styles.ColorTextPrimary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted)

	if item.value == "" || item.value == "(not set)" {
		valueStyle = valueStyle.Foreground(styles.ColorMuted)
	}

	row := fmt.Sprintf("  %s %s %s",
		labelStyle.Render(item.label+":"),
		valueStyle.Render(item.value),
		descStyle.Render(item.description),
	)

	if selected {
		return styles.SelectedItemStyle.Render(row)
	}
	return row
}

func (m SettingsModel) renderCredentialItem(label, value string) string {
	labelStyle := lipgloss.NewStyle().
		Width(18).
		Foreground(styles.ColorTextSecondary)

	return fmt.Sprintf("  %s %s",
		labelStyle.Render(label+":"),
		value,
	)
}

func maskValue(s string) string {
	if s == "" {
		return styles.WarningStyle.Render("not set")
	}
	if len(s) <= 8 {
		return styles.SuccessStyle.Render("****")
	}
	return styles.SuccessStyle.Render(s[:4] + "..." + s[len(s)-4:])
}
