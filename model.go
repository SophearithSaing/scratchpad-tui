package main

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	colorInk     = lipgloss.Color("#E9E5D6")
	colorMuted   = lipgloss.Color("#77756F")
	colorFaint   = lipgloss.Color("#34332F")
	colorPaper   = lipgloss.Color("#171816")
	colorRaised  = lipgloss.Color("#232420")
	colorAccent  = lipgloss.Color("#6EA8FE")
	colorBlue    = lipgloss.Color("#9CCAFF")
	colorDanger  = lipgloss.Color("#FF7A72")
	styleLogo    = lipgloss.NewStyle().Bold(true).Foreground(colorPaper).Background(colorAccent).Padding(0, 1)
	styleTab     = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1)
	styleTabOn   = lipgloss.NewStyle().Bold(true).Foreground(colorInk).Background(colorRaised).Padding(0, 1)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleAccent  = lipgloss.NewStyle().Foreground(colorAccent)
	styleDanger  = lipgloss.NewStyle().Foreground(colorDanger)
	styleOverlay = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorFaint).Background(colorRaised).Padding(1, 2)
	styleEditor  = lipgloss.NewStyle().Margin(0, 1).PaddingTop(1).PaddingLeft(1).PaddingRight(1)
)

const autosaveDelay = 1 * time.Second

type tab struct {
	id            int
	fallbackTitle string
	editor        textarea.Model
}

type inputMode int

const (
	editMode inputMode = iota
	viewMode
	exportMode
)

// Bubble Tea application state
type appModel struct {
	store       sessionStore
	tabs        []tab
	active      int
	nextID      int
	width       int
	height      int
	mode        inputMode
	exportMode  inputMode
	preview     viewport.Model
	help        viewport.Model
	pathInput   textinput.Model
	showHelp    bool
	status      string
	statusError bool
	revision    uint64
	savedRev    uint64
}

type autosaveMsg uint64

// type autosaveMsg struct {
// 	revision uint64
// }

func newAppModel(store sessionStore, session savedSession) appModel {
	m := appModel{store: store, nextID: 1, preview: viewport.New(), help: viewport.New()}
	m.preview.FillHeight = true
	m.help.SoftWrap = true
	m.help.FillHeight = true

	for _, saved := range session.Tabs {
		m.tabs = append(m.tabs, newTab(saved.ID, saved.Title, saved.Content))
		if saved.ID >= m.nextID {
			m.nextID = saved.ID + 1
		}
	}
	if session.NextID > m.nextID {
		m.nextID = session.NextID
	}
	if len(m.tabs) == 0 {
		m.addTab()
	}
	if session.Active >= 0 && session.Active < len(m.tabs) {
		m.active = session.Active
	}

	m.pathInput = textinput.New()
	m.pathInput.Prompt = " Save to "
	m.pathInput.Placeholder = "notes.md"
	m.pathInput.CharLimit = 4096
	pathStyles := m.pathInput.Styles()
	pathStyles.Cursor.Color = colorAccent
	m.pathInput.SetStyles(pathStyles)
	m.focusActive()

	return m
}

func (m appModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+n":
			m.addTab()
			return m, nil
		}
		return m.updateContent(msg)
	}

	return m.updateContent(msg)
}

func (m appModel) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Scratchpad"
	return view
}

func (m appModel) render() string {
	if m.width == 0 || m.height == 0 {
		return "Loading scratchpad..."
	}
	header := m.renderHeader()
	body := styleEditor.Render(m.tabs[m.active].editor.View())
	page := lipgloss.JoinVertical(lipgloss.Left, header, body)
	return page
}

func (m appModel) renderHeader() string {
	logo := styleLogo.Render("SCRATCHPAD")
	available := max(0, m.width-lipgloss.Width(logo)-1)
	renderTab := func(i int) string {
		tab := m.tabs[i]
		label := fmt.Sprintf("%d %s", i+1, tab.title())
		if i == m.active {
			return styleTabOn.Render(label)
		}
		return styleTab.Render(label)
	}

	start, end := m.active, m.active
	used := lipgloss.Width(renderTab(m.active))
	for start > 0 {
		candidate := renderTab(start - 1)
		if used+1+lipgloss.Width(candidate) > available {
			break
		}
		start--
		used += 1 + lipgloss.Width(candidate)
	}
	for end+1 < len(m.tabs) {
		candidate := renderTab(end + 1)
		if used+1+lipgloss.Width(candidate) > available {
			break
		}
		end++
		used += 1 + lipgloss.Width(candidate)
	}

	tabs := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		tabs = append(tabs, renderTab(i))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, logo, " ", strings.Join(tabs, " "))
}

func (m *appModel) addTab() {
	m.tabs = append(m.tabs, newTab(m.nextID, "", ""))
	m.active = len(m.tabs) - 1
	m.nextID++
}

func (m *appModel) focusActive() tea.Cmd {
	var cmd tea.Cmd
	for i := range m.tabs {
		if i == m.active && m.mode == editMode {
			cmd = m.tabs[i].editor.Focus()
		} else {
			m.tabs[i].editor.Blur()
		}
	}
	return cmd
}

func (m *appModel) blurActive() {
	if len(m.tabs) > 0 {
		m.tabs[m.active].editor.Blur()
	}
}

func (m appModel) updateContent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	before := m.tabs[m.active].editor.Value()
	m.tabs[m.active].editor, cmd = m.tabs[m.active].editor.Update(msg)
	if m.tabs[m.active].editor.Value() != before {
		updated, saveCmd := m.changed()
		m = updated.(appModel)
		return m, tea.Batch(cmd, saveCmd)
	}
	return m, cmd
}

func (m appModel) changed() (tea.Model, tea.Cmd) {
	m.revision++
	revision := m.revision
	return m, tea.Tick(autosaveDelay, func(time.Time) tea.Msg { return autosaveMsg(revision) })
}

func (t tab) title() string {
	for line := range strings.SplitSeq(t.editor.Value(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return truncate(line, 22)
		}
	}
	return truncate(t.fallbackTitle, 22)
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-3]) + "..."
}

func newTab(id int, fallbackTitle, content string) tab {
	if fallbackTitle == "" {
		fallbackTitle = fmt.Sprintf("Note %d", id)
	}
	editor := textarea.New()
	editor.Prompt = ""
	editor.Placeholder = "Start typing..."
	editor.SetValue(content)
	editor.ShowLineNumbers = false
	editor.CharLimit = 0

	styles := editor.Styles()
	styles.Focused.Base = lipgloss.NewStyle().Foreground(colorInk)
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Blurred.Base = lipgloss.NewStyle().Foreground(colorInk)
	styles.Cursor.Color = colorAccent
	editor.SetStyles(styles)

	return tab{id: id, fallbackTitle: fallbackTitle, editor: editor}
}
