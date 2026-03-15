package ui

import (
	"log"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/session"
	"github.com/lighto/pier/workspace"
)

// DiscoverView is the focused chat interface for discovery conversations.
type DiscoverView struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	session   *session.Session
	timeline  *Timeline
	prompt    *PromptBar

	genPlanBtn widget.Clickable

	// Persisted messages
	Messages []workspace.Message

	// Callbacks
	OnGeneratePlan func(transcript string)
}

// NewDiscoverView creates a discover view.
func NewDiscoverView(theme apptheme.Theme, matTheme *material.Theme) *DiscoverView {
	return &DiscoverView{
		theme:    theme,
		matTheme: matTheme,
		timeline: NewTimeline(theme, matTheme),
		prompt:   NewPromptBar(theme, matTheme),
	}
}

// SetSession sets the pi session for discover mode.
func (dv *DiscoverView) SetSession(sess *session.Session) {
	dv.session = sess
}

// HasSession returns whether a session is active.
func (dv *DiscoverView) HasSession() bool {
	return dv.session != nil
}

// DrainEvents drains discover session events.
func (dv *DiscoverView) DrainEvents() bool {
	if dv.session == nil {
		return false
	}
	return dv.session.DrainEvents()
}

// GetTranscript returns the full conversation as a formatted string.
func (dv *DiscoverView) GetTranscript() string {
	if dv.session == nil {
		return ""
	}
	dv.session.State.Lock()
	defer dv.session.State.Unlock()

	var text string
	for _, entry := range dv.session.State.Timeline {
		switch entry.Kind {
		case session.EntryUserMessage:
			text += "User:\n" + entry.UserText + "\n\n"
		case session.EntryAssistantMessage:
			text += "Assistant:\n" + entry.AssistantText + "\n\n"
		}
	}
	return text
}

// Layout renders the discover view.
func (dv *DiscoverView) Layout(gtx layout.Context) layout.Dimensions {
	// Check generate plan button
	if dv.genPlanBtn.Clicked(gtx) && dv.OnGeneratePlan != nil {
		transcript := dv.GetTranscript()
		if transcript != "" {
			dv.OnGeneratePlan(transcript)
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(dv.matTheme, "DISCOVER")
				lbl.Color = dv.theme.Palette.TextSecondary
				return lbl.Layout(gtx)
			})
		}),
		// Timeline
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if dv.session == nil {
				lbl := material.Body2(dv.matTheme, "Start a discovery conversation to think through your approach.")
				lbl.Color = dv.theme.Palette.TextSecondary
				return lbl.Layout(gtx)
			}
			return dv.timeline.Layout(gtx, dv.session.State)
		}),
		// Generate Plan button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if dv.session == nil {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(dv.matTheme, &dv.genPlanBtn, "Generate Plan →")
				btn.Background = dv.theme.Palette.Accent
				btn.Inset = layout.Inset{
					Left: unit.Dp(16), Right: unit.Dp(16),
					Top: unit.Dp(6), Bottom: unit.Dp(6),
				}
				return btn.Layout(gtx)
			})
		}),
		// Prompt bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			status := session.StatusIdle
			if dv.session != nil {
				dv.session.State.Lock()
				status = dv.session.State.Status
				dv.session.State.Unlock()
			}
			dims, submitted := dv.prompt.Layout(gtx, status)
			if submitted != "" && dv.session != nil {
				if err := dv.session.SendPrompt(submitted); err != nil {
					log.Printf("discover send error: %v", err)
				}
			}
			return dims
		}),
	)
}

// Stop stops the discover session.
func (dv *DiscoverView) Stop() {
	if dv.session != nil {
		dv.session.Stop()
		dv.session = nil
	}
}
