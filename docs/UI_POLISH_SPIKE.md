# UI Polish Research Spike

## What's Wrong Now

The current Pier UI uses bare Gio material widgets with minimal customization. Compared to apps like Slack, Chapar, or VS Code, the issues are:

1. **No visual hierarchy** — everything is the same weight, no breathing room between sections
2. **No hover/focus states** — sidebar items, tool blocks, and buttons don't respond to mouse hover
3. **No custom fonts** — using Gio's default system fallback, which looks inconsistent cross-platform
4. **Flat and lifeless** — no shadows, no depth cues, no subtle gradients or borders that separate regions
5. **Inconsistent spacing** — padding/margins applied ad-hoc per widget, not from a strict grid
6. **No transitions** — state changes are instant (thinking→streaming→waiting), no visual smoothness
7. **Crude prompt bar** — plain editor with a button, no polish on the input container
8. **Tool blocks look like plain text** — no clear visual containment, no hover expand hint
9. **Sidebar has no personality** — workspace items are text-only, no icons, no visual weight distinction
10. **No custom font family** — monospace code blocks use the same font as body text

## Reference: What Polished Gio Apps Do (Chapar)

Chapar (a Gio API testing tool) demonstrates production-quality UI:

- **Custom theme struct** extending `material.Theme` with 25+ semantic color tokens (SideBarBgColor, BorderColorFocused, DropDownMenuBgColor, etc.)
- **Hover states via pointer events** — sidebar items track `pointer.Enter`/`pointer.Leave`, draw overlay with alpha blending
- **Color utility functions** — `MulAlpha`, `Hovered`, `Disabled`, `WithAlpha` for consistent state transitions
- **Custom rect drawing** — `DrawRect` helper for rounded rectangles with arbitrary radius
- **Rich widget library** — custom TextField, FlatButton, Dropdown, TreeView, SplitView, Tab, Divider, Modal layer
- **Embedded fonts** — JetBrains Mono for code, Source Sans Pro for UI text
- **Icon system** — Material Icons font embedded, used throughout sidebar and buttons

## Design System: The 8px Grid

Adopt an 8px base grid (with 4px half-steps) for ALL spacing. This is what Slack, Discord, VS Code, and Figma use.

```
4px   — tight: inline spacing, badge padding
8px   — base: standard element gap
12px  — comfortable: list item padding
16px  — section: gap between groups
24px  — major: panel padding, section separators
32px  — hero: top-level content areas
```

### Current → Proposed

| Token | Current | Proposed | Usage |
|-------|---------|----------|-------|
| PaddingSmall | 6dp | 4dp | Badge padding, tight gaps |
| Padding | 12dp | 12dp | ✓ Keep (1.5× base) |
| PaddingLarge | 20dp | 24dp | Panel padding (3× base) |
| Gap | 8dp | 8dp | ✓ Keep (base) |
| GapSmall | 4dp | 4dp | ✓ Keep (half base) |
| GapLarge | 16dp | 16dp | ✓ Keep (2× base) |
| SidebarWidth | 220dp | 240dp | Wider for readability |
| PanelMinWidth | 280dp | 260dp | Slightly tighter |

## Typography

### Embed Proper Fonts

Bundle two font families:

1. **Inter** (or Source Sans Pro) — UI text. Clean, neutral, excellent at small sizes. Open source.
2. **JetBrains Mono** — Code/monospace. Ligatures, clear at 12-13px. Open source.

Chapar embeds fonts via `go:embed` and registers them with `text.NewShaper(text.WithCollection(faces))`.

### Scale (adjusted)

| Role | Size | Weight | Font |
|------|------|--------|------|
| H1 | 24sp | Bold | Inter |
| H2 | 20sp | Semibold | Inter |
| H3 | 16sp | Semibold | Inter |
| Body | 14sp | Regular | Inter |
| Body Small | 13sp | Regular | Inter |
| Caption | 11sp | Regular | Inter |
| Mono | 13sp | Regular | JetBrains Mono |
| Mono Small | 11sp | Regular | JetBrains Mono |

### Text Opacity Hierarchy (Dark Theme)

Instead of multiple gray shades, use opacity on white text:

