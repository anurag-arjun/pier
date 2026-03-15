package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// --- Color utilities ---

// WithAlpha returns the color with a new alpha value.
func WithAlpha(c color.NRGBA, a uint8) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: a}
}

// MulAlpha multiplies the existing alpha by a/255.
func MulAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = uint8(uint32(c.A) * uint32(a) / 0xFF)
	return c
}

// Hovered blends the color toward white (for dark colors) or black (for light colors).
func Hovered(c color.NRGBA) color.NRGBA {
	if c.A == 0 {
		return color.NRGBA{A: 0x44, R: 0x88, G: 0x88, B: 0x88}
	}
	const ratio = 0x20
	target := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: c.A}
	if approxLuminance(c) > 128 {
		target = color.NRGBA{A: c.A}
	}
	return mixColor(target, c, ratio)
}

// Disabled desaturates and reduces alpha for disabled state.
func Disabled(c color.NRGBA) color.NRGBA {
	const r = 80
	lum := approxLuminance(c)
	d := mixColor(c, color.NRGBA{A: c.A, R: lum, G: lum, B: lum}, r)
	return MulAlpha(d, 160)
}

// mixColor blends c1 and c2 weighted by a/256 and (256-a)/256.
func mixColor(c1, c2 color.NRGBA, a uint8) color.NRGBA {
	ai := int(a)
	return color.NRGBA{
		R: byte((int(c1.R)*ai + int(c2.R)*(256-ai)) / 256),
		G: byte((int(c1.G)*ai + int(c2.G)*(256-ai)) / 256),
		B: byte((int(c1.B)*ai + int(c2.B)*(256-ai)) / 256),
		A: byte((int(c1.A)*ai + int(c2.A)*(256-ai)) / 256),
	}
}

func approxLuminance(c color.NRGBA) byte {
	const (
		r = 13933 // 0.2126 * 65536
		g = 46871 // 0.7152 * 65536
		b = 4732  // 0.0722 * 65536
		t = r + g + b
	)
	return byte((r*int(c.R) + g*int(c.G) + b*int(c.B)) / t)
}

// --- Drawing helpers ---

// DrawRect draws a filled rounded rectangle.
func DrawRect(gtx layout.Context, c color.NRGBA, size image.Point, radius int) layout.Dimensions {
	bounds := image.Rect(0, 0, size.X, size.Y)
	paint.FillShape(gtx.Ops, c, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

// DrawBorderedRect draws a rounded rectangle with a border.
func DrawBorderedRect(gtx layout.Context, fill, border color.NRGBA, size image.Point, radius, borderWidth int) layout.Dimensions {
	bounds := image.Rect(0, 0, size.X, size.Y)
	// Outer (border color)
	paint.FillShape(gtx.Ops, border, clip.UniformRRect(bounds, radius).Op(gtx.Ops))
	// Inner (fill color)
	if borderWidth > 0 {
		inner := image.Rect(borderWidth, borderWidth, size.X-borderWidth, size.Y-borderWidth)
		innerRadius := radius - borderWidth
		if innerRadius < 0 {
			innerRadius = 0
		}
		paint.FillShape(gtx.Ops, fill, clip.UniformRRect(inner, innerRadius).Op(gtx.Ops))
	}
	return layout.Dimensions{Size: size}
}

// DrawLeftBorder draws a vertical accent bar on the left edge.
func DrawLeftBorder(gtx layout.Context, c color.NRGBA, height, width int) layout.Dimensions {
	bounds := image.Rect(0, 0, width, height)
	paint.FillShape(gtx.Ops, c, clip.Rect(bounds).Op())
	return layout.Dimensions{Size: image.Pt(width, height)}
}

// FillBackground fills the available area with a color, typically used with layout.Expanded.
func FillBackground(gtx layout.Context, c color.NRGBA) layout.Dimensions {
	size := gtx.Constraints.Min
	if size.X == 0 {
		size.X = gtx.Constraints.Max.X
	}
	if size.Y == 0 {
		size.Y = gtx.Constraints.Max.Y
	}
	paint.FillShape(gtx.Ops, c, clip.Rect(image.Rect(0, 0, size.X, size.Y)).Op())
	return layout.Dimensions{Size: size}
}
