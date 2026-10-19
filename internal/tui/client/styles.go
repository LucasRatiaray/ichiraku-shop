package client

import "charm.land/lipgloss/v2"

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7A00"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FD787"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
)