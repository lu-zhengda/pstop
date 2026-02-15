package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Title and header styles.
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	// Tab styles.
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Process row styles.
	highCPUStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red for CPU > 50%
	medCPUStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow for CPU 20-50%
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("8"))  // Highlighted row

	// Status and info styles.
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	filterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	labelStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Group header in dev view.
	groupStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
)
