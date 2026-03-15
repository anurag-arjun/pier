package ui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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

	toolBlocks []*widgets.ToolBlock
	richStates []*richtext.InteractiveText
}

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

func (t *Timeline) Layout(gtx layout.Context, state *session.SessionState) layout.Dimensions {
	state.Lock()
	entries := make([]session.TimelineEntry, len(state.Timeline))
	copy(entries, state.Timeline)
	status := state.Status
	state.Unlock()

	t.ensureWidgets(len(entries))

	return material.List(t.matTheme, &t.list).Layout(gtx, len(entries), func(gtx layout.Context, i int) layout.Dimensions {
		entry := entries[i]

		// Message grouping: tight gap for same-role consecutive, larger for role change
		gap := unit.Dp(16) // default: different role
		if i > 0 {
			prev := entries[i-1]
			if prev.Kind == entry.Kind {
				gap = unit.Dp(4) // same role → tight
			}
		}

		return layout.Inset{Bottom: gap}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				lbl := material.Label(t.matTheme, t.theme.Typo.Caption, "You")
				lbl.Color = t.theme.Palette.Accent
				lbl.Font.Weight = font.SemiBold
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(t.matTheme, t.theme.Typo.Body, entry.UserText)
				lbl.Color = t.theme.Palette.Text
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (t *Timeline) layoutAssistantMessage(gtx layout.Context, idx int, entry session.TimelineEntry, status session.Status) layout.Dimensions {
	// Left accent border + indented content
	return layout.Flex{}.Layout(gtx,
		// Left accent bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			barW := gtx.Dp(unit.Dp(2))
			barH := gtx.Constraints.Max.Y
			if barH <= 0 {
				barH = gtx.Dp(unit.Dp(32))
			}
			paint.FillShape(gtx.Ops, t.theme.Palette.Accent, clip.Rect(image.Rect(0, 0, barW, barH)).Op())
			return layout.Dimensions{Size: image.Pt(barW, barH)}
		}),
		// Content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Model badge
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						model := entry.Model
						if model == "" {
							model = "Assistant"
						}
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							// Pill badge for model name
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Stack{}.Layout(gtx,
									layout.Expanded(func(gtx layout.Context) layout.Dimensions {
										rr := gtx.Dp(t.theme.Space.BadgeRadius)
										size := gtx.Constraints.Min
										paint.FillShape(gtx.Ops, t.theme.Palette.SurfaceAlt, clip.UniformRRect(image.Rect(0, 0, size.X, size.Y), rr).Op(gtx.Ops))
										return layout.Dimensions{Size: size}
									}),
									layout.Stacked(func(gtx layout.Context) layout.Dimensions {
										return layout.Inset{Left: unit.Dp(6), Right: unit.Dp(6), Top: unit.Dp(2), Bottom: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											lbl := material.Label(t.matTheme, t.theme.Typo.MonoSmall, model)
											lbl.Color = t.theme.Palette.TextSecondary
											lbl.Font.Typeface = t.theme.Typo.MonoFace
											return lbl.Layout(gtx)
										})
									}),
								)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					// Thinking
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if entry.ThinkingText == "" {
							return layout.Dimensions{}
						}
						return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(t.matTheme, t.theme.Typo.BodySmall, entry.ThinkingText)
							lbl.Color = t.theme.Palette.TextTertiary
							lbl.Font.Style = font.Italic
							return lbl.Layout(gtx)
						})
					}),
					// Content: streaming=plain, complete=markdown
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := entry.AssistantText
						if text == "" && entry.Streaming {
							text = "▍"
						}
						if entry.Streaming {
							lbl := material.Label(t.matTheme, t.theme.Typo.Body, text)
							lbl.Color = t.theme.Palette.Text
							return lbl.Layout(gtx)
						}
						return t.mdRender.Layout(gtx, t.richStates[idx], text)
					}),
				)
			})
		}),
	)
}

func (t *Timeline) layoutToolCall(gtx layout.Context, idx int, entry session.TimelineEntry) layout.Dimensions {
	tb := t.toolBlocks[idx]
	tb.ToolName = entry.ToolName
	tb.Args = entry.ToolArgs
	tb.Result = entry.ToolResult
	tb.Partial = entry.ToolPartial
	tb.IsError = entry.ToolIsError
	return tb.Layout(gtx)
}

func (t *Timeline) layoutNotice(gtx layout.Context, text string, c color.NRGBA) layout.Dimensions {
	return layout.Inset{Left: unit.Dp(4), Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(t.matTheme, t.theme.Typo.Caption, text)
		lbl.Color = c
		return lbl.Layout(gtx)
	})
}

func (t *Timeline) CollapseAll() {
	for _, tb := range t.toolBlocks {
		tb.SetExpanded(false)
	}
}
