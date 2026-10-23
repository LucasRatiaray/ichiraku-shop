package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Tab struct {
	Key    string
	Label  string
	Detail string
	Active bool
}

type Hint struct {
	Key   string
	Label string
}

func tabContent(t Tab, compact bool) string {
	key := NameStyle
	if compact && t.Active {
		key = TitleStyle
	}
	if compact {
		return strings.TrimSpace(key.Render(t.Key) + " " + NameStyle.Render(t.Detail))
	}
	label := DimStyle.Render(t.Label)
	if t.Active {
		label = TitleStyle.Render(t.Label)
	}
	s := key.Render(t.Key) + " " + label
	if t.Detail != "" {
		s += " " + NameStyle.Render(t.Detail)
	}
	return s
}

func tabsBar(tabs []Tab, compact bool) string {
	last := len(tabs) - 1
	cells := make([]string, len(tabs))
	for i, t := range tabs {
		content := tabContent(t, compact)
		border := lipgloss.NormalBorder()
		if i > 0 {
			border.TopLeft = "┬"
			border.BottomLeft = "┴"
		}
		cells[i] = lipgloss.NewStyle().
			Border(border).
			BorderRight(i == last).
			BorderForeground(Line).
			Padding(0, 1).
			Render(content)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func Tabs(width int, tabs []Tab) string {
	bar := tabsBar(tabs, false)
	if lipgloss.Width(bar) > width {
		bar = tabsBar(tabs, true)
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(bar)
}

func HelpBar(width int, hints []Hint) string {
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = NameStyle.Render(h.Key) + " " + DimStyle.Render(h.Label)
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Join(parts, "   "))
}

func Shop(width, height int, title string, tabs []Tab, body string, hints ...Hint) tea.View {
	inner := ContentWidth(width)
	block := lipgloss.NewStyle().Width(pageWidth(width)).Padding(0, pagePadding)
	rule := lipgloss.NewStyle().Foreground(Line).Render(strings.Repeat("─", pageWidth(width)))

	heading := lipgloss.PlaceHorizontal(inner, lipgloss.Center, TitleStyle.Render("🍜 "+title))
	bar := lipgloss.PlaceHorizontal(inner, lipgloss.Center, Tabs(inner, tabs))
	head := block.Render(heading + "\n" + bar)
	shipping := lipgloss.PlaceHorizontal(inner, lipgloss.Center, DimStyle.Render("Livraison gratuite"))
	foot := block.Render(shipping) + "\n" + rule + "\n" + block.Render(HelpBar(inner, hints))
	main := block.PaddingTop(1).Height(max(height-lipgloss.Height(head)-lipgloss.Height(foot), 0)).Render(body)

	return newView(width, lipgloss.JoinVertical(lipgloss.Left, head, main, foot))
}

func newView(width int, page string) tea.View {
	v := tea.NewView(lipgloss.PlaceHorizontal(width, lipgloss.Center, page))
	v.BackgroundColor = Bg
	v.ForegroundColor = Fg
	return v
}
