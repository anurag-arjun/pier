// Package workspace manages workspace structs and persistence.
package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Message is a discovery conversation message.
type Message struct {
	Role      string  `json:"role"` // "user" | "assistant"
	Text      string  `json:"text"`
	Timestamp float64 `json:"timestamp"`
}

// SessionMeta is persisted metadata for a session (not the running process).
type SessionMeta struct {
	ID          string `json:"id"`
	TaskID      string `json:"task_id,omitempty"`
	Model       string `json:"model,omitempty"`
	SessionFile string `json:"session_file,omitempty"` // pi's session file for resume
	CreatedAt   string `json:"created_at"`
}

// Workspace is the central unit of the application.
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"` // filesystem path to the project

	Discovery struct {
		Messages []Message `json:"messages"`
		Model    string    `json:"model,omitempty"`
	} `json:"discovery"`

	Plan struct {
		Path      string `json:"path,omitempty"` // default: <workspace>/plan.md
		BrCreated bool   `json:"br_created"`
	} `json:"plan"`

	BrInitialised bool          `json:"br_initialised"`
	Sessions      []SessionMeta `json:"sessions"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

// NewWorkspace creates a workspace for the given path.
func NewWorkspace(id, name, path string) *Workspace {
	now := time.Now().Format(time.RFC3339)
	return &Workspace{
		ID:        id,
		Name:      name,
		Path:      path,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// PlanPath returns the plan.md path, defaulting to <workspace>/plan.md.
func (w *Workspace) PlanPath() string {
	if w.Plan.Path != "" {
		return w.Plan.Path
	}
	return filepath.Join(w.Path, "plan.md")
}

// Touch updates the UpdatedAt timestamp.
func (w *Workspace) Touch() {
	w.UpdatedAt = time.Now().Format(time.RFC3339)
}

// --- Persistence ---

// WorkspaceDir returns the workspaces storage directory.
func WorkspaceDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "pier", "workspaces"), nil
}

// Save persists a workspace to ~/.config/pier/workspaces/<id>.json.
func Save(w *Workspace) error {
	dir, err := WorkspaceDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	w.Touch()
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, w.ID+".json"), data, 0644)
}

// LoadAll loads all workspaces from ~/.config/pier/workspaces/.
func LoadAll() ([]*Workspace, error) {
	dir, err := WorkspaceDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var workspaces []*Workspace
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var w Workspace
		if err := json.Unmarshal(data, &w); err != nil {
			continue
		}
		workspaces = append(workspaces, &w)
	}

	return workspaces, nil
}

// LoadByID loads a single workspace by ID.
func LoadByID(id string) (*Workspace, error) {
	dir, err := WorkspaceDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, err
	}

	var w Workspace
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// Delete removes a workspace file.
func Delete(id string) error {
	dir, err := WorkspaceDir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, id+".json"))
}

// GenerateID creates a short unique workspace ID.
func GenerateID() string {
	return fmt.Sprintf("w%d", time.Now().UnixNano()%1000000)
}
