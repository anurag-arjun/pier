// Package session manages pi process lifecycle and RPC event streams.
package session

import (
	"fmt"
	"sync"
)

// Session represents a running pi session with its process and UI state.
type Session struct {
	ID          string
	WorkspaceID string
	TaskID      string // optional br task link

	State   *SessionState
	Process *Process

	mu      sync.Mutex
	stopped bool
}

// Config for creating a new session.
type Config struct {
	ID          string
	WorkspaceID string
	TaskID      string
	PiPath      string
	WorkDir     string
	Model       string
	SessionFile string // for resume
	ExtraFlags  []string
}

// New creates and starts a new pi session.
func New(cfg Config) (*Session, error) {
	proc, err := StartProcess(ProcessConfig{
		PiPath:      cfg.PiPath,
		WorkDir:     cfg.WorkDir,
		Mode:        "rpc",
		Model:       cfg.Model,
		SessionFile: cfg.SessionFile,
		ExtraFlags:  cfg.ExtraFlags,
	})
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	s := &Session{
		ID:          cfg.ID,
		WorkspaceID: cfg.WorkspaceID,
		TaskID:      cfg.TaskID,
		State:       NewSessionState(),
		Process:     proc,
	}

	return s, nil
}

// SendPrompt sends a user prompt to the pi session.
func (s *Session) SendPrompt(text string) error {
	s.State.AddUserMessage(text)
	return s.Process.SendCommand(NewPromptCmd(text))
}

// SendCommand sends an arbitrary command to the pi session.
func (s *Session) SendCommand(cmd interface{}) error {
	return s.Process.SendCommand(cmd)
}

// Stop gracefully stops the session.
func (s *Session) Stop() error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	s.mu.Unlock()

	return s.Process.Stop()
}

// DrainEvents reads all pending events and processes them.
// Returns true if any event caused a UI-relevant state change.
func (s *Session) DrainEvents() bool {
	changed := false
	for {
		select {
		case evt, ok := <-s.Process.EventCh:
			if !ok {
				return changed
			}
			if s.State.ProcessEvent(evt) {
				changed = true
			}
		default:
			return changed
		}
	}
}

// DrainErrors reads all pending stderr lines.
func (s *Session) DrainErrors() []string {
	var errs []string
	for {
		select {
		case line, ok := <-s.Process.ErrCh:
			if !ok {
				return errs
			}
			errs = append(errs, line)
		default:
			return errs
		}
	}
}

// Running returns true if the pi process is still alive.
func (s *Session) Running() bool {
	return s.Process.Running()
}

// RequestState sends get_state to refresh model/session info.
func (s *Session) RequestState() error {
	return s.Process.SendCommand(NewGetStateCmd())
}
