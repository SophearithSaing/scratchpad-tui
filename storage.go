package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

const autoSaveDelay = 500 * time.Millisecond

type diskStore struct {
	path         string
	mu           sync.Mutex
	lastRevision uint64
}

func newDiskStore() (*diskStore, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("error finding config directory %w", err)
	}
	return &diskStore{
		path: filepath.Join(dir, "scratchpad", "session.json"),
	}, nil
}

func defaultSession() session {
	return session{
		Active: 0,
		NextID: 2,
		Notes: []note{
			{
				ID:    1,
				Title: "Note 1",
			},
		},
	}
}

func (ds *diskStore) load() (session, error) {
	data, err := os.ReadFile(ds.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultSession(), nil
		}
		return session{}, fmt.Errorf("cannot read session: %w", err)
	}

	var state session
	if err := json.Unmarshal(data, &state); err != nil {
		return defaultSession(), fmt.Errorf("cannot decode session: %w", err)
	}
	return state, nil
}

func (ds *diskStore) save(revision uint64, state session) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if revision < ds.lastRevision {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(ds.path), 0o700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode session: %w", err)
	}
	if err := os.WriteFile(ds.path, data, 0o600); err != nil {
		return fmt.Errorf("cannot write session: %w", err)
	}

	ds.lastRevision = revision
	return nil
}

type saveDueMsg uint64

type saveFinishedMsg struct {
	revision uint64
	err      error
}

func scheduleSave(revision uint64) tea.Cmd {
	return tea.Tick(autoSaveDelay, func(time.Time) tea.Msg {
		return saveDueMsg(revision)
	})
}

func saveCmd(ds *diskStore, revision uint64, state session) tea.Cmd {
	snapshot := cloneSession(state)
	return func() tea.Msg {
		return saveFinishedMsg{
			revision: revision,
			err:      ds.save(revision, snapshot),
		}
	}
}
