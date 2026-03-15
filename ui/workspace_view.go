package ui

import (
	"fmt"
	"image"
	"log"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	brclient "github.com/lighto/pier/br"
	"github.com/lighto/pier/session"
)

// WorkspaceView manages multiple sessions and their display within a workspace.
type WorkspaceView struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	sessions      []*session.Session
	activeIdx     int
	timelines     []*Timeline
	prompts       []*PromptBar

	newSessBtn    widget.Clickable

	// Task linking
	BrClient      *brclient.Client

	// Callbacks
	OnSessionStatusChange func(sessions []SessionEntry)
}

// NewWorkspaceView creates a workspace view.
func NewWorkspaceView(theme apptheme.Theme, matTheme *material.Theme) *WorkspaceView {
	return &WorkspaceView{
		theme:    theme,
		matTheme: matTheme,
	}
}

// AddSession adds and starts a new session.
func (wv *WorkspaceView) AddSession(sess *session.Session) {
	wv.sessions = append(wv.sessions, sess)
	wv.timelines = append(wv.timelines, NewTimeline(wv.theme, wv.matTheme))
	wv.prompts = append(wv.prompts, NewPromptBar(wv.theme, wv.matTheme))
	wv.activeIdx = len(wv.sessions) - 1
}

// ActiveSession returns the currently focused session, or nil.
func (wv *WorkspaceView) ActiveSession() *session.Session {
	if wv.activeIdx >= 0 && wv.activeIdx < len(wv.sessions) {
		return wv.sessions[wv.activeIdx]
	}
	return nil
}

// SetActiveIndex switches to a different session.
func (wv *WorkspaceView) SetActiveIndex(idx int) {
	if idx >= 0 && idx < len(wv.sessions) {
		wv.activeIdx = idx
	}
}

// SessionCount returns number of sessions.
func (wv *WorkspaceView) SessionCount() int {
	return len(wv.sessions)
}

// Sessions returns session entries for the sidebar.
func (wv *WorkspaceView) SessionEntries() []SessionEntry {
	var entries []SessionEntry
	for _, s := range wv.sessions {
		s.State.Lock()
		status := s.State.Status.String()
		model := s.State.Model
		s.State.Unlock()
		entries = append(entries, SessionEntry{
			ID:     s.ID,
			Model:  model,
			Status: status,
			TaskID: s.TaskID,
		})
	}
	return entries
}

// StartSessionFromTask creates a session linked to a br task.
func (wv *WorkspaceView) StartSessionFromTask(task brclient.Task, cfg session.Config) error {
	cfg.TaskID = task.ID
	sess, err := session.New(cfg)
	if err != nil {
		return err
	}

	// Mark task in progress
	if wv.BrClient != nil {
		if err := wv.BrClient.UpdateStatus(task.ID, "in_progress"); err != nil {
			log.Printf("failed to mark task in_progress: %v", err)
		}
	}

	wv.AddSession(sess)
	sess.RequestState()
	return nil
}

// CloseSession closes and removes a session by index.
// Returns the linked task ID if any (for close-task dialog).
func (wv *WorkspaceView) CloseSession(idx int) (taskID string) {
	if idx < 0 || idx >= len(wv.sessions) {
		return ""
	}
	sess := wv.sessions[idx]
	taskID = sess.TaskID
	sess.Stop()

	// Remove from slices
	wv.sessions = append(wv.sessions[:idx], wv.sessions[idx+1:]...)
	wv.timelines = append(wv.timelines[:idx], wv.timelines[idx+1:]...)
	wv.prompts = append(wv.prompts[:idx], wv.prompts[idx+1:]...)

	if wv.activeIdx >= len(wv.sessions) {
		wv.activeIdx = len(wv.sessions) - 1
	}
	return taskID
}

// DrainAllEvents drains events from all sessions. Returns true if any changed.
func (wv *WorkspaceView) DrainAllEvents() bool {
	changed := false
	for _, s := range wv.sessions {
		if s.DrainEvents() {
			changed = true
		}
		for _, line := range s.DrainErrors() {
			log.Printf("pi stderr [%s]: %s", s.ID, line)
		}
	}
	return changed
}

// AnyRunning returns true if any session's process is still alive.
func (wv *WorkspaceView) AnyRunning() bool {
	for _, s := range wv.sessions {
		if s.Running() {
			return true
		}
	}
	return false
}

// StopAll stops all sessions.
func (wv *WorkspaceView) StopAll() {
	for _, s := range wv.sessions {
		s.Stop()
	}
}

// Layout renders the active session's timeline and prompt bar.
func (wv *WorkspaceView) Layout(gtx layout.Context) layout.Dimensions {
	if len(wv.sessions) == 0 {
		return layout.Dimensions{}
	}

	sess := wv.sessions[wv.activeIdx]
	timeline := wv.timelines[wv.activeIdx]
	prompt := wv.prompts[wv.activeIdx]

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Session tabs (if multiple)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(wv.sessions) <= 1 {
				return layout.Dimensions{}
			}
			return wv.layoutSessionTabs(gtx)
		}),
		// Status bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			sess.State.Lock()
			status := sess.State.Status.String()
			model := sess.State.Model
			sess.State.Unlock()

			text := status
			if model != "" {
				text = model + "  ·  " + status
			}
			if sess.TaskID != "" {
				text += "  ·  " + sess.TaskID
			}

			return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(wv.matTheme, text)
				lbl.Color = wv.theme.Palette.TextSecondary
				return lbl.Layout(gtx)
			})
		}),
		// Timeline
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return timeline.Layout(gtx, sess.State)
		}),
		// Prompt bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			sess.State.Lock()
			status := sess.State.Status
			sess.State.Unlock()

			dims, submitted := prompt.Layout(gtx, status)
			if submitted != "" {
				if err := sess.SendPrompt(submitted); err != nil {
					log.Printf("send prompt error: %v", err)
				}
			}
			return dims
		}),
	)
}

func (wv *WorkspaceView) layoutSessionTabs(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		list := layout.List{Axis: layout.Horizontal}
		return list.Layout(gtx, len(wv.sessions), func(gtx layout.Context, i int) layout.Dimensions {
			sess := wv.sessions[i]
			isActive := i == wv.activeIdx

			sess.State.Lock()
			model := sess.State.Model
			sess.State.Unlock()

			label := fmt.Sprintf("Session %d", i+1)
			if model != "" {
				label = model
			}

			return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx layout.Context) layout.Dimensions {
						if !isActive {
							return layout.Dimensions{}
						}
						rr := gtx.Dp(unit.Dp(3))
						bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
						rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
						defer rrect.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, wv.theme.Palette.SurfaceAlt)
						return layout.Dimensions{Size: bounds.Max}
					}),
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Left: unit.Dp(8), Right: unit.Dp(8),
							Top: unit.Dp(4), Bottom: unit.Dp(4),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Caption(wv.matTheme, label)
							if isActive {
								lbl.Color = wv.theme.Palette.Text
							} else {
								lbl.Color = wv.theme.Palette.TextSecondary
							}
							return lbl.Layout(gtx)
						})
					}),
				)
			})
		})
	})
}
