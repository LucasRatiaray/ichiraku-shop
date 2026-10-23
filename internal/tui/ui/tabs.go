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

func tabContent(t Tab) string {
	if t.Key == "" {
		return NameStyle.Render(t.Label)
	}
	label := DimStyle.Render(t.Label)
	if t.Active {
		label = TitleStyle.Render(t.Label)
	}
	s := NameStyle.Render(t.Key) + " " + label
	if t.Detail != "" {
		s += " " + NameStyle.Render(t.Detail)
	}
	return s
}

func Tabs(width int, tabs []Tab) string {
	last := len(tabs) - 1
	contents := make([]string, len(tabs))
	natural := 1
	for i, t := range tabs {
		contents[i] = tabContent(t)
		natural += lipgloss.Width(contents[i]) + 3
	}
	extra := max(width-natural, 0)

	cells := make([]string, len(tabs))
	for i, c := range contents {
		border := lipgloss.NormalBorder()
		if i > 0 {
			border.TopLeft = "┬"
			border.BottomLeft = "┴"
		}
		w := lipgloss.Width(c) + 3 + extra/len(tabs)
		if i == last {
			w += 1 + extra%len(tabs)
		}
		cells[i] = lipgloss.NewStyle().
			Border(border).
			BorderRight(i == last).
			BorderForeground(Line).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(w).
			Render(c)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func HelpBar(width int, hints []Hint) string {
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = NameStyle.Render(h.Key) + " " + DimStyle.Render(h.Label)
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Join(parts, "   "))
}

func Shop(width, height int, tabs []Tab, body string, hints ...Hint) tea.View {
	inner := ContentWidth(width)
	block := lipgloss.NewStyle().Width(pageWidth(width)).Padding(0, pagePadding)
	rule := lipgloss.NewStyle().Foreground(Line).Render(strings.Repeat("─", pageWidth(width)))

	head := block.Render(Tabs(inner, tabs))
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