| Role | Opacity | Current approach |
|------|---------|-----------------|
| Primary | 90% | Separate color token |
| Secondary | 60% | Separate color token |
| Tertiary | 40% | Separate color token |
| Disabled | 28% | Not implemented |

This is cleaner — one text color with alpha variations works better across theme switches.

## Color Refinement

### Dark Theme Palette (revised)

```
Background:     #1a1a1e  (not pure black — warmer dark gray)
Surface:        #242428  (cards, sidebar bg — 1 step lighter)
Surface Alt:    #2e2e33  (hover state, selected items)
Surface Raised: #38383e  (dropdowns, popovers — 1 more step)
Border:         #3a3a40  (subtle separator)
Border Subtle:  #2e2e33  (very faint divider)
Accent:         #6b8afd  (calmer blue — less saturated than current)
Accent Hover:   #829cfd  (lighter on hover)
Error:          #f87171
Warning:        #fbbf24
Success:        #4ade80
```

The current dark theme uses `rgb(24,24,27)` for background — close but slightly too blue. Shift toward warmer neutral.

### Light Theme

Keep but refine. The current light theme is fine conceptually — just needs the same opacity/grid treatment.

## Component Improvements

### 1. Sidebar

**Current:** Plain text list with colored dots.

**Target (Slack-like):**
- Each workspace item: 32dp height, 8dp left padding, 4dp rounded corner on hover
- Hover: background fills with `Surface Alt` color (alpha overlay)
- Selected: left 3px accent bar + `Surface Alt` background
- Workspace name: Body weight 500, path below in Caption at 40% opacity
- Session cards: indent 16dp under workspace, show model as pill badge
- Section headers ("WORKSPACES"): 11sp, 40% opacity, 16dp bottom margin, uppercase tracking +0.5px

### 2. Timeline

**Current:** Plain text messages, basic tool headers.

**Target:**
- **Message groups:** Same-sender messages within 5 min get 4dp gap; different senders get 16dp gap
- **User messages:** No background, just text with "You" label
- **Assistant messages:** Subtle left border (2dp accent color, 8dp left padding) to distinguish from user
- **Tool blocks:** Proper card containment — Surface background, 1px border, 6dp radius. Header has subtle bottom border when expanded. Hover on collapsed header shows "Click to expand" cursor
- **Code blocks in markdown:** Dark inset background (`#1e1e22`), monospace JetBrains Mono, 4dp padding, copy button on hover
- **Timestamps:** Right-aligned, Caption size, 30% opacity, shown per message group

### 3. Prompt Bar

**Current:** Editor with border and send button.

**Target (Slack-like):**
- Outer container: 1px border `Border` color, 8dp radius, Surface background
- Focus state: border becomes `Accent` color
- Minimum height: 40dp, grows with multiline
- Placeholder: "Message pi..." in 40% opacity
- Send button: only visible when text is non-empty, accent circle with arrow icon
- Left side: optional `/` command trigger indicator
- Above the bar: thin status line — "● Thinking..." / "● Streaming..." with pulsing dot animation

### 4. Tool Blocks

**Current:** ▸/▾ text toggle with background rect.

**Target:**
- Container: `Surface` background, 1px `Border Subtle` border, 6dp radius
- Header: 36dp height, `[bash]` in monospace accent, args in secondary color, right side shows duration ("1.2s")
- Collapsed: single line, hover brightens background
- Expanded: header gets bottom 1px border, content area gets slightly darker background
- Content: JetBrains Mono, 13sp, line-height 1.5, max-height 300dp with scroll, "Show all" button when truncated
- Error state: left 3px red border

### 5. Task Panel

**Current:** Basic text cards.

**Target:**
- Task cards: `Surface` background, 1px border, 8dp radius, 12dp padding
- Priority: colored dot (not text badge) — P1=red, P2=amber, P3=gray
- Title: Body weight 500
- ID: Caption monospace, 40% opacity
- Hover: border becomes `Border` (slightly brighter)
- Ready tasks: full opacity. Blocked: 50% opacity with "blocked by X" caption

### 6. Status Dots & Badges

**Current:** Static colored ellipses.

**Target:**
- Thinking: pulsing amber dot (opacity animates 50%→100%→50%)
- Streaming: solid blue with subtle glow
- Waiting: solid green
- Error: solid red with subtle pulse
- Model badge: pill shape (24dp height, 8dp horizontal padding, 4dp radius), Surface Alt background, monospace text

