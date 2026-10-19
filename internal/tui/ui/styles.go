package ui

import "charm.land/lipgloss/v2"

var (
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7A00"))
	ErrStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F"))
	InfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FD787"))
	HelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	NameStyle  = lipgloss.NewStyle().Bold(true)
)
