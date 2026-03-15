// Package br wraps the br CLI for task management.
package br

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client wraps the br binary for a workspace.
type Client struct {
	BrPath        string // resolved path to br binary
	WorkspacePath string // project directory containing .beads/
}

// NewClient creates a br client for the given workspace.
// brPath can be empty to auto-resolve.
func NewClient(workspacePath, brPath string) (*Client, error) {
	if brPath == "" {
		var err error
		brPath, err = ResolveBrBinary()
		if err != nil {
			return nil, err
		}
	}
	return &Client{
		BrPath:        brPath,
		WorkspacePath: workspacePath,
	}, nil
}

// ResolveBrBinary finds the br binary.
// Checks: PATH, ~/.cargo/bin/br, ~/.local/bin/br
func ResolveBrBinary() (string, error) {
	// Check PATH first
	if p, err := exec.LookPath("br"); err == nil {
		return p, nil
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		candidates := []string{
			filepath.Join(home, ".cargo", "bin", "br"),
			filepath.Join(home, ".local", "bin", "br"),
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && !info.IsDir() {
				return c, nil
			}
		}
	}

	return "", fmt.Errorf("br binary not found in PATH, ~/.cargo/bin/, or ~/.local/bin/")
}

// run executes a br command in the workspace directory and returns stdout.
func (c *Client) run(args ...string) ([]byte, error) {
	cmd := exec.Command(c.BrPath, args...)
	cmd.Dir = c.WorkspacePath
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("br %s: %s", strings.Join(args, " "), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("br %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// Init initializes br in the workspace if not already done.
func (c *Client) Init() error {
	_, err := c.run("init")
	return err
}

// IsInitialized checks if .beads/ exists in the workspace.
func (c *Client) IsInitialized() bool {
	info, err := os.Stat(filepath.Join(c.WorkspacePath, ".beads"))
	return err == nil && info.IsDir()
}

// List returns all tasks.
func (c *Client) List() ([]Task, error) {
	out, err := c.run("list", "--json")
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, fmt.Errorf("parse br list: %w", err)
	}
	return tasks, nil
}

// Ready returns unblocked tasks sorted by priority.
func (c *Client) Ready() ([]Task, error) {
	out, err := c.run("ready", "--json")
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, fmt.Errorf("parse br ready: %w", err)
	}
	return tasks, nil
}

// Show returns full detail for a single task.
func (c *Client) Show(id string) (*Task, error) {
	out, err := c.run("show", id, "--json")
	if err != nil {
		return nil, err
	}
	// br show --json returns an array with one element
	var tasks []Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, fmt.Errorf("parse br show: %w", err)
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return &tasks[0], nil
}

// UpdateStatus changes a task's status.
func (c *Client) UpdateStatus(id, status string) error {
	_, err := c.run("update", id, "--status", status)
	return err
}

// Close closes a task with a reason.
func (c *Client) Close(id, reason string) error {
	args := []string{"close", id}
	if reason != "" {
		args = append(args, "--reason", reason)
	}
	_, err := c.run(args...)
	return err
}