## Implementation Plan

### Phase A: Foundation (do first)
1. **Embed Inter + JetBrains Mono fonts** — register with Gio's text shaper
2. **Add color utility functions** — `WithAlpha`, `MulAlpha`, `Hovered`, `Disabled` (steal from Chapar)
3. **Add `DrawRect` helper** — rounded rect with fill, border, optional shadow
4. **Add hover tracking** — reusable `HoverState` widget using `pointer.Enter`/`pointer.Leave`
5. **Revise theme tokens** — add Surface Raised, Accent Hover, text opacity helpers
6. **Strict 8px grid** — audit and fix all spacing to grid

### Phase B: Core Components
7. **Restyle sidebar** — hover states, selected indicator bar, proper section headers
8. **Restyle prompt bar** — focus border, growing height, status line above
9. **Restyle tool blocks** — card containment, hover, border states, duration
10. **Restyle timeline messages** — assistant left border, message grouping, timestamps

### Phase C: Refinement
11. **Restyle task panel** — card borders, priority dots, hover states
12. **Restyle extension dialogs** — proper modal overlay with backdrop blur/dim
13. **Add dot animation** — pulsing thinking dot in sidebar and status bar
14. **Code block styling** — dark inset background, copy button, line numbers

### Phase D: Polish
15. **Scroll shadows** — subtle gradient at top/bottom of scrollable areas when content overflows
16. **Smooth transitions** — fade opacity on status changes (requires Gio animation ops)
17. **Empty states** — proper illustrations/messages when no sessions, no tasks, no plan
18. **Loading states** — skeleton placeholders while br/pi data loads

## Effort Estimate

| Phase | Tasks | Estimate |
|-------|-------|----------|
| A: Foundation | 6 | 2-3 hours |
| B: Core Components | 4 | 3-4 hours |
| C: Refinement | 4 | 2-3 hours |
| D: Polish | 4 | 2-3 hours |
| **Total** | **18** | **~10-12 hours** |

## Key Gio Techniques

### Hover Detection (from Chapar)
```go
// Register pointer events
defer pointer.PassOp{}.Push(gtx.Ops).Pop()
defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
event.Op(gtx.Ops, tag)

// Drain events
for {
    ev, ok := gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Enter | pointer.Leave})
    if !ok { break }
    if pe, ok := ev.(pointer.Event); ok {
        switch pe.Kind {
        case pointer.Enter: hovering = true
        case pointer.Leave: hovering = false
        }
    }
}
```

### Rounded Rect with Border
```go
func DrawBorderedRect(gtx C, fill, border color.NRGBA, size image.Point, radius, borderWidth int) D {
    // Outer (border)
    paint.FillShape(gtx.Ops, border, clip.UniformRRect(image.Rect(0,0,size.X,size.Y), radius).Op(gtx.Ops))
    // Inner (fill) inset by borderWidth
    inner := image.Rect(borderWidth, borderWidth, size.X-borderWidth, size.Y-borderWidth)
    paint.FillShape(gtx.Ops, fill, clip.UniformRRect(inner, radius-borderWidth).Op(gtx.Ops))
    return D{Size: size}
}
```

### Animation (pulsing dot)
```go
// In layout:
t := float32(gtx.Now.UnixMilli()%1000) / 1000.0
alpha := uint8(128 + 127*math.Sin(float64(t)*2*math.Pi))
dotColor := WithAlpha(theme.StatusThinking, alpha)
// Invalidate to keep animating
op.InvalidateCmd{}.Add(gtx.Ops)
```

### Embedded Fonts
```go
//go:embed fonts/Inter-Regular.ttf
var interRegular []byte

//go:embed fonts/JetBrainsMono-Regular.ttf  
var jetbrainsMonoRegular []byte

faces := []text.FontFace{
    {Font: font.Font{Typeface: "Inter"}, Face: opentype.Parse(interRegular)},
    {Font: font.Font{Typeface: "JetBrains Mono"}, Face: opentype.Parse(jetbrainsMonoRegular)},
}
shaper := text.NewShaper(text.WithCollection(faces))
```
