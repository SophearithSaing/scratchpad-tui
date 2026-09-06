package main

import (
	"encoding/json"
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

func newSessionStore() (sessionStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return sessionStore{}, fmt.Errorf("error finding config directory: %v", err)
	}
	return sessionStore{path: filepath.Join(dir, "scratchpad-dev", "session.json")}, nil
}

func (s sessionStore) load() (savedSession, error) {
	session, err := s.loadIndex()
	if os.IsNotExist(err) {
		return savedSession{}, nil
	}
	if err != nil {
		return savedSession{}, err
	}
	for i := range session.Tabs {
		path := s.notePath(session.Tabs[i].ID)
		data, err := os.ReadFile(path)
		if err != nil {
			return savedSession{}, fmt.Errorf("error reading note %s: %w", path, err)
		}
		session.Tabs[i].Content = string(data)
	}
	return session, nil
}

func (s sessionStore) loadIndex() (savedSession, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return savedSession{}, err
	}

	var session savedSession
	if err := json.Unmarshal(data, &session); err != nil {
		return savedSession{}, fmt.Errorf("error decoding %s: %w", s.path, err)
	}
	return session, nil
}

func (s sessionStore) notePath(id int) string {
	return filepath.Join(filepath.Dir(s.path), fmt.Sprintf("note-%d.md", id))
}
