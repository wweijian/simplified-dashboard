package finance

import "github.com/charmbracelet/lipgloss"

var financeTitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6")).
	Bold(true)

var financeMetaStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

var financeMoneyStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("2"))

var financeNegativeStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("1"))

var financeBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6"))
