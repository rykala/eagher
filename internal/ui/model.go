package ui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cli/go-gh/v2/pkg/markdown"

	"github.com/rykala/eagher/internal/github"
)

type view int

const (
	issuesView view = iota
	prsView
)

type fetchFunc func(context.Context) ([]github.Item, error)

type tab struct {
	name   string
	fetch  fetchFunc
	list   list.Model
	loaded bool
	err    error
}

type loadedMsg struct {
	view  view
	items []github.Item
	err   error
}

type Model struct {
	repo        string
	tabs        [2]*tab
	active      view
	showPreview bool
	width       int
	height      int
	rendered    map[string]string
}

func New(client *github.Client) *Model {
	return &Model{
		repo: client.Repo(),
		tabs: [2]*tab{
			newTab("Issues", "issue", "issues", client.Issues),
			newTab("PRs", "pull request", "pull requests", client.PullRequests),
		},
		showPreview: true,
		rendered:    map[string]string{},
	}
}

func newTab(name, singular, plural string, fetch fetchFunc) *tab {
	l := list.New(nil, rowDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetStatusBarItemName(singular, plural)
	l.DisableQuitKeybindings()
	l.FilterInput.Prompt = "/ "
	return &tab{name: name, fetch: fetch, list: l}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.load(issuesView), m.load(prsView))
}

func (m *Model) load(v view) tea.Cmd {
	fetch := m.tabs[v].fetch
	return func() tea.Msg {
		items, err := fetch(context.Background())
		return loadedMsg{view: v, items: items, err: err}
	}
}

func (m *Model) current() *tab {
	return m.tabs[m.active]
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil

	case loadedMsg:
		t := m.tabs[msg.view]
		t.loaded, t.err = true, msg.err
		items := make([]list.Item, len(msg.items))
		for i, it := range msg.items {
			items[i] = item{it}
		}
		return m, t.list.SetItems(items)

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.current().list.SettingFilter() {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "s":
				m.active = 1 - m.active
				return m, nil
			case "p":
				m.showPreview = !m.showPreview
				m.resize()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.current().list, cmd = m.current().list.Update(msg)
	return m, cmd
}

func (m *Model) bodySize() (listWidth, previewWidth, height int) {
	height = max(m.height-2, 0)
	if !m.showPreview {
		return m.width, 0, height
	}
	listWidth = m.width * 45 / 100
	return listWidth, m.width - listWidth, height
}

func (m *Model) resize() {
	w, _, h := m.bodySize()
	for _, t := range m.tabs {
		t.list.SetSize(w, h)
	}
}

func (m *Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.WindowTitle = "eaGHer"
	if m.width == 0 {
		return v
	}

	listWidth, previewWidth, height := m.bodySize()
	body := lipgloss.NewStyle().Width(listWidth).Height(height).Render(m.listView())
	if m.showPreview {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, m.previewView(previewWidth, height))
	}

	v.SetContent(lipgloss.JoinVertical(lipgloss.Left, m.headerView(), body, m.footerView()))
	return v
}

func (m *Model) headerView() string {
	parts := []string{brandStyle.Render("eaGHer"), " "}
	for i, t := range m.tabs {
		label := t.name
		if t.loaded && t.err == nil {
			label = fmt.Sprintf("%s (%d)", t.name, len(t.list.Items()))
		}
		style := tabStyle
		if view(i) == m.active {
			style = activeTabStyle
		}
		parts = append(parts, style.Render(label))
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	repo := mutedStyle.Render(m.repo + " ")
	gap := max(m.width-lipgloss.Width(left)-lipgloss.Width(repo), 1)
	return left + strings.Repeat(" ", gap) + repo
}

func (m *Model) listView() string {
	t := m.current()
	switch {
	case !t.loaded:
		return mutedStyle.Render("  Loading " + strings.ToLower(t.name) + "…")
	case t.err != nil:
		return errorStyle.Render("  " + t.err.Error())
	default:
		return t.list.View()
	}
}

func (m *Model) previewView(width, height int) string {
	style := previewStyle.Width(width).Height(height).MaxHeight(height)
	it, ok := m.current().list.SelectedItem().(item)
	if !ok {
		return style.Render("")
	}

	innerWidth := width - style.GetHorizontalFrameSize()
	meta := []string{"@" + it.Author, age(it.UpdatedAt) + " ago"}
	if it.IsDraft {
		meta = append(meta, "draft")
	}
	meta = append(meta, it.Labels...)

	return style.Render(lipgloss.JoinVertical(lipgloss.Left,
		previewTitle.Width(innerWidth).Render(fmt.Sprintf("#%d %s", it.Number, it.Title)),
		mutedStyle.Width(innerWidth).Render(strings.Join(meta, " · ")),
		"",
		m.renderBody(it, innerWidth),
	))
}

func (m *Model) renderBody(it item, width int) string {
	key := fmt.Sprintf("%d:%d:%d", m.active, it.Number, width)
	if out, ok := m.rendered[key]; ok {
		return out
	}

	out := mutedStyle.Render("No description provided.")
	if strings.TrimSpace(it.Body) != "" {
		rendered, err := markdown.Render(it.Body, markdown.WithTheme("dark"), markdown.WithWrap(width), markdown.WithoutIndentation())
		if err != nil {
			rendered = it.Body
		}
		out = strings.TrimSpace(rendered)
	}
	m.rendered[key] = out
	return out
}

func (m *Model) footerView() string {
	help := "j/k move · g/G top/bottom · / search · s switch view · p preview · q quit"
	if m.current().list.SettingFilter() {
		help = "type to search · enter apply · esc cancel"
	}
	return mutedStyle.Render(" " + help)
}
