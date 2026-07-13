package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const tabWidth = 4

var (
	editorBaseStyle             = lipgloss.NewStyle().Foreground(colorText).Background(colorSurface)
	editorMutedStyle            = lipgloss.NewStyle().Foreground(colorMuted).Background(colorSurface)
	editorLineNumberStyle       = lipgloss.NewStyle().Foreground(colorBorder).Background(colorSurface)
	editorActiveLineNumberStyle = lipgloss.NewStyle().Foreground(colorSelectionText).Background(colorSurface)
	editorSelectionStyle        = lipgloss.NewStyle().Foreground(colorSelectionText).Background(colorAccent)
	editorCursorStyle           = lipgloss.NewStyle().Bold(true).Foreground(colorSelectionText).Background(colorCyan)
)

type editorModel struct {
	value       []rune
	cursor      int
	anchor      int
	width       int
	height      int
	topLine     int
	leftCell    int
	goalCell    int
	placeholder string
}

func newEditor() editorModel {
	return editorModel{
		anchor:      -1,
		goalCell:    -1,
		placeholder: "Dump the thoughts before it escapes...",
	}
}

func (e editorModel) Value() string {
	return string(e.value)
}

func (e *editorModel) SetValue(value string) {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	e.value = []rune(value)
	e.cursor = len(e.value)
	e.anchor = -1
	e.topLine = 0
	e.leftCell = 0
	e.goalCell = -1
}

func (e editorModel) HasSelection() bool {
	return e.anchor >= 0 && e.anchor != e.cursor
}

func (e editorModel) SelectionBounds() (start, end int, ok bool) {
	if !e.HasSelection() {
		return 0, 0, false
	}
	if e.anchor < e.cursor {
		return e.anchor, e.cursor, true
	}
	return e.cursor, e.anchor, true
}

func (e editorModel) SelectedText() string {
	start, end, ok := e.SelectionBounds()
	if !ok {
		return ""
	}
	return string(e.value[start:end])
}
