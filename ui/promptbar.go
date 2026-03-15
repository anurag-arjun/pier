package ui

import (
	"image"
	"strings"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
	"github.com/lighto/pier/session"
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

	// Autocomplete
	commands       []SlashCommand
	filteredCmds   []SlashCommand
	showComplete   bool
	completeClicks []widget.Clickable
}

// NewPromptBar creates a prompt bar.
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

// SetCommands sets the available slash commands for autocomplete.
func (pb *PromptBar) SetCommands(cmds []SlashCommand) {
	pb.commands = cmds
}

// Layout renders the prompt bar. Returns submitted text (empty if nothing submitted).
func (pb *PromptBar) Layout(gtx layout.Context, status session.Status) (layout.Dimensions, string) {
	var submitted string

	// Check for submit events
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

	// Check button click
	if pb.submit.Clicked(gtx) {
		text := pb.editor.Text()
		if text != "" {
			submitted = text
			pb.editor.SetText("")
			pb.showComplete = false
		}
	}

	// Check autocomplete clicks
	for i := range pb.filteredCmds {
		if i < len(pb.completeClicks) && pb.completeClicks[i].Clicked(gtx) {
			pb.editor.SetText("/" + pb.filteredCmds[i].Name + " ")
			pb.showComplete = false
		}
	}

	// Update autocomplete state
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

	// Ensure enough clickables
	for len(pb.completeClicks) < len(pb.filteredCmds) {
		pb.completeClicks = append(pb.completeClicks, widget.Clickable{})
	}

	disabled := status != session.StatusWaiting && status != session.StatusIdle

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Autocomplete popup (above input)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !pb.showComplete {
				return layout.Dimensions{}
			}
			return pb.layoutAutocomplete(gtx)
		}),
		// Input bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return pb.layoutInput(gtx, disabled)
			})
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

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(unit.Dp(6))
			bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
			rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
			defer rrect.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, pb.theme.Palette.Surface)
			return layout.Dimensions{Size: bounds.Max}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(4), Bottom: unit.Dp(4),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				list := layout.List{Axis: layout.Vertical}
				return list.Layout(gtx, len(cmds), func(gtx layout.Context, i int) layout.Dimensions {
					cmd := cmds[i]
					return material.Clickable(gtx, &pb.completeClicks[i], func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Left: unit.Dp(12), Right: unit.Dp(12),
							Top: unit.Dp(3), Bottom: unit.Dp(3),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Body2(pb.matTheme, "/"+cmd.Name)
									lbl.Color = pb.theme.Palette.Accent
									lbl.Font.Weight = 700
									return lbl.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									if cmd.Description == "" {
										return layout.Dimensions{}
									}
									lbl := material.Caption(pb.matTheme, cmd.Description)
									lbl.Color = pb.theme.Palette.TextSecondary
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
}

func (pb *PromptBar) layoutInput(gtx layout.Context, disabled bool) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(unit.Dp(6))
			bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
			rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
			defer rrect.Push(gtx.Ops).Pop()
			c := pb.theme.Palette.Border
			if disabled {
				c.A = 100
			}
			paint.Fill(gtx.Ops, c)
			return layout.Dimensions{Size: bounds.Max}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(1)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx layout.Context) layout.Dimensions {
						rr := gtx.Dp(unit.Dp(5))
						bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
						rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
						defer rrect.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, pb.theme.Palette.Surface)
						return layout.Dimensions{Size: bounds.Max}
					}),
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Left: unit.Dp(12), Right: unit.Dp(12),
							Top: unit.Dp(10), Bottom: unit.Dp(10),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									hint := "Send a message..."
									if disabled {
										hint = "Agent is working..."
									}
									e := material.Editor(pb.matTheme, &pb.editor, hint)
									e.Color = pb.theme.Palette.Text
									e.HintColor = pb.theme.Palette.TextSecondary
									if disabled {
										e.Color = pb.theme.Palette.TextSecondary
									}
									return e.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(pb.matTheme, &pb.submit, "→")
									btn.Background = pb.theme.Palette.Accent
									if disabled {
										btn.Background = pb.theme.Palette.Border
									}
									btn.Inset = layout.Inset{
										Left: unit.Dp(12), Right: unit.Dp(12),
										Top: unit.Dp(4), Bottom: unit.Dp(4),
									}
									return btn.Layout(gtx)
								}),
							)
						})
					}),
				)
			})
		}),
	)
}

// Editor returns the underlying editor widget.
func (pb *PromptBar) Editor() *widget.Editor {
	return &pb.editor
}
