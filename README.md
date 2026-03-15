# Pier

A native GUI for [pi](https://pi.dev), the open-source terminal coding harness. Pier connects to pi's [RPC mode](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent/docs/rpc.md) — a JSON protocol designed for exactly this kind of embedding — and renders structured agent data that's invisible in a raw terminal.

Multiple concurrent sessions, the full discover → plan → execute workflow, and [br](https://github.com/Dicklesworthstone/beads_rust) task tracking, all in one window.

**Go + Gio · Single binary · ~11MB · Wayland/X11/Metal**

> Pier is **not** a replacement for pi. Pi does all the LLM work — tools, context management, extensions, compaction, model switching. Pier is a display layer that reads pi's event stream and provides a visual interface for managing sessions.

## Why a GUI

Pi is deliberately terminal-first and extensible. That's the right default. But some things are easier to see than read:

- Which tool is executing right now, with what arguments
- Diffs from write/edit operations, color-coded
- Agent status at a glance across multiple sessions
- Task progress without switching terminals
- Extension dialogs (confirmations, selections) as native widgets instead of TUI prompts

Pier reads pi's structured RPC event stream and renders all of this natively — no webview, no Electron, no GTK.

## Features

- **Multiple concurrent sessions** — one per task or feature branch, each with its own timeline and prompt bar
- **Structured timeline** — messages render as markdown (Inter + JetBrains Mono), tool calls are collapsible bordered cards with hover feedback
- **Live status** — sidebar shows thinking/streaming/waiting/error per session with pulsing animated status dots
- **Task panel** — displays [br](https://github.com/Dicklesworthstone/beads_rust) tasks grouped by ready/in-progress/blocked with priority dots and hover states
- **Discovery mode** — tool-free chat (pi with `--no-tools`) for thinking through the problem before writing code
- **Plan panel** — renders plan.md as markdown, one click to open in your editor
- **Slash command autocomplete** — type `/` in the prompt bar to see all pi commands in a bordered popup
- **Extension UI** — renders pi extension dialogs (select, confirm, input) as centered modal cards with dimmed backdrop
- **Diff rendering** — green/red line coloring for unified diffs in tool output
- **Pi keybindings** — Escape aborts, Ctrl+P cycles models, Shift+Tab cycles thinking level — same keys as pi's terminal UI, forwarded as RPC commands

## How It Works

Pier spawns pi as a child process in RPC mode (`pi --mode rpc`). Pi emits newline-delimited JSON events on stdout (agent lifecycle, streaming text deltas, tool executions, compaction, extension UI requests). Pier decodes the stream, updates a state machine, and renders the result with Gio's immediate-mode GPU-accelerated UI.

Commands go the other way: when you type a prompt or press Ctrl+P, Pier writes a JSON command to pi's stdin. Pi handles it and emits the resulting events.

Pi manages everything: the LLM, tools, sessions, extensions, compaction, auth. Pier manages the display.

## Requirements

- **Go 1.21+** (build dependency)
- **C compiler** — `gcc` or `clang` (Gio CGo requirement, present by default on most Linux distros)
- **[pi](https://pi.dev)** — installed and authenticated (`pi /login`)
- **[br](https://github.com/Dicklesworthstone/beads_rust)** (optional) — for task panel functionality

## Install

```bash
git clone https://github.com/anurag-arjun/pier.git
cd pier
go build -ldflags="-s -w" -o pier .
```

Or just run directly:

```bash
go run .
```

No install step needed. Copy the binary wherever you want.

### Cross-compile for macOS

```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o pier-mac .
```

## Usage

```bash
cd ~/your-project
pier
```

Pier opens in the current directory. Click **Start Pi Session** to spawn pi in RPC mode. Type a prompt, hit Enter, watch the timeline.

See [GUIDE.md](GUIDE.md) for the full user guide.

## Keyboard Shortcuts

### Pi passthrough (sent as RPC commands)

| Key | Action |
|-----|--------|
| Escape | Abort current agent run |
| Ctrl+P | Cycle to next model |
| Ctrl+Shift+P | Cycle model backward |
| Shift+Tab | Cycle thinking level |
| Ctrl+L | Select model |

### Session lifecycle

| Key | Action |
|-----|--------|
| Ctrl+C | Clear prompt bar |
| Ctrl+C ×2 | Kill pi session (same as pi) |
| Ctrl+W | Close session |
| Ctrl+Shift+N | New session |

### Pier UI

| Key | Action |
|-----|--------|
| Ctrl+O | Collapse/expand tool output |
| Ctrl+B | Toggle task panel |
| Ctrl+R | Refresh task list |
| Ctrl+/ | Focus prompt bar |
| Ctrl+Shift+C | Copy last assistant message |
| Ctrl+1–9 | Switch workspace |
| Ctrl+Tab | Cycle workspaces |

## Design

### Visual Design

Pier follows a Slack-inspired design language:

- **8px grid** — all spacing derived from 4/8/12/16/24/32dp increments
- **4-level surface stack** — Background → Surface → SurfaceAlt → SurfaceRaised for visual depth
- **Opacity-based text hierarchy** — 90% primary, 60% secondary, 40% tertiary, 28% disabled
- **Custom fonts** — [Inter](https://rsms.me/inter/) for UI text (4 weights), [JetBrains Mono](https://www.jetbrains.com/lp/mono/) for code (2 weights)
- **Hover states** — sidebar items, tool blocks, task cards, autocomplete options all respond to mouse hover
- **Accent indicators** — left border on assistant messages, selected workspace accent bar, error accent bars
- **Pulsing status dots** — sine wave animation for thinking/streaming states

### Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design document covering:

- RPC protocol integration (all event/command types)
- Rendering performance strategy (streaming vs. complete mode, lazy layout)
- Workspace model and persistence
- Session lifecycle and state machine

## Project Structure

```
pier/
├── main.go                    app wiring, event loop, key dispatch
├── app/
│   ├── theme.go               4-level dark/light palettes, 8px grid, typography
│   └── keymap.go              pi passthrough + Pier UI action dispatch
├── fonts/
│   ├── fonts.go               go:embed Inter + JetBrains Mono, NewShaper()
│   ├── Inter-*.ttf            Inter Regular/Medium/SemiBold/Bold
│   └── JetBrainsMono-*.ttf    JetBrains Mono Regular/Bold
├── session/
│   ├── events.go              all RPC event types (agent, tool, compaction, extension)
│   ├── commands.go            all RPC command types + ExtensionUIResponse
│   ├── decoder.go             JSONL decoder (\n-only split, not Unicode separators)
│   ├── state.go               session state machine, timeline model
│   ├── process.go             pi process spawn/pipe/stop
│   └── session.go             session orchestration (DrainEvents, SendPrompt)
├── ui/
│   ├── sidebar.go             workspace list, hover states, accent bar, pill badges
│   ├── timeline.go            left-bordered messages, grouped spacing, model badges
│   ├── promptbar.go           bordered input, status line, slash autocomplete
│   ├── taskpanel.go           bordered cards, priority dots, hover, error accent
│   ├── planpanel.go           plan.md markdown rendering, Open in $EDITOR
│   ├── discover_view.go       tool-free chat, Generate Plan handoff
│   ├── extensionui.go         dimmed backdrop modal, select/confirm/input dialogs
│   ├── workspace_view.go      multi-session tabs, task linking
│   └── widgets/
│       ├── draw.go            WithAlpha, MulAlpha, Hovered, DrawRect, DrawBorderedRect
│       ├── hover.go           HoverState (pointer enter/leave tracking)
│       ├── toolblock.go       bordered collapsible card with error accent
│       ├── markdown.go        cached markdown renderer (Inter + JetBrains Mono)
│       ├── diff.go            green/red unified diff rendering
│       ├── animated_dot.go    pulsing sine wave status dot
│       ├── empty_state.go     centered placeholder messages
│       └── scroll_shadow.go   top/bottom gradient overflow indicators
├── br/                        br CLI wrapper (List, Ready, Show, Close)
├── config/                    AppConfig load/save (~/.config/pier/)
├── workspace/                 workspace persistence (discovery, plan, sessions)
└── resources/                 embedded prompts via go:embed
```

## Testing

```bash
go test ./... -timeout 30s
```

29 tests across `session/` and `ui/widgets/`:
- JSONL decoder: lifecycle, tools, streaming, Unicode, malformed lines (9 tests)
- State machine: transitions, tool tracking, compaction, retry, model updates (9 tests)
- Stress: concurrent decoders, 100-entry timeline, concurrent state access (3 tests)
- Color utilities: WithAlpha, MulAlpha, Hovered, Disabled, luminance (8 tests)

## License

MIT
