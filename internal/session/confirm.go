package session

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// PendingAction represents an action awaiting confirmation.
type PendingAction struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      string    `json:"args"`
	CreatedAt time.Time `json:"createdAt"`
	done      chan bool // true=confirmed, false=denied
}

// ConfirmStore manages pending actions that require confirmation.
type ConfirmStore struct {
	mu      sync.Mutex
	pending map[string]*PendingAction
	actions map[string]bool // set of action categories requiring confirmation
	timeout time.Duration
}

// NewConfirmStore creates a new confirmation store.
// categories is a set of action names that require confirmation.
func NewConfirmStore(categories []string, timeout time.Duration) *ConfirmStore {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	actions := make(map[string]bool)
	for _, c := range categories {
		actions[c] = true
	}
	return &ConfirmStore{
		pending: make(map[string]*PendingAction),
		actions: actions,
		timeout: timeout,
	}
}

// NeedsConfirmation checks if a command category requires confirmation.
func (cs *ConfirmStore) NeedsConfirmation(command string) bool {
	if len(cs.actions) == 0 {
		return false
	}
	return cs.actions[command]
}

// Add registers a pending action and returns its ID.
func (cs *ConfirmStore) Add(command, argsDesc string) *PendingAction {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	id := generateConfirmID()
	pa := &PendingAction{
		ID:        id,
		Command:   command,
		Args:      argsDesc,
		CreatedAt: time.Now(),
		done:      make(chan bool, 1),
	}
	cs.pending[id] = pa

	// Auto-deny after timeout
	go func() {
		time.Sleep(cs.timeout)
		cs.mu.Lock()
		if _, ok := cs.pending[id]; ok {
			delete(cs.pending, id)
			select {
			case pa.done <- false:
			default:
			}
		}
		cs.mu.Unlock()
	}()

	return pa
}

// Confirm approves a pending action.
func (cs *ConfirmStore) Confirm(id string) error {
	cs.mu.Lock()
	pa, ok := cs.pending[id]
	if ok {
		delete(cs.pending, id)
	}
	cs.mu.Unlock()

	if !ok {
		return fmt.Errorf("confirmation %q not found or expired", id)
	}
	select {
	case pa.done <- true:
	default:
	}
	return nil
}

// Deny rejects a pending action.
func (cs *ConfirmStore) Deny(id string) error {
	cs.mu.Lock()
	pa, ok := cs.pending[id]
	if ok {
		delete(cs.pending, id)
	}
	cs.mu.Unlock()

	if !ok {
		return fmt.Errorf("confirmation %q not found or expired", id)
	}
	select {
	case pa.done <- false:
	default:
	}
	return nil
}

// Wait blocks until the pending action is confirmed or denied.
// Returns true if confirmed, false if denied or timed out.
func (pa *PendingAction) Wait() bool {
	return <-pa.done
}

func generateConfirmID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("c_%x", b)
}
