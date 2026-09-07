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

func (s sessionStore) save(session savedSession) error {
	_, err := s.loadIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading previous index: %w", err)
	}

	current := make(map[int]bool, len(session.Tabs))
	for _, tab := range session.Tabs {
		current[tab.ID] = true
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	data = append(data, '\n')

	for _, tab := range session.Tabs {
		path := s.notePath(tab.ID)
		err = writeFile(path, []byte(tab.Content), true)
		if err != nil {
			return err
		}
	}
	if err := writeFile(s.path, data, true); err != nil {
		return err
	}
	return nil
}

func writeFile(path string, data []byte, replace bool) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".scratchpad-*")
	if err != nil {
		return fmt.Errorf("create temporary file for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("secure temporary file for %s: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temporary file for %s: %w", path, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temporary file for %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary file for %s: %w", path, err)
	}

	if replace {
		err = os.Rename(tmpName, path)
	} else {
		err = os.Link(tmpName, path)
	}
	if err != nil {
		return fmt.Errorf("publish %s: %w", path, err)
	}
	return nil
}
