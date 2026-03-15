package widgets

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// EmptyState renders a centered placeholder message with an optional hint.
type EmptyState struct {
	Message string
	Hint    string // e.g. "Press Ctrl+Shift+N to start"
}

// Layout renders the empty state centered in the available space.
func (es EmptyState) Layout(gtx layout.Context, th *material.Theme, msgColor, hintColor color.NRGBA) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th, unit.Sp(14), es.Message)
				lbl.Color = msgColor
				lbl.Alignment = 1 // center
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if es.Hint == "" {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(12), es.Hint)
					lbl.Color = hintColor
					lbl.Font.Typeface = "JetBrains Mono"
					lbl.Font.Weight = font.Normal
					lbl.Alignment = 1
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}
