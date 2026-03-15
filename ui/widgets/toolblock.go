package widgets

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
)

// ToolBlock renders a collapsible tool call block.
type ToolBlock struct {
	theme    apptheme.Theme
	matTheme *material.Theme

	ToolName string
	Args     string
	Result   string
	Partial  string
	IsError  bool

	expanded bool
	toggle   widget.Clickable
}

// NewToolBlock creates a tool block widget.
func NewToolBlock(theme apptheme.Theme, matTheme *material.Theme) *ToolBlock {
	return &ToolBlock{
		theme:    theme,
		matTheme: matTheme,
	}
}

// SetExpanded sets the expanded state.
func (tb *ToolBlock) SetExpanded(v bool) {
	tb.expanded = v
}

// Expanded returns whether the block is expanded.
func (tb *ToolBlock) Expanded() bool {
	return tb.expanded
}

// Layout renders the tool block.
func (tb *ToolBlock) Layout(gtx layout.Context) layout.Dimensions {
	// Toggle on click
	if tb.toggle.Clicked(gtx) {
		tb.expanded = !tb.expanded
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header (always visible, clickable)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Clickable(gtx, &tb.toggle, func(gtx layout.Context) layout.Dimensions {
				return tb.layoutHeader(gtx)
			})
		}),
		// Body (only when expanded)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !tb.expanded {
				return layout.Dimensions{}
			}
			return tb.layoutBody(gtx)
		}),
	)
}

func (tb *ToolBlock) layoutHeader(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		// Background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(unit.Dp(4))
			bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
			rrect := clip.RRect{Rect: bounds, NE: rr, NW: rr, SE: rr, SW: rr}
			defer rrect.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, tb.theme.Palette.ToolBlockBg)
			return layout.Dimensions{Size: bounds.Max}
		}),
		// Content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left: unit.Dp(8), Right: unit.Dp(8),
				Top: unit.Dp(5), Bottom: unit.Dp(5),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					// Expand indicator
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						arrow := "▸"
						if tb.expanded {
							arrow = "▾"
						}
						lbl := material.Caption(tb.matTheme, arrow)
						lbl.Color = tb.theme.Palette.TextSecondary
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
					// Tool name
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Caption(tb.matTheme, "["+tb.ToolName+"]")
						lbl.Color = tb.theme.Palette.Accent
						lbl.Font.Weight = 700
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					// Args summary
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						args := tb.Args
						if len(args) > 80 {
							args = args[:80] + "…"
						}
						lbl := material.Caption(tb.matTheme, args)
						lbl.Color = tb.theme.Palette.TextSecondary
						lbl.MaxLines = 1
						return lbl.Layout(gtx)
					}),
					// Error badge
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if !tb.IsError {
							return layout.Dimensions{}
						}
						return layout.Inset{Left: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Caption(tb.matTheme, "ERROR")
							lbl.Color = tb.theme.Palette.Error
							lbl.Font.Weight = 700
							return lbl.Layout(gtx)
						})
					}),
				)
			})
		}),
	)
}

func (tb *ToolBlock) layoutBody(gtx layout.Context) layout.Dimensions {
	result := tb.Result
	if result == "" {
		result = tb.Partial
	}
	if result == "" {
		return layout.Dimensions{}
	}

	// Truncate very long output for display
	if len(result) > 4000 {
		result = result[:4000] + "\n\n… (truncated, expand full output)"
	}

	return layout.Stack{}.Layout(gtx,
		// Background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(unit.Dp(4))
			bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
			rrect := clip.RRect{Rect: bounds, NE: 0, NW: 0, SE: rr, SW: rr}
			defer rrect.Push(gtx.Ops).Pop()
			c := tb.theme.Palette.ToolBlockBg
			c.A = 180
			paint.Fill(gtx.Ops, c)
			return layout.Dimensions{Size: bounds.Max}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left: unit.Dp(12), Right: unit.Dp(12),
				Top: unit.Dp(4), Bottom: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(tb.matTheme, result)
				lbl.Color = tb.theme.Palette.TextSecondary
				lbl.TextSize = tb.theme.Typo.Mono
				return lbl.Layout(gtx)
			})
		}),
	)
}
