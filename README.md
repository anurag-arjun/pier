# Pier

A native GUI for the [pi](https://pi.dev) coding agent. Manages multiple concurrent pi sessions, surfaces structured data invisible in a raw terminal, and covers the full workflow — discover, plan, create tasks, execute — in one tool.

**Go + Gio · Single binary · ~9MB · Wayland/X11/Metal**

## Why

Running pi in a terminal works, but you can't see:
- Which tool is executing and what it's doing
- When the agent is thinking vs. streaming vs. waiting
- File diffs from write/edit operations at a glance
- Multiple sessions side by side
- Task progress without switching to another terminal

Pier reads pi's structured RPC event stream and renders all of this natively — no webview, no Electron, no GTK.

## Features

- **Multiple concurrent sessions** — one per task or feature branch, each with its own timeline and prompt bar
- **Structured timeline** — messages render as markdown, tool calls are collapsible blocks with `[read]`, `[write]`, `[bash]` headers
- **Live status** — sidebar shows thinking/streaming/waiting/error per session with colored status dots
- **Task panel** — displays [br](https://github.com/Dicklesworthstone/beads_rust) tasks grouped by ready/in-progress/blocked, start sessions linked to tasks
- **Discovery mode** — tool-free chat for thinking through the problem before writing code
- **Plan panel** — renders plan.md as markdown, one click to open in your editor
- **Slash command autocomplete** — type `/` in the prompt bar to see all pi commands
- **Extension UI** — renders pi extension dialogs (select, confirm, input) natively
- **Pi keybindings** — Escape aborts, Ctrl+P cycles models, Shift+Tab cycles thinking level — same as pi's terminal UI, forwarded as RPC commands

## Requirements

- **Go 1.21+** (build dependency)
- **C compiler** — `gcc` or `clang` (Gio CGo requirement, present by default on most Linux distros)
- **pi** — installed and authenticated (`pi /login`)
- **br** (optional) — for task panel functionality

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

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design document covering:

- RPC protocol integration (all event/command types)
- Rendering performance strategy (streaming vs. complete mode, lazy layout)
- Workspace model and persistence
- Session lifecycle and state machine
- Build phases and rationale

## Project Structure

```
pier/
├── main.go                    app wiring, event loop, key dispatch
├── app/
│   ├── theme.go               light/dark palettes, typography, spacing
│   └── keymap.go              shortcut definitions and dispatch
├── session/
│   ├── events.go              RPC event type definitions
│   ├── commands.go            RPC command type definitions
│   ├── decoder.go             JSONL stream decoder (\n-only split)
│   ├── state.go               session state machine, timeline model
│   ├── process.go             pi process spawn/pipe/stop
│   └── session.go             session orchestration
├── ui/
│   ├── sidebar.go             workspace list, session cards
│   ├── timeline.go            message + tool call rendering
│   ├── promptbar.go           input with slash autocomplete
│   ├── taskpanel.go           br task display
│   ├── planpanel.go           plan.md markdown rendering
│   ├── discover_view.go       tool-free chat
│   ├── extensionui.go         extension dialog rendering
│   ├── workspace_view.go      multi-session management
│   └── widgets/
│       ├── markdown.go        cached markdown renderer
│       └── toolblock.go       collapsible tool block
├── br/                        br CLI wrapper
├── config/                    app config load/save
├── workspace/                 workspace persistence
└── resources/                 embedded prompts (go:embed)
```

## Status

Phase 1 (core loop) and Phase 2 (full UI) are complete. The app builds, runs, and connects to pi in RPC mode. Remaining:

- [ ] Stress test with 5+ concurrent sessions
- [ ] Custom diff widget for write/edit tool results

## License

MIT
