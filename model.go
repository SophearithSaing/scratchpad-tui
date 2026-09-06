package main

import (
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
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
	store       *sessionStore
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

func newModel(store *sessionStore, session *savedSession) *appModel {
	m := &appModel{store: store, nextID: 1, preview: viewport.New(), help: viewport.New()}
	m.preview.FillHeight = true
	m.help.SoftWrap = true
	m.help.FillHeight = true

	for _, saved := range session.Tabs {
		// TODO: append new tabs
		if saved.ID >= m.nextID {
			m.nextID = saved.ID + 1
		}
	}
	if session.NextID > m.nextID {
		m.nextID = session.NextID
	}
	if len(m.tabs) == 0 {
		// TODO: add tab
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
	// TODO: focus

	return m
}
