package habits

import "github.com/charmbracelet/lipgloss"

var (
	habitMetaStyle       = lipgloss.NewStyle().Faint(true)
	habitHeaderStyle     = lipgloss.NewStyle().Bold(true)
	habitActiveStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	habitCompleteStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	habitIncompleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	habitSkippedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	habitMutedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	habitSelectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
)
