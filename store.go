package main

import (
	"errors"
	"sync"
)

type Store interface {
	// defines the methods required for a store to back
	// the queueing service
	Enqueue(destination string, message Frame) error
	Pop(destination string) (Frame, error)
	Len(destination string) (int, error)
	Destinations() []string
}

type MemoryStore struct {
	// Defines a basic in-memory queue store
	// Concurrency protected by sync.Mutex
	sync.Mutex
	Queues map[string][]Frame
}

func (m *MemoryStore) Enqueue(destination string, message Frame) error {
	m.Lock()
	defer m.Unlock()
	q, prs := m.Queues[destination]
	if prs {
		m.Queues[destination] = append(q, message)
		return nil
	} else {
		return errors.New("No such destination.")
	}
}

func (m *MemoryStore) Pop(destination string) (Frame, error) {
	m.Lock()
	defer m.Unlock()
	l, err := m.Len(destination)
	if l == 0 {
		return Frame{}, errors.New("Destination queue empty.")
	}
	if err != nil {
		return Frame{}, err
	}

	q := m.Queues[destination] // guaranteed to work by above conditions, hopefully
	f := q[0]
	m.Queues[destination] = q[1:]
	return f, nil
}

func (m *MemoryStore) Len(destination string) (int, error) {
	// do we need to lock the mutex here? if it is called within Pop, the mutex will be locked
	// but if it is called in another context it may not be
	// which could give us trouble. Will have to write with this in mind.
	q, prs := m.Queues[destination]
	if !prs {
		return -1, errors.New("No such destination.")
	}

	return len(q), nil
}

func (m *MemoryStore) Destinations() []string {
	// also not locking the Mutex for here for now
	// will reconsider if race conditions seem possible
	keys := make([]string, len(m.Queues))
	i := 0

	for k := range m.Queues {
		keys[i] = k
		i++
	}
	return keys
}
