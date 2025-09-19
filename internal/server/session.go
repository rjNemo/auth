package server

import "sync"

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

// IsAuthenticated reports whether a user is currently considered logged in.
func (m *SessionManager) IsAuthenticated() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.authenticated
}

// CurrentAccount returns the email associated with the active session.
func (m *SessionManager) CurrentAccount() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentAccount
}
