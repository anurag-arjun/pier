// Package app holds the top-level application struct and Gio event loop.
package app

import (
	"image/color"

	"gioui.org/unit"
)

// Palette defines all colors used by the app.
type Palette struct {
	Background    color.NRGBA
	Surface       color.NRGBA
	SurfaceAlt    color.NRGBA
	Text          color.NRGBA
	TextSecondary color.NRGBA
	Accent        color.NRGBA
	Error         color.NRGBA
	Border        color.NRGBA
	BorderSubtle  color.NRGBA

	// Session status badge colors
	StatusThinking  color.NRGBA // amber
	StatusStreaming  color.NRGBA // blue
	StatusWaiting   color.NRGBA // green
	StatusError     color.NRGBA // red
	StatusCompacting color.NRGBA // purple
	StatusRetrying  color.NRGBA // yellow

	// Priority badge colors
	PriorityP1 color.NRGBA
	PriorityP2 color.NRGBA
	PriorityP3 color.NRGBA

	// Tool block
	ToolBlockBg    color.NRGBA
	ToolBlockBorder color.NRGBA

	// Diff
	DiffAddBg    color.NRGBA
	DiffRemoveBg color.NRGBA
}

// Typography defines text sizes.
type Typography struct {
	H1       unit.Sp
	H2       unit.Sp
	H3       unit.Sp
	H4       unit.Sp
	Body     unit.Sp
	BodySmall unit.Sp
	Mono     unit.Sp
	Caption  unit.Sp
}

// Spacing defines layout constants.
type Spacing struct {
	Padding     unit.Dp
	PaddingSmall unit.Dp
	PaddingLarge unit.Dp
	Gap         unit.Dp
	GapSmall    unit.Dp
	GapLarge    unit.Dp
	SidebarWidth unit.Dp
	PanelMinWidth unit.Dp
	BorderRadius unit.Dp
	BadgeRadius  unit.Dp
}

// Theme holds the full visual theme.
type Theme struct {
	Name    string // "light" | "dark"
	Palette Palette
	Typo    Typography
	Space   Spacing
}

// Common typography and spacing (shared between themes).
var (
	DefaultTypography = Typography{
		H1:        unit.Sp(28),
		H2:        unit.Sp(24),
		H3:        unit.Sp(20),
		H4:        unit.Sp(17),
		Body:      unit.Sp(14),
		BodySmall: unit.Sp(12),
		Mono:      unit.Sp(13),
		Caption:   unit.Sp(11),
	}

	DefaultSpacing = Spacing{
		Padding:      unit.Dp(12),
		PaddingSmall: unit.Dp(6),
		PaddingLarge: unit.Dp(20),
		Gap:          unit.Dp(8),
		GapSmall:     unit.Dp(4),
		GapLarge:     unit.Dp(16),
		SidebarWidth: unit.Dp(220),
		PanelMinWidth: unit.Dp(280),
		BorderRadius: unit.Dp(6),
		BadgeRadius:  unit.Dp(3),
	}
)

func rgb(r, g, b uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: 0xFF}
}

func rgba(r, g, b, a uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// LightTheme returns the light palette.
func LightTheme() Theme {
	return Theme{
		Name: "light",
		Palette: Palette{
			Background:    rgb(250, 250, 250),
			Surface:       rgb(255, 255, 255),
			SurfaceAlt:    rgb(245, 245, 245),
			Text:          rgb(28, 28, 30),
			TextSecondary: rgb(120, 120, 128),
			Accent:        rgb(0, 122, 255),
			Error:         rgb(215, 58, 73),
			Border:        rgb(209, 213, 219),
			BorderSubtle:  rgb(229, 231, 235),

			StatusThinking:  rgb(245, 158, 11),
			StatusStreaming:  rgb(59, 130, 246),
			StatusWaiting:   rgb(34, 197, 94),
			StatusError:     rgb(239, 68, 68),
			StatusCompacting: rgb(168, 85, 247),
			StatusRetrying:  rgb(234, 179, 8),

			PriorityP1: rgb(239, 68, 68),
			PriorityP2: rgb(245, 158, 11),
			PriorityP3: rgb(107, 114, 128),

			ToolBlockBg:     rgb(248, 249, 250),
			ToolBlockBorder: rgb(229, 231, 235),

			DiffAddBg:    rgba(34, 197, 94, 30),
			DiffRemoveBg: rgba(239, 68, 68, 30),
		},
		Typo:  DefaultTypography,
		Space: DefaultSpacing,
	}
}

// DarkTheme returns the dark palette.
func DarkTheme() Theme {
	return Theme{
		Name: "dark",
		Palette: Palette{
			Background:    rgb(24, 24, 27),
			Surface:       rgb(39, 39, 42),
			SurfaceAlt:    rgb(49, 49, 53),
			Text:          rgb(244, 244, 245),
			TextSecondary: rgb(161, 161, 170),
			Accent:        rgb(96, 165, 250),
			Error:         rgb(248, 113, 113),
			Border:        rgb(63, 63, 70),
			BorderSubtle:  rgb(52, 52, 56),

			StatusThinking:  rgb(251, 191, 36),
			StatusStreaming:  rgb(96, 165, 250),
			StatusWaiting:   rgb(74, 222, 128),
			StatusError:     rgb(248, 113, 113),
			StatusCompacting: rgb(192, 132, 252),
			StatusRetrying:  rgb(250, 204, 21),

			PriorityP1: rgb(248, 113, 113),
			PriorityP2: rgb(251, 191, 36),
			PriorityP3: rgb(161, 161, 170),

			ToolBlockBg:     rgb(30, 30, 33),
			ToolBlockBorder: rgb(52, 52, 56),

			DiffAddBg:    rgba(74, 222, 128, 25),
			DiffRemoveBg: rgba(248, 113, 113, 25),
		},
		Typo:  DefaultTypography,
		Space: DefaultSpacing,
	}
}

// ThemeByName returns a theme by name ("light" or "dark").
func ThemeByName(name string) Theme {
	if name == "dark" {
		return DarkTheme()
	}
	return LightTheme()
}
