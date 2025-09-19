package server

import "sync"

// SessionState represents the snapshot of session metadata for a request.
type SessionState struct {
	Authenticated bool
	Email         string
}

// SessionManager is a placeholder for future session persistence.
type SessionManager struct {
	mu             sync.RWMutex
	authenticated  bool
	currentAccount string
}

// NewSessionManager constructs an empty session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// SetAuthenticated marks the provided account as the active authenticated user.
func (m *SessionManager) SetAuthenticated(email string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.authenticated = true
	m.currentAccount = email
}

// Clear removes any active authentication data.
func (m *SessionManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.authenticated = false
	m.currentAccount = ""
}

// Snapshot captures the current session state for contextual use.
func (m *SessionManager) Snapshot() SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return SessionState{
		Authenticated: m.authenticated,
		Email:         m.currentAccount,
	}
}
