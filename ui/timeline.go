package ui

import (
	"fmt"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/session"
	"github.com/lighto/pier/ui/widgets"
)

// Timeline renders a session's message/tool call history.
type Timeline struct {
	theme    apptheme.Theme
	matTheme *material.Theme
	list     widget.List
	mdRender *widgets.MarkdownRenderer

	// Per-entry state (grows as timeline grows)
	toolBlocks  []*widgets.ToolBlock
	richStates  []*richtext.InteractiveText
	entryCount  int
}

// NewTimeline creates a timeline renderer.
func NewTimeline(theme apptheme.Theme, matTheme *material.Theme) *Timeline {
	return &Timeline{
		theme:    theme,
		matTheme: matTheme,
		list: widget.List{
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: true,
			},
		},
		mdRender: widgets.NewMarkdownRenderer(
			matTheme.Shaper,
			theme.Palette.Text,
			theme.Palette.Accent,
		),
	}
}

// Layout renders the timeline from session state.
func (t *Timeline) Layout(gtx layout.Context, state *session.SessionState) layout.Dimensions {
	state.Lock()
	entries := make([]session.TimelineEntry, len(state.Timeline))
	copy(entries, state.Timeline)
	status := state.Status
	state.Unlock()

	// Ensure we have enough per-entry widgets
	t.ensureWidgets(len(entries))

	return material.List(t.matTheme, &t.list).Layout(gtx, len(entries), func(gtx layout.Context, i int) layout.Dimensions {
		entry := entries[i]
		return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			switch entry.Kind {
			case session.EntryUserMessage:
				return t.layoutUserMessage(gtx, entry)
			case session.EntryAssistantMessage:
				return t.layoutAssistantMessage(gtx, i, entry, status)
			case session.EntryToolCall:
				return t.layoutToolCall(gtx, i, entry)
			case session.EntryCompactionNotice:
				return t.layoutNotice(gtx, "⟳ Compacting: "+entry.CompactReason, t.theme.Palette.StatusCompacting)
			case session.EntryRetryNotice:
				msg := fmt.Sprintf("↻ Retry %d/%d: %s", entry.RetryAttempt, entry.RetryMaxAttempts, entry.RetryError)
				return t.layoutNotice(gtx, msg, t.theme.Palette.StatusRetrying)
			default:
				return layout.Dimensions{}
			}
		})
	})
}

func (t *Timeline) ensureWidgets(n int) {
	for len(t.toolBlocks) < n {
		t.toolBlocks = append(t.toolBlocks, widgets.NewToolBlock(t.theme, t.matTheme))
	}
	for len(t.richStates) < n {
		t.richStates = append(t.richStates, &richtext.InteractiveText{})
	}
}

func (t *Timeline) layoutUserMessage(gtx layout.Context, entry session.TimelineEntry) layout.Dimensions {
	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(t.matTheme, "You")
				lbl.Color = t.theme.Palette.Accent
				lbl.Font.Weight = 700
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(t.matTheme, entry.UserText)
				lbl.Color = t.theme.Palette.Text
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (t *Timeline) layoutAssistantMessage(gtx layout.Context, idx int, entry session.TimelineEntry, status session.Status) layout.Dimensions {
	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Model label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Assistant"
				if entry.Model != "" {
					label = entry.Model
				}
				lbl := material.Caption(t.matTheme, label)
				lbl.Color = t.theme.Palette.StatusStreaming
				lbl.Font.Weight = 700
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
			// Thinking block (collapsed, secondary)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if entry.ThinkingText == "" {
					return layout.Dimensions{}
				}
				return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(t.matTheme, entry.ThinkingText)
					lbl.Color = t.theme.Palette.TextSecondary
					return lbl.Layout(gtx)
				})
			}),
			// Content: streaming = plain text, complete = markdown
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := entry.AssistantText
				if text == "" && entry.Streaming {
					text = "…"
				}
				if entry.Streaming {
					// Plain text during streaming (no markdown parse)
					lbl := material.Body1(t.matTheme, text)
					lbl.Color = t.theme.Palette.Text
					return lbl.Layout(gtx)
				}
				// Complete: render as markdown
				return t.mdRender.Layout(gtx, t.richStates[idx], text)
			}),
		)
	})
}

func (t *Timeline) layoutToolCall(gtx layout.Context, idx int, entry session.TimelineEntry) layout.Dimensions {
	tb := t.toolBlocks[idx]
	tb.ToolName = entry.ToolName
	tb.Args = entry.ToolArgs
	tb.Result = entry.ToolResult
	tb.Partial = entry.ToolPartial
	tb.IsError = entry.ToolIsError

	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return tb.Layout(gtx)
	})
}

func (t *Timeline) layoutNotice(gtx layout.Context, text string, c color.NRGBA) layout.Dimensions {
	return layout.Inset{Left: unit.Dp(4), Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Caption(t.matTheme, text)
		lbl.Color = c
		return lbl.Layout(gtx)
	})
}

// CollapseAll collapses all expanded tool blocks.
func (t *Timeline) CollapseAll() {
	for _, tb := range t.toolBlocks {
		tb.SetExpanded(false)
	}
}
