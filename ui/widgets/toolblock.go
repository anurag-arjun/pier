package widgets

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
)

// ToolBlock renders a collapsible tool call block with card containment.
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
	hover    HoverState
}

func NewToolBlock(theme apptheme.Theme, matTheme *material.Theme) *ToolBlock {
	return &ToolBlock{theme: theme, matTheme: matTheme}
}

func (tb *ToolBlock) SetExpanded(v bool) { tb.expanded = v }
func (tb *ToolBlock) Expanded() bool     { return tb.expanded }

func (tb *ToolBlock) Layout(gtx layout.Context) layout.Dimensions {
	if tb.toggle.Clicked(gtx) {
		tb.expanded = !tb.expanded
	}
	tb.hover.Update(gtx)

	return layout.Stack{}.Layout(gtx,
		// Card container with border
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(tb.theme.Space.BorderRadius)
			size := gtx.Constraints.Min
			if size.X == 0 {
				size.X = gtx.Constraints.Max.X
			}
			if size.Y == 0 {
				size.Y = gtx.Constraints.Max.Y
			}
			borderColor := tb.theme.Palette.ToolBlockBorder
			if tb.hover.Hovered() && !tb.expanded {
				borderColor = tb.theme.Palette.Border
			}
			// Draw card
			DrawBorderedRect(gtx, tb.theme.Palette.ToolBlockBg, borderColor, size, rr, 1)
			// Error: left red border
			if tb.IsError {
				errBar := image.Rect(0, 0, gtx.Dp(unit.Dp(3)), size.Y)
				paint.FillShape(gtx.Ops, tb.theme.Palette.Error, clip.UniformRRect(errBar, rr).Op(gtx.Ops))
			}
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Header
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &tb.toggle, func(gtx layout.Context) layout.Dimensions {
						return tb.layoutHeader(gtx)
					})
				}),
				// Separator (when expanded)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !tb.expanded {
						return layout.Dimensions{}
					}
					h := gtx.Dp(unit.Dp(1))
					paint.FillShape(gtx.Ops, tb.theme.Palette.ToolBlockBorder, clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, h)).Op())
					return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, h)}
				}),
				// Body
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !tb.expanded {
						return layout.Dimensions{}
					}
					return tb.layoutBody(gtx)
				}),
			)
		}),
	)
}

func (tb *ToolBlock) layoutHeader(gtx layout.Context) layout.Dimensions {
	headerH := gtx.Dp(unit.Dp(36))
	_ = headerH

	return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			// Expand indicator
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				arrow := "▸"
				if tb.expanded {
					arrow = "▾"
				}
				lbl := material.Label(tb.matTheme, tb.theme.Typo.Caption, arrow)
				lbl.Color = tb.theme.Palette.TextTertiary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
			// Tool name
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(tb.matTheme, tb.theme.Typo.BodySmall, tb.ToolName)
				lbl.Color = tb.theme.Palette.Accent
				lbl.Font.Weight = font.Medium
				lbl.Font.Typeface = tb.theme.Typo.MonoFace
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			// Args
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				args := tb.Args
				if len(args) > 80 {
					args = args[:80] + "…"
				}
				lbl := material.Label(tb.matTheme, tb.theme.Typo.Caption, args)
				lbl.Color = tb.theme.Palette.TextTertiary
				lbl.MaxLines = 1
				lbl.Font.Typeface = tb.theme.Typo.MonoFace
				return lbl.Layout(gtx)
			}),
			// Error badge
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !tb.IsError {
					return layout.Dimensions{}
				}
				return tb.layoutErrorBadge(gtx)
			}),
		)
	})
}

func (tb *ToolBlock) layoutErrorBadge(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				rr := gtx.Dp(unit.Dp(3))
				size := gtx.Constraints.Min
				paint.FillShape(gtx.Ops, WithAlpha(tb.theme.Palette.Error, 30), clip.UniformRRect(image.Rect(0, 0, size.X, size.Y), rr).Op(gtx.Ops))
				return layout.Dimensions{Size: size}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(6), Right: unit.Dp(6), Top: unit.Dp(1), Bottom: unit.Dp(1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(tb.matTheme, tb.theme.Typo.MonoSmall, "ERROR")
					lbl.Color = tb.theme.Palette.Error
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (tb *ToolBlock) layoutBody(gtx layout.Context) layout.Dimensions {
	result := tb.Result
	if result == "" {
		result = tb.Partial
	}
	if result == "" {
		return layout.Dimensions{}
	}

	if len(result) > 4000 {
		result = result[:4000] + "\n\n… (truncated)"
	}

	// Slightly darker background for content area
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := gtx.Constraints.Min
			if size.X == 0 {
				size.X = gtx.Constraints.Max.X
			}
			c := darken(tb.theme.Palette.ToolBlockBg, 8)
			paint.FillShape(gtx.Ops, c, clip.Rect(image.Rect(0, 0, size.X, size.Y)).Op())
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12), Top: unit.Dp(8), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(tb.matTheme, tb.theme.Typo.Mono, result)
				lbl.Color = tb.theme.Palette.TextSecondary
				lbl.Font.Typeface = tb.theme.Typo.MonoFace
				return lbl.Layout(gtx)
			})
		}),
	)
}

// darken subtracts n from each RGB channel.
func darken(c color.NRGBA, n uint8) color.NRGBA {
	sub := func(v, s uint8) uint8 {
		if v < s {
			return 0
		}
		return v - s
	}
	return color.NRGBA{R: sub(c.R, n), G: sub(c.G, n), B: sub(c.B, n), A: c.A}
}
