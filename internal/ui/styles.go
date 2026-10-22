package ui

import "charm.land/lipgloss/v2"

var (
	accent = lipgloss.Color("#bd93f9")
	muted  = lipgloss.Color("#6272a4")
	fg     = lipgloss.Color("#f8f8f2")
	green  = lipgloss.Color("#50fa7b")
	red    = lipgloss.Color("#ff5555")

	brandStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#282a36")).Background(accent).Padding(0, 1)
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(accent).Underline(true).Padding(0, 1)
	tabStyle         = lipgloss.NewStyle().Foreground(muted).Padding(0, 1)
	cursorStyle      = lipgloss.NewStyle().Foreground(accent)
	rowStyle         = lipgloss.NewStyle().Foreground(fg)
	selectedRowStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	numberStyle      = lipgloss.NewStyle().Foreground(green)
	authorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd"))
	mutedStyle       = lipgloss.NewStyle().Foreground(muted)
	errorStyle       = lipgloss.NewStyle().Foreground(red)
	previewStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(muted).PaddingLeft(1)
	previewTitle     = lipgloss.NewStyle().Bold(true).Foreground(fg)
)
