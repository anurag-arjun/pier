// Package app holds the top-level application struct and Gio event loop.
package app

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
)

// Palette defines all colors used by the app.
type Palette struct {
	// Core surfaces (layered from dark → light)
	Background    color.NRGBA // deepest background
	Surface       color.NRGBA // cards, sidebar
	SurfaceAlt    color.NRGBA // hover, selected items
	SurfaceRaised color.NRGBA // dropdowns, popovers, modals

	// Text (use TextPrimary/Secondary/Tertiary helpers for opacity-based hierarchy)
	Text          color.NRGBA // base text color (white in dark, black in light)
	TextSecondary color.NRGBA // secondary text
	TextTertiary  color.NRGBA // captions, timestamps
	TextDisabled  color.NRGBA // disabled elements

	// Accent
	Accent      color.NRGBA
	AccentHover color.NRGBA

	// Semantic
	Error   color.NRGBA
	Warning color.NRGBA
	Success color.NRGBA

	// Borders
	Border       color.NRGBA // visible separator
	BorderSubtle color.NRGBA // faint divider

	// Session status badge colors
	StatusThinking   color.NRGBA
	StatusStreaming   color.NRGBA
	StatusWaiting    color.NRGBA
	StatusError      color.NRGBA
	StatusCompacting color.NRGBA
	StatusRetrying   color.NRGBA

	// Priority badge colors
	PriorityP1 color.NRGBA
	PriorityP2 color.NRGBA
	PriorityP3 color.NRGBA

	// Tool block
	ToolBlockBg     color.NRGBA
	ToolBlockBorder color.NRGBA

	// Code blocks (inside markdown)
	CodeBlockBg color.NRGBA

	// Diff
	DiffAddBg    color.NRGBA
	DiffRemoveBg color.NRGBA
}

// Typography defines text sizes and typeface names.
type Typography struct {
	H1        unit.Sp
	H2        unit.Sp
	H3        unit.Sp
	H4        unit.Sp
	Body      unit.Sp
	BodySmall unit.Sp
	Mono      unit.Sp
	MonoSmall unit.Sp
	Caption   unit.Sp

	// Typeface names (registered via fonts package)
	UIFace   font.Typeface
	MonoFace font.Typeface
}

