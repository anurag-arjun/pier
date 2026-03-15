package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

// Process manages a single pi RPC process.
type Process struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser

	EventCh chan Event  // parsed RPC events from stdout
	ErrCh   chan string // stderr lines
	DoneCh  chan error  // process exit (nil = clean exit)

	mu      sync.Mutex
	closed  bool
}

// ProcessConfig configures how to spawn pi.
type ProcessConfig struct {
	PiPath     string   // path to pi binary
	WorkDir    string   // working directory
	Mode       string   // "rpc" (always)
	ExtraFlags []string // additional CLI flags

	// Session resume
	SessionFile string // --session <path> for resume

	// Model override
	Model string // --model <model>
}

// StartProcess spawns pi in RPC mode and begins reading events.
func StartProcess(cfg ProcessConfig) (*Process, error) {
	args := []string{"--mode", "rpc"}

	if cfg.SessionFile != "" {
		args = append(args, "--session", cfg.SessionFile)
	}
	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}
	args = append(args, cfg.ExtraFlags...)

	piPath := cfg.PiPath
	if piPath == "" {
		piPath = "pi"
	}

	cmd := exec.Command(piPath, args...)
	cmd.Dir = cfg.WorkDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start pi: %w", err)
	}

	p := &Process{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		EventCh: make(chan Event, 128),
		ErrCh:   make(chan string, 32),
		DoneCh:  make(chan error, 1),
	}

	// Goroutine: decode JSONL events from stdout
	go func() {
		defer close(p.EventCh)
		ch := DecodeEvents(stdout)
		for evt := range ch {
			p.EventCh <- evt
		}
	}()

	// Goroutine: read stderr lines
	go func() {
		defer close(p.ErrCh)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				p.ErrCh <- line
			}
		}
	}()

	// Goroutine: wait for process exit
	go func() {
		err := cmd.Wait()
		p.DoneCh <- err
		close(p.DoneCh)
	}()

	return p, nil
}

// SendCommand marshals a command to JSON and writes it to pi's stdin.
func (p *Process) SendCommand(cmd interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("process is closed")
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	data = append(data, '\n')
	_, err = p.stdin.Write(data)
	if err != nil {
		return fmt.Errorf("write to stdin: %w", err)
	}

	return nil
}

// Stop gracefully stops the pi process.
// Sends abort, waits briefly, then kills if necessary.
func (p *Process) Stop() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	// Try abort first
	abortData, _ := json.Marshal(NewAbortCmd())
	abortData = append(abortData, '\n')
	p.stdin.Write(abortData) // ignore error, process may already be dead

	// Close stdin to signal EOF
	p.stdin.Close()

	// Wait with timeout
	select {
	case <-p.DoneCh:
		return nil
	case <-time.After(3 * time.Second):
		log.Printf("pi process did not exit gracefully, killing")
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
		<-p.DoneCh
		return nil
	}
}

// Running returns true if the process hasn't exited yet.
func (p *Process) Running() bool {
	select {
	case <-p.DoneCh:
		return false
	default:
		return true
	}
}
