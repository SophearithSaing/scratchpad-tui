package main

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/viewport"
)

type note struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type session struct {
	Active int    `json:"active"`
	NextID int    `json:"next_id"`
	Notes  []note `json:"notes"`
}

type viewMode int

const (
	modeEdit viewMode = iota
	modePreview
)

type model struct {
	width    int
	height   int
	ready    bool
	mode     viewMode
	state    session
	editor   editorModel
	preview  viewport.Model
	store    *diskStore
	revision uint64
	status   string
}

func titleFor(body string, id int) string {
	for _, line := range strings.Split(body, "\n") {
		title := strings.TrimSpace(line)
		title = strings.TrimLeft(title, "#>*-` ")
		if title == "" {
			continue
		}
		runes := []rune(title)
		if len(runes) > 22 {
			title = string(runes[:21]) + "..."
		}
		return title
	}
	return fmt.Sprintf("Note %d", id)
}

func newModel(state session, store *diskStore, loadErr error) model {
	editor := newEditor()
	editor.SetValue(state.Notes[state.Active].Body)

	preview := viewport.New()
	preview.SoftWrap = false
	preview.FillHeight = true
	preview.Style = lipgloss.NewStyle().Background(color.Transparent)

	status := "restore session"
	if loadErr != nil {
		status = "session was not readable, started fresh"
	}

	m := model{
		mode:    modeEdit,
		state:   state,
		editor:  editor,
		preview: preview,
		store:   store,
		status:  status,
	}
	return m
}
