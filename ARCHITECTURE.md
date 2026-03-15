# Pier — Architecture Document

A lightweight, native GUI for [pi](https://pi.dev), the open-source terminal coding harness. Connects to pi's [RPC mode](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent/docs/rpc.md) (JSON protocol over stdin/stdout) to render structured agent data invisible in a raw terminal. Manages multiple concurrent pi sessions, covers the complete workflow from discovery through task execution, and integrates with [br](https://github.com/Dicklesworthstone/beads_rust) for task tracking — all in one tool.

Pier is a display layer. Pi does all the LLM work (tools, context, extensions, compaction, model switching). Pier reads the event stream and provides the visual interface.

Stack: Go + Gio · Single binary (~11MB) · Wayland/X11 (CachyOS) · Mac-compatible (Metal) via `GOOS=darwin`

## Table of Contents

1. [Goals and Non-Goals](#1-goals-and-non-goals)
2. [Core Concepts](#2-core-concepts)
3. [Stack Rationale](#3-stack-rationale)
4. [Workspace Model](#4-workspace-model)
5. [Discover View](#5-discover-view)
6. [Planning Workflow](#6-planning-workflow)
7. [Execute View](#7-execute-view)
8. [Pi Integration Layer](#8-pi-integration-layer)
9. [br Integration Layer](#9-br-integration-layer)
10. [Bundled Pi Resources](#10-bundled-pi-resources)
11. [Application Structure](#11-application-structure)
12. [UI Layout](#12-ui-layout)
13. [Session Lifecycle](#13-session-lifecycle)
14. [Persistence Model](#14-persistence-model)
15. [Keyboard Shortcuts](#15-keyboard-shortcuts)
16. [Configuration](#16-configuration)
17. [Visual Design](#17-visual-design)
18. [Rendering Performance](#18-rendering-performance)
19. [Authentication](#19-authentication)
20. [Cross-Platform Notes](#20-cross-platform-notes)
21. [Build and Distribution](#21-build-and-distribution)
22. [Build Order](#22-build-order)
23. [What You Get Over the Current Workflow](#23-what-you-get-over-the-current-workflow)

---

## 1. Goals and Non-Goals

### Goals

- Expose the full power of pi's tool harness in a structured visual interface
- Cover the complete workflow in one tool: discover → plan → create tasks → execute
- Multiple pi sessions running concurrently, each fully independent
- Surface structured data (model, agent state, tool calls, task progress) invisible in a raw terminal
- Absolute minimum resource footprint — no webview, no Electron, no GTK
- Single static binary, zero install friction

### Non-Goals

- Not a terminal emulator
- Not an agent orchestrator that spawns sub-agents autonomously
- Not a replacement for pi — pi does all the LLM harness work
- Not a beads/br replacement — br is the source of truth for tasks
- Not a cost tracker — you are on Max

---

## 2. Core Concepts

### Workspace

The central unit. A workspace maps to a project directory. It holds a discovery conversation, a plan, br tasks, and one or more coding sessions. There is no enforced stage progression — all parts of the workspace are accessible at any time. The current view is a display hint, not a gate.

### Session

A single running pi process within a workspace. A workspace can have multiple concurrent sessions (e.g., one per feature branch or task). Each session has its own event stream, prompt bar, and timeline.

### Task

A br issue (`br` = beads_rust CLI). Tasks live in `.beads/` inside the workspace. The app reads task state by shelling out to `br --json`. Pi creates and updates tasks; the app displays them.

---

## 3. Stack Rationale

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Language | Go | Goroutines map perfectly to the IO-bound concurrency model. One goroutine per pi session reading the RPC stream is idiomatic and cheap. No borrow checker friction across async boundaries. |
| GUI | Gio (`gioui.org`) | Immediate-mode, GPU-accelerated via OpenGL/Vulkan. Speaks directly to Wayland — no GTK, no Qt, no system UI framework. On macOS uses Metal. Same code, zero changes. |
| Markdown | `gioui.org/x/markdown` + `goldmark` | Official Gio extension. Parses markdown into `richtext.SpanStyle` slices for native rendering. goldmark (CommonMark compliant) available for custom AST walking if needed. |
| Pi integration | `os/exec` + JSON stream decoder | Stdlib only. Pi's RPC mode emits newline-delimited JSON on stdout. One `bufio.Scanner` per session goroutine. |
| br integration | `os/exec` + `br --json` | Shell out to the br binary. No direct SQLite dependency. On-demand only, not polled. |
| Persistence | `encoding/json` | Single config file + per-workspace state file. No database. |
| Diffing | Custom (`ui/widgets/diff.go`) | Unified diff parser with green/red line rendering. No external dependency — parses +/-/@ prefixes directly. |
| Fonts | Inter + JetBrains Mono (`go:embed`) | Inter for UI text (4 weights), JetBrains Mono for code (2 weights). Embedded in binary via `go:embed`. OFL licensed. |
| Binary | `go build` | Single static binary (~11MB). No runtime deps beyond libc. |

**Why not Rust + egui:** Rust's async ergonomics for process management (`Arc<Mutex<>>` across async boundaries) adds friction for a primarily IO-bound app. Go's goroutines handle this more naturally.

**Why not Tauri:** WebKitGTK on Linux introduces a webview subprocess, rendering quirks, and memory overhead. Gio speaks directly to the compositor.

---

## 4. Workspace Model

```
Workspace
├── id:               string
├── name:             string           (user-editable)
├── path:             string           (filesystem path)
├── discovery:
│   ├── messages:     []Message        (full conversation, persisted)
│   └── model:        string
├── plan:
│   ├── path:         string           (plan.md location, default <workspace>/plan.md)
│   └── br_created:   bool
├── br_initialised:   bool
└── sessions:         []Session
```

Persisted to `~/.config/pier/workspaces/<id>.json` after every meaningful state change.

There is no stage machine. The workspace view shows the discover panel, plan panel, and task panel simultaneously. Pi sessions run alongside all of these at all times. Nothing is gated.

---

## 5. Discover View

A focused chat interface for thinking through the problem before committing to a plan. This is where you work out what you actually want to build — the shape of the solution, the constraints, the approach.

### What it is

- A clean message thread: user and assistant turns
- Pi is spawned with `--no-tools --no-extensions --no-skills --mode rpc` and a planning-focused system prompt
- Streaming responses via pi's existing `message_update` RPC events with `text_delta` — no new event types needed
- Model selector: any model available through your pi Max subscription
- The full conversation is persisted in the workspace permanently — always accessible alongside the plan and sessions
- Auth flows through pi's existing `/login` OAuth — no credentials in Pier
- Pi's auto-compaction handles context limits automatically — no token tracking needed in the UI

### What it is not

- Not a general-purpose chat tab
- Scoped to one workspace — the conversation lives next to the code it's about

### The Handoff

When the user clicks "Generate Plan":

1. The app takes the full discovery conversation transcript
2. Spawns a new pi session (tools enabled) with a structured prompt: read the transcript, write `plan.md` to the workspace root
3. The pi session emits `tool_execution_end` for the write to `plan.md` — Pier detects this event and reads the file
4. The plan panel updates

The pi session that generated the plan stays in the session list.

---

## 6. Planning Workflow

The planning workflow is a lightweight wrapper over pi, not a separate system.

### Plan Panel

Populated by the handoff from the Discover view. Shows:

- `plan.md` rendered as formatted markdown via `gioui.org/x/markdown` — read only
- "Open in $EDITOR" button — that is where editing happens
- Plan re-renders when the workspace is focused or switched to (no file watcher)
- "Create br Tasks" button beneath the plan

When the user clicks "Create br Tasks":

1. If br is not initialised, run `br init` first
2. Send `/create-tasks` as a prompt to the active pi session via RPC — this invokes the existing user-level prompt template (`~/.pi/agent/prompts/create-tasks.md`) which instructs pi to read `plan.md`, create tasks with proper titles, descriptions (why/what/acceptance criteria/key files), priorities, and dependencies via `br create` and `br depend`
3. The pi session's tool call timeline is visible live — each `br create` appears as a new task in the task panel immediately
4. When pi finishes, the task panel refreshes

No custom prompt construction needed in Pier — the user's existing `/create-tasks` prompt template handles the task decomposition logic. Task creation can be re-run at any time after editing `plan.md`.

---

## 7. Execute View

The main working environment. Pi sessions run here against the codebase, guided by br tasks.

### Task Panel

Docked panel showing the br task state for the workspace. **On-demand only** — no background polling. Task state refreshes when:

- A session linked to a task emits `agent_end` (task work completed)
- The user clicks "Refresh Tasks" in the panel
- A "Create br Tasks" flow completes
- The user manually closes a task
- The workspace is first opened
- The user switches to or focuses the task panel

Each refresh runs `br ready --json` and `br list --json` once in a goroutine. Results are cached until the next trigger. Shows:

- Ready tasks: unblocked, sorted by priority
- In Progress tasks: currently being worked
- Blocked tasks: with a note on what is blocking them

Each task card shows: id, title, priority, type, description excerpt.

### Session ↔ Task Linking

When the user starts a pi session from a task card:

- The task is automatically marked `in_progress` (`br update <id> --status in_progress`)
- The task id is shown in the session card in the sidebar

When the user closes a session with a task linked, a prompt appears: "Mark task complete?" with a one-line reason field → `br close <id> --reason "..."`. The task panel refreshes immediately after close.

---

## 8. Pi Integration Layer

Pi is run in RPC mode — structured JSON events on stdout, commands written as JSON to stdin. This is pi's designed embedding interface.

### Process Model

```
exec.Cmd: pi --mode rpc [options]
    │
    ├── stdout → bufio.Scanner goroutine → session.EventCh (chan Event)
    ├── stderr → bufio.Scanner goroutine → session.ErrCh  (chan string)
    └── stdin  ← JSON command writes from UI
```

The Gio render loop drains `session.EventCh` each frame and updates `session.State`. No locking needed on the read path — the render loop is single-threaded.

### RPC Commands (stdin → pi)

Commands are JSON objects written to pi's stdin, one per line.

| Command | Purpose |
|---------|---------|
| `prompt` | Send a user prompt. Supports `streamingBehavior: "steer"/"followUp"` during streaming |
| `steer` | Queue an interrupt message (delivered after current tool, skips remaining) |
| `follow_up` | Queue a message for after agent finishes |
| `abort` | Abort current operation |
| `get_state` | Get model, streaming status, session file, message count |
| `get_messages` | Get full conversation history |
| `set_model` | Switch model by provider + modelId |
| `cycle_model` | Cycle to next available model |
| `get_available_models` | List all configured models |
| `set_thinking_level` | Set thinking level: off/minimal/low/medium/high/xhigh |
| `get_commands` | Get available slash commands (for autocomplete) |
| `new_session` | Start fresh session |
| `switch_session` | Load a different session file |
| `fork` | Branch from a previous user message |
| `get_fork_messages` | Get messages available for forking |
| `compact` | Manual compaction |
| `set_auto_compaction` | Enable/disable auto-compaction |
| `get_session_stats` | Get token usage statistics |
| `bash` | Execute a shell command, add output to context |

All commands support an optional `id` field for request/response correlation.

### RPC Events (pi → stdout)

Events are JSONL, split on `\n` only. **Do not use generic line readers that split on Unicode line separators** (`U+2028`, `U+2029`) — these are valid inside JSON strings.

| Event | Description | Key Fields |
|-------|-------------|------------|
| `agent_start` | Agent begins processing a prompt | — |
| `agent_end` | Agent completes | `messages[]` — all messages generated this run |
| `turn_start` | New turn begins (one LLM response + tool calls) | — |
| `turn_end` | Turn completes | `message`, `toolResults[]` |
| `message_start` | Message begins | `message` (AgentMessage) |
| `message_update` | Streaming delta | `message`, `assistantMessageEvent` (see below) |
| `message_end` | Message completes | `message` (AgentMessage) |
| `tool_execution_start` | Tool begins | `toolCallId`, `toolName`, `args` |
| `tool_execution_update` | Tool streams partial output | `toolCallId`, `toolName`, `partialResult` (accumulated, not delta) |
| `tool_execution_end` | Tool completes | `toolCallId`, `toolName`, `result`, `isError` |
| `auto_compaction_start` | Auto-compaction triggered | `reason`: "threshold" or "overflow" |
| `auto_compaction_end` | Auto-compaction finished | `result.summary`, `aborted`, `willRetry` |
| `auto_retry_start` | Retrying after transient error | `attempt`, `maxAttempts`, `delayMs`, `errorMessage` |
| `auto_retry_end` | Retry resolved | `success`, `attempt`, `finalError` |
| `extension_error` | Extension threw an error | `extensionPath`, `event`, `error` |
| `extension_ui_request` | Extension requests UI interaction | `id`, `method`, fields per method |

### Streaming Delta Types (`assistantMessageEvent`)

The `message_update` event contains an `assistantMessageEvent` field with one of these types:

| Type | Description |
|------|-------------|
| `start` | Message generation started |
| `text_start` | Text content block started |
| `text_delta` | Text content chunk (`delta` field) |
| `text_end` | Text content block ended (`content` field has full text) |
| `thinking_start` | Thinking block started |
| `thinking_delta` | Thinking content chunk |
| `thinking_end` | Thinking block ended |
| `toolcall_start` | Tool call started |
| `toolcall_delta` | Tool call arguments chunk |
| `toolcall_end` | Tool call ended (includes full `toolCall` object) |
| `done` | Message complete (reason: `"stop"`, `"length"`, `"toolUse"`) |
| `error` | Error occurred (reason: `"aborted"`, `"error"`) |

### Extension UI Sub-Protocol

Extensions can request user interaction. Dialog methods (`select`, `confirm`, `input`, `editor`) emit an `extension_ui_request` and block until Pier sends back an `extension_ui_response` with the matching `id`. Fire-and-forget methods (`notify`, `setStatus`, `setWidget`, `setTitle`) do not expect a response.

Pier must handle these to support extensions that use `ctx.ui.*` methods (e.g., confirmation dialogs for dangerous bash commands).

### Deriving UI State from Events

Some UI states are derived, not directly emitted:

| UI State | How to Detect |
|----------|---------------|
| **Agent thinking** | Between `agent_start` and first `text_delta` or `toolcall_start` |
| **Agent streaming** | Between `text_start` and `text_end` |
| **Agent waiting for input** | After `agent_end` — no more events until next `prompt` |
| **Current model** | Response to `get_state` command (poll after `agent_end`) or track `set_model` responses |
| **Tool in progress** | Between `tool_execution_start` and `tool_execution_end` for a given `toolCallId` |
| **Compacting** | Between `auto_compaction_start` and `auto_compaction_end` |
| **Retrying** | Between `auto_retry_start` and `auto_retry_end` |

### Tool Output Truncation

Pi truncates tool output (default ~50KB). The `tool_execution_end` event's `result.details` may include `truncated: true` and `fullOutputPath` pointing to the complete output file. Pier should:

- Render the truncated output by default (fast, fits in view)
- Show an "Expand full output" action when `truncated: true`
- On expand, read the file at `fullOutputPath` and render in a scrollable block

`tool_execution_update` events contain accumulated `partialResult` (the full output so far, not just the new delta). When rendering streaming tool output, replace the display buffer entirely on each update rather than appending.

### Session State

```
Session
├── id:           string
├── workspace_id: string
├── task_id:      string           (optional br task link)
├── session_file: string           (pi's session file path, from get_state)
├── model:        string           (current model, from get_state)
├── status:       Thinking|Streaming|Waiting|Error|Compacting|Retrying
├── timeline:     []TimelineEntry  (rendered event history)
└── proc:         *exec.Cmd        (nil if not running)
```

---

## 9. br Integration Layer

The app never writes to `.beads/` directly. All br operations go through the br CLI. br is the authoritative source of truth; the app is a read-mostly consumer.

### Operations

| Action | Command | Trigger |
|--------|---------|---------|
| Initialise workspace | `br init` | First "Create br Tasks" click |
| Get ready tasks | `br ready --json` | On-demand (see below) |
| Get all tasks | `br list --json` | On-demand |
| Mark in progress | `br update <id> --status in_progress` | User starts session from task card |
| Close task | `br close <id> --reason "<text>"` | User confirms task complete |
| Show task detail | `br show <id> --json` | User clicks task card |

### On-Demand Refresh (No Polling)

Task state refreshes **only** when triggered by a user action or a meaningful session event:

- Workspace opened or focused
- Task panel manually refreshed (button or shortcut)
- `agent_end` received on a task-linked session
- Task created, closed, or status changed via the UI
- "Create br Tasks" flow completes

Each refresh runs in a goroutine, caches results, and invalidates Gio only if the output changed (diff the JSON). No timers, no background polling.

### br Binary Resolution

Resolved at startup: check config, then `PATH`, then `~/.cargo/bin/br` and `~/.local/bin/br`. If not found, the task panel shows a clear install prompt.

---

## 10. Bundled Pi Resources

Pier ships its own pi configuration rather than depending on the user's `~/.pi/agent/` directory. The user's local pi config is for their own terminal usage — Pier manages its own set of prompts, extensions, and settings that it passes to pi via CLI flags.

### What Pier Bundles

Embedded in the binary (via Go `embed`) or shipped alongside it:

```
pier/resources/
├── prompts/
│   └── create-tasks.md        task decomposition prompt (from plan.md → br tasks)
├── extensions/
│   └── (none initially — add as needed)
└── system-prompts/
    └── discover.md            system prompt for Discover mode (planning-focused)
```

### How Pier Passes Resources to Pi

Pi's CLI flags control resource loading per-session. Pier uses these to isolate its config from the user's:

| Pi Flag | Purpose |
|---------|---------|
| `--no-extensions` | Don't load user extensions |
| `--no-skills` | Don't load user skills |
| `--no-prompt-templates` | Don't load user prompts |
| `-e <path>` | Load a specific Pier-bundled extension |
| `--prompt-template <path>` | Load a specific Pier-bundled prompt template |
| `--append-system-prompt <text>` | Append Pier's system prompt additions |

**Discover sessions:** `pi --mode rpc --no-tools --no-extensions --no-skills --no-prompt-templates --append-system-prompt <discover.md>`

**Coding sessions:** `pi --mode rpc --prompt-template <pier-prompts-dir>` — loads Pier's bundled prompts (like `/create-tasks`) alongside pi's defaults. User extensions and skills are left enabled for coding sessions since they add capability.

### What Pier Does NOT Bundle

- **Auth** — pi's existing `auth.json` / OAuth is used as-is. Pier never touches credentials.
- **Models** — pi's built-in model registry + user's `models.json` are used as-is.
- **Skills for coding** — user's installed skills (br-tracker, commit-discipline, etc.) remain active in coding sessions. Pier doesn't replace them.
- **Packages** — user's installed pi packages (pi-subagents, pi-web-access, etc.) remain active in coding sessions.

### Why Not Use the User's Config

1. **Portability** — Pier must work on a fresh machine with just pi installed and authenticated. No assumption about what prompts/skills exist in `~/.pi/agent/`.
2. **Stability** — Pier's `/create-tasks` prompt is tuned for its workflow. If the user modifies their personal copy, Pier shouldn't break.
3. **Isolation** — Discover mode needs a clean, tool-free session. User extensions that assume tools exist would error.

### Resource Extraction

On every startup, Pier extracts bundled resources to `~/.config/pier/resources/`, overwriting unconditionally. The files are small (a few prompt templates and a system prompt). No version tracking needed.

---

## 11. Application Structure

```
pier/
├── main.go                    app wiring, event loop, key dispatch
│
├── app/
│   ├── app.go              package declaration
│   ├── keymap.go           pi passthrough + Pier UI action dispatch
│   └── theme.go            4-level dark/light palettes, 8px grid, typography
│
├── fonts/
│   ├── fonts.go            go:embed Inter + JetBrains Mono, NewShaper()
│   ├── Inter-*.ttf         Inter Regular/Medium/SemiBold/Bold (OFL)
│   └── JetBrainsMono-*.ttf JetBrains Mono Regular/Bold (OFL)
│
├── workspace/
│   └── workspace.go        Workspace struct, save/load ~/.config/pier/
│
├── session/
│   ├── session.go          Session struct, DrainEvents, SendPrompt
│   ├── process.go          pi process spawn/pipe/stop
│   ├── events.go           all RPC event/response type definitions
│   ├── commands.go         all RPC command types + ExtensionUIResponse
│   ├── decoder.go          JSONL decoder (\n-only split)
│   └── state.go            session state machine, timeline model
│
├── br/
│   ├── br.go               br CLI wrapper (List, Ready, Show, Close)
│   └── types.go            Task struct
│
├── ui/
│   ├── layout.go           package declaration
│   ├── sidebar.go          workspace list, hover states, accent bar, pill badges
│   ├── workspace_view.go   multi-session tabs, task linking
│   ├── discover_view.go    tool-free chat, Generate Plan handoff
│   ├── timeline.go         left-bordered messages, grouped spacing, model badges
│   ├── promptbar.go        bordered input, status line, slash autocomplete
│   ├── taskpanel.go        bordered cards, priority dots, hover, error accent
│   ├── planpanel.go        plan.md markdown rendering, Open in $EDITOR
│   ├── extensionui.go      dimmed backdrop modal, select/confirm/input dialogs
│   └── widgets/
│       ├── draw.go         WithAlpha, MulAlpha, Hovered, DrawRect, DrawBorderedRect
│       ├── hover.go        HoverState (pointer enter/leave tracking)
│       ├── toolblock.go    bordered collapsible card with error accent
│       ├── markdown.go     cached markdown renderer (Inter + JetBrains Mono)
│       ├── diff.go         green/red unified diff rendering
│       ├── animated_dot.go pulsing sine wave status dot
│       ├── empty_state.go  centered placeholder messages
│       ├── scroll_shadow.go top/bottom gradient overflow indicators
│       └── badge.go        package declaration
│
├── config/
│   └── config.go           AppConfig load/save (~/.config/pier/)
│
└── resources/
    ├── embed.go            go:embed FS for bundled prompts
    ├── extract.go          ExtractTo for startup extraction
    ├── system-prompts/
    │   └── discover.md     planning-focused system prompt
    └── prompts/
        └── create-tasks.md task decomposition prompt template
```

---

## 12. UI Layout

### Overall

```
┌──────────────────────────────────────────────────────────────────┐
│  Pier                                               [⚙ Settings]│
├───────────────┬──────────────────────────────────────────────────┤
│               │                                                  │
│  WORKSPACES   │   [Workspace view]                               │
│               │                                                  │
│  ● my-api     │                                                  │
│    2 sessions │                                                  │
│    ● thinking │                                                  │
│               │                                                  │
│  ● frontend   │                                                  │
│    1 session  │                                                  │
│    ○ waiting  │                                                  │
│               │                                                  │
│  [+ workspace]│                                                  │
└───────────────┴──────────────────────────────────────────────────┘
```

### Workspace View

```
┌──────────────────────────────────────────────────────────────────┐
│  my-api  ~/projects/api                                          │
├──────────────────────────┬───────────────────────────────────────┤
│                          │                                       │
│  DISCOVER                │  TASKS                    [↻ Refresh] │
│  ┌────────────────────┐  │                                       │
│  │ You                │  │  READY (3)                            │
│  │ I want to add      │  │  P1 bd-e9b1  Token exchange           │
│  │ OAuth2. Current    │  │  P1 bd-f2c3  DB schema                │
│  │ auth uses...       │  │  P2 bd-a1b4  Middleware hook          │
│  │                    │  │                                       │
│  │ Claude             │  │  IN PROGRESS (1)                      │
│  │ Good approach.     │  │  P1 bd-7f3a  OAuth flow  ●            │
│  │ Start with token   │  │                                       │
│  │ exchange before... │  │  BLOCKED (2)                          │
│  │                    │  │  P2 bd-c9d1  ← bd-7f3a                │
│  │                    │  │  P3 bd-d4e2  ← bd-c9d1                │
│  │ > [____________]   │  │                                       │
│  └────────────────────┘  ├───────────────────────────────────────┤
│  [Generate Plan →]       │                                       │
│                          │  SESSIONS                             │
│  PLAN                    │  SESSION 1  claude-s4  ● thinking     │
│  ┌────────────────────┐  │  SESSION 2  gpt-4o     ○ waiting      │
│  │ # OAuth2 Plan      │  │                                       │
│  │ ## Phase 1         │  │  [+ session]                          │
│  │ Token exchange...  │  │                                       │
│  └────────────────────┘  │                                       │
│  [Open in $EDITOR]       │                                       │
│  [Create br Tasks]       │                                       │
└──────────────────────────┴───────────────────────────────────────┘
```

### Session Timeline (focused session)

```
┌──────────────────────────────────────────────────────────────────┐
│  Session 1  [claude-s4]  ● thinking                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [read] src/auth/token.go                               [▼]      │
│                                                                  │
│  I can see the token validator is missing the expiry             │
│  check. I'll add it to oauth.go along with...                    │
│                                                                  │
│  [write] src/auth/oauth.go                              [▼]      │
│  ┌─ diff ─────────────────────────────────────────────────────┐  │
│  │  + func Exchange(code string) (*Token, error) {            │  │
│  │  +     resp, err := http.Post(tokenURL, ...                │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  [bash] go test ./auth/...                              [▼]      │
│  ok  github.com/user/api/auth  0.312s                            │
│                                                                  │
│  [bash] long-running-command (truncated)         [Expand full ↓] │
│  first 50KB of output shown...                                   │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  > [                                                  ] [send]   │
└──────────────────────────────────────────────────────────────────┘
```

### Sidebar Session Card

- Workspace name and short path
- Number of active sessions
- Status of most active session: thinking / streaming / waiting / error
- Unread indicator when a session needs attention and is not focused
- Multi-session priority: **most recently active** session gets attention first

---

## 13. Session Lifecycle

```
Click "+ session" or click a task card in READY list
        │
        ├── if from task card: task marked in_progress immediately
        │
        ▼
Choose model (optional, defaults to config default)
        │
        ▼
Spawn: pi --mode rpc [--session <file>] [--model <model>]
   in working directory: workspace.path
        │
        ▼
goroutine: stdout → JSONL decode → session.EventCh
goroutine: stderr → session.ErrCh
        │
        ▼
Gio render loop: drain EventCh each frame → update session.State
        │
        ├── user types prompt → write {"type":"prompt","message":"..."} to pi stdin
        ├── pi emits events → timeline updates live
        ├── agent_end → status becomes Waiting, prompt bar activates, session card rings
        ├── track model from get_state after agent_end
        └── extension_ui_request → show dialog, send extension_ui_response
        │
        ▼
Session ends (pi exits or user closes)
        │
        ├── if task linked: "Mark task complete?" prompt → br close
        ├── if task linked: task panel refreshes
        ├── session_file path saved for resume (from get_state)
        └── state persisted to workspace file
```

### Resuming a Session

Launch pi with `--session <path>` where path is the `session_file` stored in the workspace JSON (obtained from `get_state` response). Pi picks up from where it left off. The timeline is reconstructed from `get_messages` after startup.

---

## 14. Persistence Model

### Config File

`~/.config/pier/config.json`

```json
{
  "pi_path": "",
  "br_path": "",
  "default_model": "claude-sonnet-4-5",
  "theme": "dark"
}
```

Auth is handled entirely by pi (see [§19 Authentication](#19-authentication)). Run `pi /login` once in a terminal. Pier inherits that auth for all sessions. No credentials in Pier.

### Workspace State Files

`~/.config/pier/workspaces/<id>.json` — one per workspace.

Contains: workspace struct, discovery messages, plan path, br init status, session metadata (including `session_file` paths for resume).

### What pi Owns

Pi manages its own session files (default `~/.pi/agent/sessions/`). Pier stores only the `session_file` path pointer from `get_state`. Pier never reads or writes pi session files directly.

### What br Owns

br manages `.beads/beads.db` and `.beads/issues.jsonl` inside the workspace directory. Pier never touches these directly.

---

## 15. Keyboard Shortcuts

Shortcuts are split into three categories. Pi passthrough shortcuts send RPC commands to pi — they use the same keys as pi's terminal UI.

### Pi Passthrough (sent as RPC commands)

| Shortcut | RPC Command | Action |
|----------|-------------|--------|
| `Escape` | `abort` | Abort current agent run |
| `Ctrl+P` | `cycle_model` | Cycle to next model |
| `Ctrl+Shift+P` | `cycle_model` (backward) | Cycle model backward |
| `Shift+Tab` | `cycle_thinking_level` | Cycle thinking level |
| `Ctrl+L` | `cycle_model` | Select model |

### Session Lifecycle (pi process management)

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Clear prompt bar (first press) |
| `Ctrl+C` ×2 | Kill pi session (double-tap within 1s) |
| `Ctrl+W` | Close session |
| `Ctrl+Shift+N` | New session in current workspace |

### Pier UI

| Shortcut | Action |
|----------|--------|
| `Ctrl+O` | Collapse/expand all tool blocks |
| `Ctrl+B` | Toggle task panel |
| `Ctrl+R` | Refresh task panel |
| `Ctrl+/` | Focus prompt bar |
| `Ctrl+Shift+C` | Copy last assistant message |
| `Ctrl+N` | New workspace |
| `Ctrl+1..9` | Switch to workspace N |
| `Ctrl+Tab` | Cycle workspaces |

Shortcuts are hardcoded in `app/keymap.go`.

---

## 16. Configuration

| Key | Default | Description |
|-----|---------|-------------|
| `pi_path` | resolved from PATH | Path to pi binary |
| `br_path` | resolved from PATH | Path to br binary |
| `default_model` | `"claude-sonnet-4-5"` | Default model for new sessions |
| `theme` | `"dark"` | light / dark |

---

## 17. Visual Design

Pier follows a Slack-inspired design language built on consistent primitives.

### Design Tokens

**8px Grid.** All spacing derived from a 4px half-step base: 4 (XXS), 8 (XS), 12 (S), 16 (M), 24 (L), 32 (XL) dp. No magic numbers.

**4-Level Surface Stack.** Background → Surface → SurfaceAlt → SurfaceRaised. Each level is slightly lighter, creating depth without shadows.

| Level | Dark | Light | Usage |
|-------|------|-------|-------|
| Background | `#1a1a1e` | `#f8f8fa` | Deepest layer, main content area |
| Surface | `#242428` | `#ffffff` | Cards, sidebar, panels |
| SurfaceAlt | `#2e2e33` | `#f2f2f5` | Hover states, selected items |
| SurfaceRaised | `#38383e` | `#ffffff` | Modals, dropdowns, popovers |

**Opacity-Based Text Hierarchy.** One base text color with alpha variations:

| Role | Opacity | Usage |
|------|---------|-------|
| Primary | 90% | Body text, titles |
| Secondary | 60% | Supporting text, descriptions |
| Tertiary | 40% | Timestamps, captions, section headers |
| Disabled | 28% | Inactive elements |

### Typography

Two font families embedded via `go:embed`:

- **Inter** (Regular, Medium, SemiBold, Bold) — all UI text
- **JetBrains Mono** (Regular, Bold) — code blocks, tool output, task IDs, model badges

Scale: H1=24sp, H2=20sp, H3=16sp, Body=14sp, BodySmall=13sp, Mono=13sp, Caption=11sp.

### Component Patterns

**Hover feedback.** Sidebar items, tool blocks, task cards, and autocomplete options track `pointer.Enter`/`pointer.Leave` via a reusable `HoverState` widget. Hover brightens the background or border.

**Accent indicators.** Left border bars for visual anchoring: 2dp accent on assistant messages, 3dp accent on selected workspace, 3dp red on error tool calls.

**Pill badges.** Model names and task IDs render as rounded pills (SurfaceAlt background, MonoFace, BadgeRadius corners).

**Animated status dots.** Thinking and streaming states pulse via a sine wave (1.5s period, alpha 100→255→100) using `gtx.Execute(op.InvalidateCmd{})`. Static when idle.

**Bordered cards.** Tool blocks and task cards use `DrawBorderedRect` (fill + 1px border + radius). Border brightens on hover.

**Dimmed modal backdrop.** Extension dialogs overlay a 70% opacity Background dim, with the dialog card centered in SurfaceRaised.

### Color Utilities (`ui/widgets/draw.go`)

- `WithAlpha(c, a)` — return color with new alpha
- `MulAlpha(c, a)` — multiply existing alpha
- `Hovered(c)` — blend toward white (dark) or black (light)
- `Disabled(c)` — desaturate + reduce alpha
- `DrawRect`, `DrawBorderedRect`, `DrawLeftBorder`, `FillBackground` — shape helpers

---

## 18. Rendering Performance

This section documents the rendering performance strategy. The app is IO-bound (reading pi's RPC stream), not CPU-bound, so all performance risks are in the rendering layer.

### Streaming Text (`text_delta`)

During fast streaming, pi may emit 50+ `message_update` events per second per session. Each triggers a Gio invalidation.

**Strategy: two render modes per message.**

- **Streaming mode** (active message): Append-only plain text rendering. No markdown parsing. Each `text_delta` appends to a string buffer, Gio re-renders only the active message widget. Completed messages above are cached widget trees — not re-laid-out.
- **Complete mode** (after `message_end`): Full markdown render via `gioui.org/x/markdown`. Result is cached as `[]richtext.SpanStyle` and never re-parsed until the message is scrolled off-screen and back.

**Transition:** On `text_end` or `message_end`, switch from streaming mode to complete mode. Parse markdown once, cache the result.

### Tool Output Streaming (`tool_execution_update`)

Pi sends the **full accumulated output** on every `tool_execution_update`, not just the delta. For a command producing 10KB of output, later updates re-send the whole buffer.

**Strategy:** Track `len(previousPartialResult)`. On each update, render only the new bytes (from the previous length to the current length). If the tool block is collapsed, skip rendering entirely — just update the cached buffer.

### Timeline Scrollback

A long session may have 100+ tool calls and messages. Rendering all of them on every frame would be expensive.

**Strategy:** Use Gio's `widget.List` with lazy layout. Only visible items are laid out and drawn. Collapsed tool blocks render as a single-line header — their content is not laid out until expanded.

### Multi-Session Fan-In

With N sessions streaming simultaneously, N goroutines write to N channels. The Gio render loop drains all channels each frame.

**Strategy:** This is inherently fast — channel reads are nanoseconds. The bottleneck is always rendering, never event dispatch. Each session's timeline widget is independent. Only the focused session renders its full timeline; background sessions update their state but don't trigger layout for non-visible content.

### Markdown Rendering Budget

`gioui.org/x/markdown` is called only:

1. On `message_end` — once per completed assistant message
2. On plan.md when workspace is focused — once per focus switch
3. Never during streaming

If a single markdown render takes >16ms (blocking a frame), break it into chunks across frames. In practice this is unlikely for typical message sizes.

### Summary

| Hot Path | Mitigation |
|----------|------------|
| `text_delta` × 50/s | Plain text append, no markdown, only active message redrawn |
| `tool_execution_update` with full buffer | Diff against previous length, render delta only |
| Long timeline scroll | `widget.List` lazy layout, collapsed = single line |
| Multi-session streaming | Only focused session renders full timeline |
| Markdown parse | Only on `message_end`, cached, never during streaming |
| Tool output truncation | Show truncated by default, expand from `fullOutputPath` on demand |

---

## 19. Authentication

Pier never handles credentials directly. All authentication is managed by pi.

### How It Works

1. **One-time setup:** The user runs `pi /login` in a terminal (interactive mode only). Pi opens a browser for OAuth, receives tokens, and saves them to `~/.pi/agent/auth.json`.
2. **Pier inherits auth:** When Pier spawns `pi --mode rpc`, pi reads `auth.json` from disk. OAuth tokens auto-refresh via file locking. Multiple pi instances (including Pier's) can share the same auth file safely.
3. **No RPC login:** The `/login` command is interactive-mode-only — it requires a browser and user interaction that doesn't map to a JSON protocol. Pier cannot trigger OAuth flows.

### Auth Resolution (pi's priority order)

1. Runtime override (`--api-key` CLI flag)
2. API key from `auth.json`
3. OAuth token from `auth.json` (auto-refreshed with file locking)
4. Environment variable (e.g., `ANTHROPIC_API_KEY`)
5. Fallback resolver (custom providers from `models.json`)

### What Pier Should Do

- **On session start failure:** If pi errors because no auth is configured, show: "No authentication found. Run `pi /login` in a terminal first."
- **Never read `auth.json`:** Pier has no reason to inspect credentials. Pi handles all auth resolution internally.

---

## 20. Cross-Platform Notes

Gio abstracts the platform layer entirely. The same Go code produces:

| Platform | Backend | Window System |
|----------|---------|---------------|
| CachyOS / Linux | OpenGL or Vulkan | Wayland (preferred) or X11 |
| macOS | Metal | Cocoa |
| Windows | Direct3D | Win32 |

No conditional compilation needed. No platform-specific code.

Platform-sensitive items:

- **Config path:** `os.UserConfigDir()` — `~/.config` on Linux, `~/Library/Application Support` on macOS
- **Binary name:** `pier` / `pier.exe` — handled by `go build`
- **Pi auth:** `pi /login` writes credentials to pi's own config directory, cross-platform

Mac build: `GOOS=darwin GOARCH=arm64 go build -o pier-mac .`

---

## 21. Build and Distribution

```bash
# Development
go run .

# Release build (CachyOS)
go build -ldflags="-s -w" -o pier .

# Mac build (cross-compile from Linux)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o pier-mac .

# Install
install -m755 pier /usr/local/bin/pier
```

No install step required. Copy binary anywhere and run.

Gio requires a C compiler for CGo (`gcc` or `clang`). Only build dependency beyond Go itself. Present by default on CachyOS.

---

## 22. Build Order (completed)

All phases have been implemented.

### Phase 1 — Core Loop ✓

1. Pi RPC event parser — all event/command structs, JSONL decoder with `\n`-only splitting
2. Session process manager — spawn/pipe/stop with goroutine fan-out to channels
3. br wrapper — shell out to `br --json`, parse task structs
4. Minimal Gio window — sidebar + main area layout with theme
5. Wire one session end to end — full data path: prompt → pi stdin → events → state → timeline

### Phase 2 — Full Execute UI ✓

1. Timeline widget — markdown for complete messages, plain text during streaming, collapsible tool blocks
2. Extension UI handler — select, confirm, input, editor dialogs with modal overlay
3. Sidebar session cards — model pill badges, status dots, unread indicators
4. Prompt bar — slash command autocomplete from `get_commands`, bordered container
5. Task panel — ready/in-progress/blocked sections, on-demand refresh
6. Session ↔ task linking — start from task card auto-marks in_progress, close prompts completion
7. Multi-session — concurrent sessions with tab switching, only focused renders full timeline

### Phase 3 — Planning Workflow ✓

1. Discover view — pi with `--no-tools --no-extensions --no-skills`, chat thread
2. Plan handoff — conversation transcript → pi session → writes plan.md
3. Plan panel — markdown renderer, "Open in $EDITOR", re-render on focus
4. Create br Tasks — `/create-tasks` prompt template, task panel refreshes

### Phase 4 — Persistence and Polish ✓

1. Session persistence — workspace state saved to `~/.config/pier/`, session file paths for resume
2. Theme — dark (default) and light with 4-level surface stack, opacity-based text hierarchy
3. Config file — `~/.config/pier/config.json` with sensible defaults
4. Keyboard shortcuts — pi passthrough (Escape, Ctrl+P, Shift+Tab) + Pier UI (Ctrl+B, Ctrl+R, etc.)
5. Bundled resources — discover.md and create-tasks.md embedded via `go:embed`

### Phase 5 — UI Polish ✓

1. Embedded fonts — Inter (4 weights) + JetBrains Mono (2 weights) via `go:embed`
2. Color utilities — WithAlpha, MulAlpha, Hovered, Disabled + DrawRect/DrawBorderedRect helpers
3. Hover tracking — reusable HoverState widget via pointer events
4. Restyled sidebar — hover states, accent bar selection, pill badges
5. Restyled prompt bar — bordered container, status line, conditional send button
6. Restyled tool blocks — bordered cards, hover feedback, error accent bars
7. Restyled timeline — left accent border on assistant messages, message grouping, model badges
8. Restyled task panel — bordered cards, priority dots, hover, error states
9. Restyled extension dialogs — dimmed backdrop, centered card
10. Animated status dots — pulsing sine wave for thinking/streaming
11. Code block styling — JetBrains Mono in markdown renderer
12. Diff widget — green/red unified diff rendering
13. Empty states — contextual placeholder messages
14. Scroll shadows — top/bottom gradient overflow indicators
15. Stress tests — concurrent decoders, long timelines, concurrent state access

---

## 23. What You Get Over the Current Workflow

### Over Running pi in a Terminal

| Capability | Raw Terminal | Pier |
|------------|-------------|------|
| See which tool is running | No | Yes, with params, live |
| Inspect file diffs inline | No | Yes, collapsible diff widget |
| Know when pi is waiting | OSC hacks | First-class UI state, session card rings |
| Multiple sessions at once | tmux patchwork | Native, structured |
| Resume previous sessions | Manual session file mgmt | One click |
| Extension dialogs | Terminal TUI widgets | Native GUI dialogs |
| Tool output expansion | Ctrl+O in terminal | Click to expand from full output file |

### Over the Current Planning Workflow

| Step | Current | Pier |
|------|---------|------|
| Discovery conversation | Separate claude.ai tab, no connection to the codebase | In-app, persisted in the workspace alongside the plan and sessions |
| Plan generation | Copy/paste from claude.ai to pi manually | One button, transcript sent automatically |
| Task creation | Instruct pi manually, watch terminal | One button, tasks appear live in task panel |
| Tracking what's left | `br ready` in a separate terminal | Task panel alongside the coding session, refreshes on demand |
| Plan evolution | Edit plan.md externally, re-instruct pi | Edit in $EDITOR, re-run task creation from UI |
| Revisiting the why | Find the claude.ai conversation again | Always open in the workspace discover panel |
