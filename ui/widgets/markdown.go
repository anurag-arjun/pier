// Package widgets holds reusable UI components.
package widgets

import (
	"image/color"
	"sync"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/x/markdown"
	"gioui.org/x/richtext"
)

// MarkdownRenderer wraps gioui.org/x/markdown with caching.
type MarkdownRenderer struct {
	renderer *markdown.Renderer
	shaper   *text.Shaper
	color    color.NRGBA

	mu       sync.Mutex
	cache    map[string][]richtext.SpanStyle
}

// NewMarkdownRenderer creates a markdown renderer with theme colors.
func NewMarkdownRenderer(shaper *text.Shaper, textColor, accentColor color.NRGBA) *MarkdownRenderer {
	r := markdown.NewRenderer()
	r.Config = markdown.Config{
		DefaultFont:      font.Font{Typeface: "Inter"},
		MonospaceFont:    font.Font{Typeface: "JetBrains Mono"},
		DefaultSize:      unit.Sp(14),
		H1Size:           unit.Sp(24),
		H2Size:           unit.Sp(20),
		H3Size:           unit.Sp(16),
		H4Size:           unit.Sp(15),
		DefaultColor:     textColor,
		InteractiveColor: accentColor,
	}
	return &MarkdownRenderer{
		renderer: r,
		shaper:   shaper,
		color:    textColor,
		cache:    make(map[string][]richtext.SpanStyle),
	}
}

// Render converts markdown text to richtext spans, with caching.
func (mr *MarkdownRenderer) Render(src string) []richtext.SpanStyle {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if spans, ok := mr.cache[src]; ok {
		return spans
	}

	spans, err := mr.renderer.Render([]byte(src))
	if err != nil {
		// Fallback: plain text
		spans = []richtext.SpanStyle{{
			Content: src,
			Color:   mr.color,
			Size:    unit.Sp(14),
		}}
	}

	mr.cache[src] = spans
	return spans
}

// Layout renders markdown content.
func (mr *MarkdownRenderer) Layout(gtx layout.Context, state *richtext.InteractiveText, src string) layout.Dimensions {
	spans := mr.Render(src)
	if len(spans) == 0 {
		return layout.Dimensions{}
	}
	return richtext.Text(state, mr.shaper, spans...).Layout(gtx)
}

// InvalidateCache clears the render cache (e.g., on theme change).
func (mr *MarkdownRenderer) InvalidateCache() {
	mr.mu.Lock()
	mr.cache = make(map[string][]richtext.SpanStyle)
	mr.mu.Unlock()
}
