package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// ScrollShadow overlays top/bottom gradient shadows on a scrollable list
// to indicate more content exists outside the viewport.
type ScrollShadow struct {
	Color  color.NRGBA // base color (typically Background)
	Height unit.Dp     // shadow height (default 16dp)
}

// Layout wraps a list layout and draws shadow overlays.
// showTop/showBottom indicate whether the list has content above/below.
func (ss ScrollShadow) Layout(gtx layout.Context, list *widget.List, content layout.Widget) layout.Dimensions {
	h := gtx.Dp(ss.Height)
	if h == 0 {
		h = gtx.Dp(unit.Dp(16))
	}

	return layout.Stack{}.Layout(gtx,
		// Content
		layout.Stacked(content),
		// Top shadow
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if list.Position.First == 0 && list.Position.Offset == 0 {
				return layout.Dimensions{} // at top, no shadow
			}
			return ss.drawGradient(gtx, h, true)
		}),
		// Bottom shadow
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if list.Position.BeforeEnd {
				return ss.drawGradientBottom(gtx, h)
			}
			return layout.Dimensions{} // at bottom, no shadow
		}),
	)
}

func (ss ScrollShadow) drawGradient(gtx layout.Context, h int, top bool) layout.Dimensions {
	width := gtx.Constraints.Max.X
	steps := 8
	stepH := h / steps
	if stepH < 1 {
		stepH = 1
	}

	for i := 0; i < steps; i++ {
		alpha := uint8(255 - (255 * i / steps))
		c := WithAlpha(ss.Color, alpha)
		y := i * stepH
		rect := image.Rect(0, y, width, y+stepH)
		paint.FillShape(gtx.Ops, c, clip.Rect(rect).Op())
	}
	return layout.Dimensions{Size: image.Pt(width, h)}
}

func (ss ScrollShadow) drawGradientBottom(gtx layout.Context, h int) layout.Dimensions {
	width := gtx.Constraints.Max.X
	totalH := gtx.Constraints.Max.Y
	steps := 8
	stepH := h / steps
	if stepH < 1 {
		stepH = 1
	}

	for i := 0; i < steps; i++ {
		alpha := uint8(255 * i / steps)
		c := WithAlpha(ss.Color, alpha)
		y := totalH - h + (i * stepH)
		rect := image.Rect(0, y, width, y+stepH)
		paint.FillShape(gtx.Ops, c, clip.Rect(rect).Op())
	}
	return layout.Dimensions{Size: image.Pt(width, totalH)}
}
