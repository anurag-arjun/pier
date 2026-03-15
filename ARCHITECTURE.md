# Pier — Architecture Document

A lightweight, native GUI for the pi coding agent. Manages multiple concurrent pi sessions, surfaces structured agent data invisible in a raw terminal, and covers the complete workflow from discovery conversation through task execution — all in one tool.

Stack: Go + Gio · Single binary · Wayland/X11 (CachyOS) · Mac-compatible (Metal) via `GOOS=darwin`

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
17. [Rendering Performance](#17-rendering-performance)
18. [Cross-Platform Notes](#18-cross-platform-notes)
19. [Build and Distribution](#19-build-and-distribution)
20. [Build Order](#20-build-order)
21. [What You Get Over the Current Workflow](#21-what-you-get-over-the-current-workflow)

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
| Diffing | `go-difflib` (Phase 5) | Custom diff widget for write-tool results. Initially, tool output rendered as monospace text — pi's output already contains diff information. |
| Binary | `go build` | Single static binary. No runtime deps beyond libc. |

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
├── main.go
│
├── app/
│   ├── app.go              top-level App struct, workspace list, Gio event loop
│   ├── keymap.go           global shortcut definitions and dispatch
│   └── theme.go            colours, typography, spacing constants
│
├── workspace/
│   ├── workspace.go        Workspace struct
│   └── persist.go          save/load ~/.config/pier/
│
├── session/
│   ├── session.go          Session struct and lifecycle (start/stop/restart)
│   ├── process.go          pi process management, stdin/stdout pipes
│   ├── events.go           RPC event parsing and dispatch
│   └── state.go            per-session UI state (timeline, model, status)
│
├── discover/
│   └── discover.go         DiscoverySession struct, message history, handoff
│
├── plan/
│   ├── plan.go             plan state, brief handling
│   └── renderer.go         markdown → Gio richtext via gioui.org/x/markdown
│
├── br/
│   ├── br.go               br CLI wrapper
│   ├── refresh.go          on-demand refresh goroutine with caching
│   └── types.go            Task, Status, Priority structs
│
├── ui/
│   ├── layout.go           top-level layout: sidebar + main area
│   ├── sidebar.go          workspace list, session cards
│   ├── workspace_view.go   discover panel, plan panel, task panel, session area
│   ├── discover_view.go    chat thread, model selector, generate plan button
│   ├── timeline.go         message/tool call timeline widget
│   ├── promptbar.go        input widget, slash command autocomplete
│   ├── taskpanel.go        br task list, ready/in-progress/blocked sections
│   ├── extensionui.go      extension UI dialog handler (select, confirm, input, editor)
│   └── widgets/
│       ├── badge.go        model/status badges
│       ├── toolblock.go    collapsible tool call block
│       └── markdown.go     markdown renderer wrapper (streaming vs complete modes)
│
└── config/
    └── config.go           user config struct, load/save, defaults
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
  "theme": "light"
}
```

Auth is handled entirely by pi. Run `pi /login` once to authenticate with your Claude Max subscription. Pier inherits that auth for all sessions. No credentials in Pier.

### Workspace State Files

`~/.config/pier/workspaces/<id>.json` — one per workspace.

Contains: workspace struct, discovery messages, plan path, br init status, session metadata (including `session_file` paths for resume).

### What pi Owns

Pi manages its own session files (default `~/.pi/agent/sessions/`). Pier stores only the `session_file` path pointer from `get_state`. Pier never reads or writes pi session files directly.

### What br Owns

br manages `.beads/beads.db` and `.beads/issues.jsonl` inside the workspace directory. Pier never touches these directly.

---

## 15. Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+N` | New workspace |
| `Ctrl+Shift+N` | New session in current workspace |
| `Ctrl+1..9` | Switch to workspace N |
| `Ctrl+Tab` | Cycle workspaces |
| `Ctrl+W` | Close focused session |
| `Ctrl+/` | Focus prompt bar |
| `Ctrl+B` | Toggle task panel |
| `Ctrl+Shift+C` | Copy last assistant message |
| `Ctrl+R` | Refresh task panel |
| `Escape` | Collapse all open tool blocks |

Shortcuts are hardcoded.

---

## 16. Configuration

| Key | Default | Description |
|-----|---------|-------------|
| `pi_path` | resolved from PATH | Path to pi binary |
| `br_path` | resolved from PATH | Path to br binary |
| `default_model` | `"claude-sonnet-4-5"` | Default model for new sessions |
| `theme` | `"light"` | light / dark |

---

## 17. Rendering Performance

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

## 18. Cross-Platform Notes

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

## 19. Build and Distribution

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

## 20. Build Order

### Phase 1 — Core Loop

Validate pi's RPC protocol before writing any UI.

1. **Pi RPC event parser** — define all event structs matching Section 8's event table, write the JSONL decoder, write a test harness that spawns `pi --mode rpc` and prints parsed events to stdout
2. **Session process manager** — spawn pi, read events into a channel, write commands to stdin
3. **br wrapper** — shell out to `br ready --json`, parse output, model the task structs
4. **Minimal Gio window** — blank window with a sidebar and main area
5. **Wire one session end to end** — spawn pi, send a prompt, see streaming text rendered in the timeline using the two-mode renderer (streaming plain text → complete markdown)

Phase 1 ends with a working single-session app, no persistence, minimal styling.

### Phase 2 — Full Execute UI

1. **Timeline widget** — messages, tool call blocks (with `toolCallId` correlation), collapsible monospace output, truncated output with expand
2. **Extension UI handler** — dialog rendering for `extension_ui_request` events (select, confirm, input)
3. **Sidebar session cards** with model and status badges
4. **Prompt bar** with slash command autocomplete (populated via `get_commands`)
5. **Task panel** — on-demand br refresh, showing ready/in-progress/blocked
6. **Session ↔ task linking** — start from task card, close task on session end, task panel refreshes
7. **Session switching** between multiple concurrent sessions

### Phase 3 — Planning Workflow

1. **Discover view** — spawn pi with `--no-tools --no-extensions --no-skills --mode rpc`, chat thread, model selector
2. **Handoff** — conversation transcript → plan generation pi session, detect `plan.md` write via `tool_execution_end` event
3. **Plan panel** — markdown renderer via `gioui.org/x/markdown`, "Open in $EDITOR" button, re-render on focus
4. **Create br Tasks flow** — send `/create-tasks` to pi session (user's existing prompt template), task panel refreshes on completion

### Phase 4 — Persistence and Polish

1. **Session persistence** — save/restore workspace state, resume pi sessions via `--session <path>`, reconstruct timeline via `get_messages`
2. **Theme** — light (default) and dark, hardcoded
3. **Config file** — full options with sensible defaults
4. **Mac build validation** — confirm Gio Metal backend, test config paths
5. **Rendering stress test** — 5 concurrent sessions, long timelines, verify frame budget holds

### Phase 5 — Deferred Enhancements

1. **Custom diff widget** (`ui/diff.go` + `go-difflib`) — green/red inline diff rendering for write/edit tool results. Until then, tool output rendered as monospace text (pi's output already contains diff information)

---

## 21. What You Get Over the Current Workflow

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
