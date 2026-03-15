package widgets

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// AnimatedDot renders a pulsing status dot.
// When animate is true, the dot's alpha oscillates on a sine wave.
type AnimatedDot struct{}

// Layout renders the dot. If animate is true, it pulses and invalidates for continuous animation.
func (AnimatedDot) Layout(gtx layout.Context, c color.NRGBA, size unit.Dp, animate bool) layout.Dimensions {
	sz := gtx.Dp(size)

	if animate {
		// Sine wave: 1.5s period, alpha 100→255→100
		ms := gtx.Now.UnixMilli() % 1500
		t := float64(ms) / 1500.0
		alpha := uint8(128 + 127*math.Sin(t*2*math.Pi))
		c = WithAlpha(c, alpha)
		// Keep animating
		gtx.Execute(op.InvalidateCmd{})
	}

	defer clip.Ellipse{Max: image.Pt(sz, sz)}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, c)
	return layout.Dimensions{Size: image.Pt(sz, sz)}
}
