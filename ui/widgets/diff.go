package widgets

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	apptheme "github.com/lighto/pier/app"
)

// DiffView renders a unified diff with green/red line coloring.
type DiffView struct {
	theme    apptheme.Theme
	matTheme *material.Theme
	list     widget.List
}

// NewDiffView creates a diff renderer.
func NewDiffView(theme apptheme.Theme, matTheme *material.Theme) *DiffView {
	return &DiffView{
		theme:    theme,
		matTheme: matTheme,
		list:     widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

// DiffLine is a parsed diff line.
type DiffLine struct {
	Type rune   // '+', '-', ' ', '@' (hunk header), or 0 (other)
	Text string
}

// ParseUnifiedDiff parses a unified diff string into typed lines.
func ParseUnifiedDiff(text string) []DiffLine {
	var lines []DiffLine
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			lines = append(lines, DiffLine{Type: ' ', Text: ""})
			continue
		}
		switch line[0] {
		case '+':
			lines = append(lines, DiffLine{Type: '+', Text: line})
		case '-':
			lines = append(lines, DiffLine{Type: '-', Text: line})
		case '@':
			lines = append(lines, DiffLine{Type: '@', Text: line})
		default:
			lines = append(lines, DiffLine{Type: ' ', Text: line})
		}
	}
	return lines
}

// Layout renders diff lines with color coding.
func (dv *DiffView) Layout(gtx layout.Context, text string) layout.Dimensions {
	lines := ParseUnifiedDiff(text)
	if len(lines) == 0 {
		return layout.Dimensions{}
	}

	return material.List(dv.matTheme, &dv.list).Layout(gtx, len(lines), func(gtx layout.Context, i int) layout.Dimensions {
		line := lines[i]
		return dv.layoutLine(gtx, line)
	})
}

func (dv *DiffView) layoutLine(gtx layout.Context, line DiffLine) layout.Dimensions {
	var bg color.NRGBA
	var fg color.NRGBA

	switch line.Type {
	case '+':
		bg = dv.theme.Palette.DiffAddBg
		fg = dv.theme.Palette.Success
	case '-':
		bg = dv.theme.Palette.DiffRemoveBg
		fg = dv.theme.Palette.Error
	case '@':
		bg = dv.theme.Palette.SurfaceAlt
		fg = dv.theme.Palette.Accent
	default:
		bg = color.NRGBA{} // transparent
		fg = dv.theme.Palette.TextSecondary
	}

	return layout.Stack{}.Layout(gtx,
		// Background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if bg.A == 0 {
				return layout.Dimensions{}
			}
			size := gtx.Constraints.Min
			if size.X == 0 {
				size.X = gtx.Constraints.Max.X
			}
			if size.Y == 0 {
				size.Y = gtx.Constraints.Max.Y
			}
			paint.FillShape(gtx.Ops, bg, clip.Rect(image.Rect(0, 0, size.X, size.Y)).Op())
			return layout.Dimensions{Size: size}
		}),
		// Text
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(1), Bottom: unit.Dp(1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(dv.matTheme, dv.theme.Typo.Mono, line.Text)
				lbl.Color = fg
				lbl.Font.Typeface = dv.theme.Typo.MonoFace
				lbl.Font.Weight = font.Normal
				lbl.MaxLines = 1
				return lbl.Layout(gtx)
			})
		}),
	)
}

// IsDiff returns true if the text looks like a unified diff.
func IsDiff(text string) bool {
	for _, line := range strings.SplitN(text, "\n", 20) {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "@@") {
			return true
		}
	}
	return false
}