// Spacing defines layout constants on a strict 8px grid (4px half-steps).
type Spacing struct {
	XXS unit.Dp // 4dp  — tight: badge padding, inline gaps
	XS  unit.Dp // 8dp  — base: standard element gap
	S   unit.Dp // 12dp — comfortable: list item padding
	M   unit.Dp // 16dp — section: gap between groups
	L   unit.Dp // 24dp — major: panel padding
	XL  unit.Dp // 32dp — hero: top-level content areas

	SidebarWidth  unit.Dp
	PanelMinWidth unit.Dp
	BorderRadius  unit.Dp
	BadgeRadius   unit.Dp
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
		H1:        unit.Sp(24),
		H2:        unit.Sp(20),
		H3:        unit.Sp(16),
		H4:        unit.Sp(15),
		Body:      unit.Sp(14),
		BodySmall: unit.Sp(13),
		Mono:      unit.Sp(13),
		MonoSmall: unit.Sp(11),
		Caption:   unit.Sp(11),
		UIFace:    "Inter",
		MonoFace:  "JetBrains Mono",
	}

	DefaultSpacing = Spacing{
		XXS:           unit.Dp(4),
		XS:            unit.Dp(8),
		S:             unit.Dp(12),
		M:             unit.Dp(16),
		L:             unit.Dp(24),
		XL:            unit.Dp(32),
		SidebarWidth:  unit.Dp(240),
		PanelMinWidth: unit.Dp(260),
		BorderRadius:  unit.Dp(6),
		BadgeRadius:   unit.Dp(4),
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
			Background:    rgb(248, 248, 250),
			Surface:       rgb(255, 255, 255),
			SurfaceAlt:    rgb(242, 242, 245),
			SurfaceRaised: rgb(255, 255, 255),

			Text:          rgba(28, 28, 30, 230),   // 90%
			TextSecondary: rgba(28, 28, 30, 153),   // 60%
			TextTertiary:  rgba(28, 28, 30, 102),   // 40%
			TextDisabled:  rgba(28, 28, 30, 71),    // 28%

			Accent:      rgb(59, 91, 219),
			AccentHover: rgb(79, 111, 235),

			Error:   rgb(215, 58, 73),
			Warning: rgb(202, 138, 4),
			Success: rgb(22, 163, 74),

			Border:       rgb(216, 218, 224),
			BorderSubtle: rgb(232, 234, 238),

			StatusThinking:   rgb(217, 119, 6),
			StatusStreaming:   rgb(37, 99, 235),
			StatusWaiting:    rgb(22, 163, 74),
			StatusError:      rgb(220, 38, 38),
			StatusCompacting: rgb(147, 51, 234),
			StatusRetrying:   rgb(202, 138, 4),

			PriorityP1: rgb(220, 38, 38),
			PriorityP2: rgb(217, 119, 6),
			PriorityP3: rgb(156, 163, 175),

			ToolBlockBg:     rgb(245, 246, 248),
			ToolBlockBorder: rgb(232, 234, 238),

			CodeBlockBg: rgb(240, 240, 242),

			DiffAddBg:    rgba(22, 163, 74, 30),
			DiffRemoveBg: rgba(220, 38, 38, 30),
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
			Background:    rgb(26, 26, 30),    // #1a1a1e
			Surface:       rgb(36, 36, 40),    // #242428
			SurfaceAlt:    rgb(46, 46, 51),    // #2e2e33
			SurfaceRaised: rgb(56, 56, 62),    // #38383e

			Text:          rgba(244, 244, 245, 230),  // 90%
			TextSecondary: rgba(244, 244, 245, 153),  // 60%
			TextTertiary:  rgba(244, 244, 245, 102),  // 40%
			TextDisabled:  rgba(244, 244, 245, 71),   // 28%

			Accent:      rgb(107, 138, 253),   // #6b8afd
			AccentHover: rgb(130, 156, 253),   // #829cfd

			Error:   rgb(248, 113, 113),
			Warning: rgb(251, 191, 36),
			Success: rgb(74, 222, 128),

			Border:       rgb(58, 58, 64),     // #3a3a40
			BorderSubtle: rgb(46, 46, 51),     // #2e2e33

			StatusThinking:   rgb(251, 191, 36),
			StatusStreaming:   rgb(96, 165, 250),
			StatusWaiting:    rgb(74, 222, 128),
			StatusError:      rgb(248, 113, 113),
			StatusCompacting: rgb(192, 132, 252),
			StatusRetrying:   rgb(250, 204, 21),

			PriorityP1: rgb(248, 113, 113),
			PriorityP2: rgb(251, 191, 36),
			PriorityP3: rgba(244, 244, 245, 102), // TextTertiary

			ToolBlockBg:     rgb(30, 30, 34),   // slightly darker than Background
			ToolBlockBorder: rgb(46, 46, 51),    // BorderSubtle

			CodeBlockBg: rgb(22, 22, 26),       // #16161a — darkest inset

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

// --- Text opacity helpers ---
// These derive text colors from the base Text color with consistent opacity.

// TextPrimary returns the primary text color (90% opacity).
func TextPrimary(p Palette) color.NRGBA { return p.Text }

// TextSecondary returns secondary text color (60% opacity).
func TextSecondary(p Palette) color.NRGBA { return p.TextSecondary }

// TextTertiary returns tertiary text color (40% opacity).
func TextTertiary(p Palette) color.NRGBA { return p.TextTertiary }

// TextDisabledColor returns disabled text color (28% opacity).
func TextDisabledColor(p Palette) color.NRGBA { return p.TextDisabled }
