package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const (
	MaxWidth    = 80
	pagePadding = 2
	cardChrome  = 4
)

var (
	Accent = lipgloss.Color("#F5A524")
	Bg     = lipgloss.Color("#0D0D0D")
	Fg     = lipgloss.Color("#E8E8E8")
	Muted  = lipgloss.Color("#808080")
	Dim    = lipgloss.Color("#6B6B6B")
	Line   = lipgloss.Color("#3A3A3A")
	Green  = lipgloss.Color("#5FD787")
	Red    = lipgloss.Color("#E5484D")
	Blue   = lipgloss.Color("#5FAFFF")
	Yellow = lipgloss.Color("#FFD75F")
)

var (
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	ErrStyle   = lipgloss.NewStyle().Foreground(Red)
	InfoStyle  = lipgloss.NewStyle().Foreground(Green)
	HelpStyle  = lipgloss.NewStyle().Foreground(Muted)
	DimStyle   = lipgloss.NewStyle().Foreground(Dim)
	NameStyle  = lipgloss.NewStyle().Bold(true).Foreground(Fg)
	PriceStyle = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	CardStyle  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Line).
			Padding(0, 1)
)

func pageWidth(termWidth int) int {
	if termWidth <= 0 || termWidth > MaxWidth {
		return MaxWidth
	}
	return termWidth
}

func ContentWidth(termWidth int) int {
	return pageWidth(termWidth) - 2*pagePadding
}

func Badge(label string, bg color.Color) string {
	return lipgloss.NewStyle().Bold(true).Foreground(Bg).Background(bg).Padding(0, 1).Render(label)
}

func Row(width int, left, right string) string {
	return left + strings.Repeat(" ", max(width-lipgloss.Width(left)-lipgloss.Width(right), 1)) + right
}

func CardInner(width int) int {
	return width - cardChrome
}

func Card(width int, left, right, body string) string {
	content := Row(CardInner(width), left, right)
	if body != "" {
		content += "\n" + body
	}
	return CardStyle.Width(width).Render(content)
}

func Page(width, height int, brand, screen, body, help string) tea.View {
	w := pageWidth(width)
	block := lipgloss.NewStyle().Width(w).Padding(0, pagePadding)
	rule := lipgloss.NewStyle().Foreground(Line).Render(strings.Repeat("─", w))

	left := TitleStyle.Render("🍜 " + brand)
	right := HelpStyle.Render(screen)

	head := block.Render(Row(ContentWidth(width), left, right)) + "\n" + rule
	foot := rule + "\n" + block.Render(HelpStyle.Render(help))
	main := block.PaddingTop(1).Height(max(height-lipgloss.Height(head)-lipgloss.Height(foot), 0)).Render(body)

	return newView(width, lipgloss.JoinVertical(lipgloss.Left, head, main, foot))
}

func NewForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithTheme(huh.ThemeFunc(HuhTheme)).WithShowHelp(false)
}

func HuhTheme(isDark bool) *huh.Styles {
	t := huh.ThemeBase(isDark)

	t.Focused.Base = t.Focused.Base.BorderForeground(Accent)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(Accent).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(Accent).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(Muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(Red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(Red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(Accent)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(Accent)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(Accent)
	t.Focused.Option = t.Focused.Option.Foreground(Fg)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(Accent)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(Accent)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(Green).SetString("✓ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(Muted).SetString("• ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(Fg)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(Bg).Background(Accent)
	t.Focused.Next = t.Focused.FocusedButton
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(Fg).Background(Line)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(Accent)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(Muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(Accent)
	t.Focused.TextInput.Text = t.Focused.TextInput.Text.Foreground(Fg)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.Title = t.Focused.Title.Foreground(Muted)
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	return t
}
