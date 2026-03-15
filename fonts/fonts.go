// Package fonts embeds Inter and JetBrains Mono font files and provides
// a pre-built text.Shaper for use with Gio.
package fonts

import (
	_ "embed"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
)

//go:embed Inter-Regular.ttf
var interRegular []byte

//go:embed Inter-Medium.ttf
var interMedium []byte

//go:embed Inter-SemiBold.ttf
var interSemiBold []byte

//go:embed Inter-Bold.ttf
var interBold []byte

//go:embed JetBrainsMono-Regular.ttf
var jbMonoRegular []byte

//go:embed JetBrainsMono-Bold.ttf
var jbMonoBold []byte

const (
	// TypefaceInter is the typeface name for Inter.
	TypefaceInter = "Inter"
	// TypefaceMono is the typeface name for JetBrains Mono.
	TypefaceMono = "JetBrains Mono"
)

// Collection returns all font faces for registration with text.NewShaper.
func Collection() []text.FontFace {
	return []text.FontFace{
		mustParse(interRegular, font.Font{Typeface: TypefaceInter, Weight: font.Normal}),
		mustParse(interMedium, font.Font{Typeface: TypefaceInter, Weight: font.Medium}),
		mustParse(interSemiBold, font.Font{Typeface: TypefaceInter, Weight: font.SemiBold}),
		mustParse(interBold, font.Font{Typeface: TypefaceInter, Weight: font.Bold}),
		mustParse(jbMonoRegular, font.Font{Typeface: TypefaceMono, Weight: font.Normal}),
		mustParse(jbMonoBold, font.Font{Typeface: TypefaceMono, Weight: font.Bold}),
	}
}

// NewShaper creates a text.Shaper with the embedded font collection.
func NewShaper() *text.Shaper {
	return text.NewShaper(text.WithCollection(Collection()))
}

func mustParse(data []byte, fnt font.Font) text.FontFace {
	faces, err := opentype.Parse(data)
	if err != nil {
		panic("fonts: failed to parse " + string(fnt.Typeface) + ": " + err.Error())
	}
	return text.FontFace{Font: fnt, Face: faces}
}
