package appssh

import "sync"

// Pool manages a single SSH session per host.
type Pool struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

// Global pool instance.
var Global = &Pool{sessions: make(map[string]*Session)}

// Get returns an existing session for key (user@host) or nil.
func (p *Pool) Get(key string) *Session {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sessions[key]
}

// Set stores a session.
func (p *Pool) Set(key string, s *Session) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[key] = s
}

// Remove closes and removes a session.
func (p *Pool) Remove(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if s, ok := p.sessions[key]; ok {
		s.Close()
		delete(p.sessions, key)
	}
}
