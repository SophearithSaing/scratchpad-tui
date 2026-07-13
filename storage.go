package main

import "sync"

type diskStore struct {
	path         string
	mu           sync.Mutex
	lastRevision uint64
}
