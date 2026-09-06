package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type savedTab struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"-"`
}

type savedSession struct {
	Active int        `json:"active"`
	NextID int        `json:"nextId"`
	Tabs   []savedTab `json:"tabs"`
}

type sessionStore struct {
	path string
}

func newSessionStore() (*sessionStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("error finding config directory: %v", err)
	}
	return &sessionStore{path: filepath.Join(dir, "scratchpad", "session.json")}, nil
}
