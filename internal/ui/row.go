package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/rykala/eagher/internal/github"
)

const (
	numberWidth = 7
	authorWidth = 16
	ageWidth    = 4
)

type item struct {
	github.Item
	pinned bool
}

func (i item) FilterValue() string {
	return fmt.Sprintf("#%d %s @%s", i.Number, i.Title, i.Author)
}

type rowDelegate struct{}

func (rowDelegate) Height() int                         { return 1 }
func (rowDelegate) Spacing() int                        { return 0 }
func (rowDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (rowDelegate) Render(w io.Writer, m list.Model, index int, li list.Item) {
	it := li.(item)
	selected := index == m.Index()

	cursor := "  "
	if selected {
		cursor = cursorStyle.Render("▌ ")
	}
	pin := "  "
	if it.pinned {
		pin = pinStyle.Render("★ ")
	}

	titleWidth := max(m.Width()-lipgloss.Width(cursor)-lipgloss.Width(pin)-numberWidth-authorWidth-ageWidth-2, 0)
	title := ansi.Truncate(it.Title, titleWidth, "…")
	title += strings.Repeat(" ", titleWidth-lipgloss.Width(title))

	titleStyle := rowStyle
	switch {
	case selected:
		titleStyle = selectedRowStyle
	case it.IsDraft:
		titleStyle = mutedStyle
	}

	fmt.Fprint(w, cursor+pin+
		numberStyle.Render(fmt.Sprintf("%-*s", numberWidth, fmt.Sprintf("#%d", it.Number)))+
		titleStyle.Render(title)+" "+
		authorStyle.Render(fmt.Sprintf("%-*s", authorWidth, ansi.Truncate("@"+it.Author, authorWidth, "…")))+" "+
		mutedStyle.Render(fmt.Sprintf("%*s", ageWidth, age(it.UpdatedAt))))
}

func age(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/24/7))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo", int(d.Hours()/24/30))
	default:
		return fmt.Sprintf("%dy", int(d.Hours()/24/365))
	}
}
