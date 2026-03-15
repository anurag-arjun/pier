package ui

import (
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/session"
	"github.com/lighto/pier/ui/widgets"
)

// SlashCommand is a pi command for autocomplete.
type SlashCommand struct {
	Name        string
	Description string
}

// PromptBar is the input widget at the bottom of a session view.
type PromptBar struct {
	theme    apptheme.Theme
	matTheme *material.Theme
	editor   widget.Editor
	submit   widget.Clickable

	commands       []SlashCommand
	filteredCmds   []SlashCommand
	showComplete   bool
	completeClicks []widget.Clickable

	hover widgets.HoverState
}

func NewPromptBar(theme apptheme.Theme, matTheme *material.Theme) *PromptBar {
	return &PromptBar{
		theme:    theme,
		matTheme: matTheme,
		editor: widget.Editor{
			SingleLine: true,
			Submit:     true,
		},
	}
}

func (pb *PromptBar) SetCommands(cmds []SlashCommand) { pb.commands = cmds }

func (pb *PromptBar) Layout(gtx layout.Context, status session.Status) (layout.Dimensions, string) {
	var submitted string

	for {
		evt, ok := pb.editor.Update(gtx)
		if !ok {
			break
		}
		if _, isSubmit := evt.(widget.SubmitEvent); isSubmit {
			submitted = pb.editor.Text()
			pb.editor.SetText("")
			pb.showComplete = false
		}
	}
	if pb.submit.Clicked(gtx) {
		text := pb.editor.Text()
		if text != "" {
			submitted = text
			pb.editor.SetText("")
			pb.showComplete = false
		}
	}
	for i := range pb.filteredCmds {
		if i < len(pb.completeClicks) && pb.completeClicks[i].Clicked(gtx) {
			pb.editor.SetText("/" + pb.filteredCmds[i].Name + " ")
			pb.showComplete = false
		}
	}

	// Autocomplete logic
	text := pb.editor.Text()
	if strings.HasPrefix(text, "/") && len(text) > 1 {
		query := strings.ToLower(text[1:])
		pb.filteredCmds = nil
		for _, cmd := range pb.commands {
			if strings.Contains(strings.ToLower(cmd.Name), query) {
				pb.filteredCmds = append(pb.filteredCmds, cmd)
			}
		}
		pb.showComplete = len(pb.filteredCmds) > 0
	} else if text == "/" {
		pb.filteredCmds = pb.commands
		pb.showComplete = len(pb.commands) > 0
	} else {
		pb.showComplete = false
	}
	for len(pb.completeClicks) < len(pb.filteredCmds) {
		pb.completeClicks = append(pb.completeClicks, widget.Clickable{})
	}

	disabled := status != session.StatusWaiting && status != session.StatusIdle
	isActive := status == session.StatusThinking || status == session.StatusStreaming

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Status line above prompt
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !isActive {
				return layout.Dimensions{}
			}
			return layout.Inset{Bottom: unit.Dp(4), Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					// Status dot
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						s := "thinking"
						if status == session.StatusStreaming {
							s = "streaming"
						}
						return layoutStatusDot(gtx, pb.theme, s)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "Thinking..."
						if status == session.StatusStreaming {
							label = "Streaming..."
						}
						lbl := material.Label(pb.matTheme, pb.theme.Typo.Caption, label)
						lbl.Color = pb.theme.Palette.TextTertiary
						return lbl.Layout(gtx)
					}),
				)
			})
		}),
		// Autocomplete popup
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !pb.showComplete {
				return layout.Dimensions{}
			}
			return pb.layoutAutocomplete(gtx)
		}),
		// Input container
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return pb.layoutInput(gtx, disabled)
		}),
	)
	return dims, submitted
}

func (pb *PromptBar) layoutAutocomplete(gtx layout.Context) layout.Dimensions {
	maxShow := 8
	cmds := pb.filteredCmds
	if len(cmds) > maxShow {
		cmds = cmds[:maxShow]
	}

	return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				rr := gtx.Dp(pb.theme.Space.BorderRadius)
				size := gtx.Constraints.Min
				widgets.DrawBorderedRect(gtx, pb.theme.Palette.Surface, pb.theme.Palette.Border, size, rr, 1)
				return layout.Dimensions{Size: size}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					list := layout.List{Axis: layout.Vertical}
					return list.Layout(gtx, len(cmds), func(gtx layout.Context, i int) layout.Dimensions {
						cmd := cmds[i]
						return material.Clickable(gtx, &pb.completeClicks[i], func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12), Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(pb.matTheme, pb.theme.Typo.BodySmall, "/"+cmd.Name)
										lbl.Color = pb.theme.Palette.Accent
										lbl.Font.Weight = font.Medium
										lbl.Font.Typeface = pb.theme.Typo.MonoFace
										return lbl.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										if cmd.Description == "" {
											return layout.Dimensions{}
										}
										lbl := material.Label(pb.matTheme, pb.theme.Typo.Caption, cmd.Description)
										lbl.Color = pb.theme.Palette.TextTertiary
										lbl.MaxLines = 1
										return lbl.Layout(gtx)
									}),
								)
							})
						})
					})
				})
			}),
		)
	})
}

func (pb *PromptBar) layoutInput(gtx layout.Context, disabled bool) layout.Dimensions {
	borderColor := pb.theme.Palette.Border
	// TODO: track focus state via editor events when Gio adds Focused() API
	if disabled {
		borderColor = widgets.WithAlpha(pb.theme.Palette.Border, 100)
	}

	hasText := pb.editor.Text() != ""

	return layout.Stack{}.Layout(gtx,
		// Bordered container
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(unit.Dp(8))
			size := gtx.Constraints.Min
			if size.X == 0 {
				size.X = gtx.Constraints.Max.X
			}
			if size.Y == 0 {
				size.Y = gtx.Constraints.Max.Y
			}
			widgets.DrawBorderedRect(gtx, pb.theme.Palette.Surface, borderColor, size, rr, 1)
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			minH := gtx.Dp(unit.Dp(40))
			return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(8), Top: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if gtx.Constraints.Min.Y < minH {
					gtx.Constraints.Min.Y = minH
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						hint := "Message pi..."
						if disabled {
							hint = "Agent is working..."
						}
						e := material.Editor(pb.matTheme, &pb.editor, hint)
						e.Color = pb.theme.Palette.Text
						e.HintColor = pb.theme.Palette.TextTertiary
						if disabled {
							e.Color = pb.theme.Palette.TextDisabled
						}
						return e.Layout(gtx)
					}),
					// Send button (only when text present)
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if !hasText || disabled {
							return layout.Dimensions{}
						}
						return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(pb.matTheme, &pb.submit, "→")
							btn.Background = pb.theme.Palette.Accent
							btn.CornerRadius = unit.Dp(16)
							btn.Inset = layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10), Top: unit.Dp(4), Bottom: unit.Dp(4)}
							return btn.Layout(gtx)
						})
					}),
				)
			})
		}),
	)
}

func (pb *PromptBar) Editor() *widget.Editor { return &pb.editor }
