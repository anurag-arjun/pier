# Pier User Guide

## Getting Started

### First Launch

```bash
cd ~/your-project
pier
```

Pier opens a window with three panels:

```
┌──────────┬────────────────────────────────┬──────────┐
│          │                                │          │
│ Sidebar  │         Main Area              │  Tasks   │
│          │                                │          │
│          │   Click "Start Pi Session"     │          │
│          │   to begin                     │          │
│          │                                │          │
└──────────┴────────────────────────────────┴──────────┘
```

- **Sidebar** (left, 240dp) — your workspaces and sessions
- **Main area** (center, flexible) — session timeline, prompts, plan, discover
- **Task panel** (right, 260dp) — br tasks, toggleable with Ctrl+B

On first launch, Pier creates a workspace for the current directory and saves config to `~/.config/pier/`.

### Prerequisites

Make sure pi is installed and authenticated:

```bash
pi /login
```

For task tracking, install [br](https://github.com/Dicklesworthstone/beads_rust):

```bash
curl -fsSL "https://raw.githubusercontent.com/Dicklesworthstone/beads_rust/main/install.sh" | bash
```

## Sessions

### Starting a Session

Click **Start Pi Session** in the main area. Pier spawns `pi --mode rpc` in your workspace directory. The session appears in the sidebar with a colored status dot.

### The Timeline

Every interaction with pi appears in the timeline:

- **Your messages** — labeled "You" in accent color, no left border
- **Assistant messages** — marked by a 2dp left accent border and indented, with the model name shown as a pill badge (e.g., `claude-sonnet-4-5` in monospace). Rendered as plain text while streaming, converted to full markdown (with JetBrains Mono code blocks) on completion
- **Tool calls** — bordered card widgets showing the tool name in accent monospace and args in tertiary text. Click the header to expand and see the full output. Error tool calls get a red left accent bar and an `ERROR` badge. Output renders in JetBrains Mono, truncated at 4000 characters with a "truncated" indicator
- **Notices** — compaction and retry events appear as inline colored notices

Messages from the same role are grouped with tight spacing (4dp); role changes get 16dp breathing room.

A streaming cursor (▍) appears at the end of text while the assistant is generating.

### Sending Prompts

Type in the prompt bar at the bottom and press **Enter** to send. The prompt bar has:

- A 1px border that highlights when focused
- A send button (→) that only appears when you've typed something
- A status line above showing "● Thinking..." or "● Streaming..." with a colored dot while the agent works
- "Agent is working..." placeholder when the agent is busy

### Slash Commands

Type `/` in the prompt bar to see all available pi commands. The autocomplete popup shows commands in a bordered card with monospace names:

```
/model          Open model selector
/new            Start a new session
/session        Show session info
/compact        Compact context
/tree           Open session tree
...
```

Click a command or keep typing to filter (max 8 shown), then press Enter.

### Model and Thinking Control

These key shortcuts are forwarded directly to pi as RPC commands — the same bindings pi uses in its terminal UI:

| Key | What happens |
|-----|-------------|
| **Ctrl+P** | Cycle to the next model |
| **Shift+Tab** | Cycle thinking level (off → low → medium → high) |
| **Ctrl+L** | Select model (currently cycles) |

The model name updates in the timeline pill badge and sidebar session card after switching.

### Exiting a Session

| Method | What happens |
|--------|-------------|
| **Escape** | Sends `abort` to pi (stops the current agent run) |
| **Ctrl+C** | Clears the prompt bar |
| **Ctrl+C twice** (within 1 second) | Kills the pi process, returns to start screen |
| **Ctrl+W** | Closes the session immediately |
| Close the window | Stops all sessions and exits |

## Multiple Sessions

Press **Ctrl+Shift+N** to start another session in the same workspace. When you have multiple sessions:

- **Session tabs** appear above the timeline to switch between them
- The **sidebar** shows session cards under the workspace — each with a model pill badge, status dot, and optional task ID
- Only the **focused session** renders its full timeline (others update status only — saves CPU)
- Each session has its own prompt bar and independent state

## Sidebar

The sidebar shows your workspaces and their sessions:

- **Workspace items** highlight on hover (SurfaceAlt background) and show a 3px left accent bar when selected
- **Session cards** appear indented under the active workspace, showing model as a rounded pill badge in monospace
- **Status dots** use animated pulsing for thinking (amber) and streaming (blue), solid for waiting (green) and error (red)
- **Unread indicator** — a small accent dot appears on session cards that received `agent_end` while unfocused
- **Section header** "WORKSPACES" in uppercase tertiary text

## Task Panel

The task panel shows [br](https://github.com/Dicklesworthstone/beads_rust) tasks for the workspace, grouped into three sections:

### Ready
Tasks with no blockers, sorted by priority. These are what you should work on next.

### In Progress
Tasks currently being worked on.

### Blocked
Tasks waiting on other tasks to complete. Cards appear dimmed at 50% opacity.

Each task card is a bordered widget with:
- **Priority dot** — 8dp colored circle: P1 (red), P2 (amber), P3 (gray)
- **Task ID** — monospace, tertiary opacity
- **Title** — medium weight, truncated to 2 lines
- **Hover** — border brightens for feedback

### Empty States

When there are no tasks, the panel shows contextual messages:
- "Task tracking not initialized." with `br init` hint if br hasn't been set up
- "No tasks yet. Create a plan, then generate tasks" otherwise

### Error State

If br is not found or returns an error, a red-accented error card appears with the error message.

### Refreshing

Tasks refresh when:
- You press **Ctrl+R**
- You click the **↻** button in the task panel header
- A session linked to a task completes

No background polling — refresh is always explicit.

### Toggling

Press **Ctrl+B** to show/hide the task panel.

### Starting a Session from a Task

When you click a task card in the Ready section, Pier:
1. Creates a new pi session linked to that task
2. Automatically marks the task as `in_progress` via br
3. Shows the task ID in the session's status bar

When you close a session that has a linked task, you'll be prompted to mark it complete.

## The Discover → Plan → Execute Workflow

Pier supports a complete workflow without switching tools.

### 1. Discover

The Discover view is a focused chat for thinking through your approach **before writing code**. Pi runs without tools here — no file reads, no code writes, just conversation.

Use it to:
- Work out requirements and constraints
- Discuss architectural decisions
- Identify tradeoffs before committing

When no session is active, the discover view shows: "Think through your approach before writing code."

### 2. Generate Plan

When your thinking is clear, click **Generate Plan →**. Pier:
1. Takes the full discovery conversation transcript
2. Spawns a new pi session (with tools enabled)
3. Sends a structured prompt to write `plan.md` in your workspace root

The plan appears in the Plan panel as rendered markdown.

### 3. Plan Panel

Shows `plan.md` as formatted markdown with Inter body text and JetBrains Mono code blocks. You can:
- **Read it** directly in Pier
- Click **Open in $EDITOR** to edit in your preferred editor (uses `$EDITOR`, `$VISUAL`, or `xdg-open`)
- The plan re-renders when you switch back to the workspace

When no plan exists, the panel shows a placeholder message.

### 4. Create Tasks

Click **Create br Tasks** below the plan to decompose it into trackable work. Pier sends the `/create-tasks` prompt to pi, which reads `plan.md` and creates br tasks with proper titles, descriptions, priorities, and dependencies.

Tasks appear in the task panel as they're created. You can re-run task creation after editing the plan.

### 5. Execute

Start sessions from task cards and work through them. The task panel shows your progress at all times.

## Extension Dialogs

Some pi extensions request user input (e.g., confirmation before running a dangerous command). When this happens, Pier shows a modal dialog:

- **Dimmed backdrop** — the entire window behind the dialog is dimmed at 70% opacity
- **Centered card** — SurfaceRaised background, 12dp rounded corners, 1px border, 24dp padding
- **Select** — hoverable option rows (SurfaceAlt highlight on hover)
- **Confirm** — title, message, accent "Yes" and subtle "Cancel" buttons
- **Input** — bordered text field matching the prompt bar style, with Submit and Cancel buttons

The dialog blocks interaction with the session until dismissed. Escape closes dialogs.

## Configuration

Config is stored at `~/.config/pier/config.json`:

```json
{
  "pi_path": "",
  "br_path": "",
  "default_model": "claude-sonnet-4-5",
  "theme": "dark"
}
```

| Key | Default | Description |
|-----|---------|-------------|
| `pi_path` | (resolved from PATH) | Path to pi binary |
| `br_path` | (resolved from PATH) | Path to br binary |
| `default_model` | `claude-sonnet-4-5` | Default model for new sessions |
| `theme` | `dark` | `light` or `dark` |

Leave `pi_path` and `br_path` empty to auto-resolve from PATH, `~/.cargo/bin/`, or `~/.local/bin/`.

### Theme

Pier ships with two themes. Both use a 4-level surface stack for visual depth and opacity-based text hierarchy:

| Level | Dark | Light |
|-------|------|-------|
| Background | #1a1a1e | #f8f8fa |
| Surface | #242428 | #ffffff |
| SurfaceAlt | #2e2e33 | #f2f2f5 |
| SurfaceRaised | #38383e | #ffffff |

Text uses the base color at 90% (primary), 60% (secondary), 40% (tertiary), or 28% (disabled) opacity.

## Workspace Persistence

Workspaces are saved to `~/.config/pier/workspaces/<id>.json` and restored on restart. Persisted state includes:

- Discovery conversation messages
- Plan file path
- Session metadata (including session file paths for resume)
- br initialization status

Pi's own session files (stored in `~/.pi/agent/sessions/`) are referenced by path — Pier never reads or writes them directly.

## Performance Notes

Pier is designed to stay smooth even with multiple sessions streaming simultaneously:

- **Streaming text** — plain text append during streaming, markdown parsing deferred until message completes
- **Tool output** — only new bytes rendered (diffed against previous partial result length)
- **Background sessions** — update status badges only, no timeline layout
- **Long timelines** — lazy list rendering via `widget.List`, only visible entries laid out
- **Collapsed tool blocks** — single-line header, body content not rendered until expanded
- **Markdown caching** — rendered spans cached per message, invalidated only on theme change

## Troubleshooting

### "pi not found"

Make sure pi is in your PATH:

```bash
which pi
pi --version
```

### "br not found" / "Task tracking not initialized"

The task panel shows these when br isn't installed or initialized. Install br:

```bash
curl -fsSL "https://raw.githubusercontent.com/Dicklesworthstone/beads_rust/main/install.sh" | bash
```

Initialize in your project:

```bash
cd ~/your-project
br init
```

Or set `br_path` in config.json to the full path.

### Session starts but nothing happens

Check that pi is authenticated:

```bash
pi /login
```

### Window doesn't open

Gio requires a working display server. On Wayland:

```bash
# Should already work on most Wayland compositors
pier
```

On X11:

```bash
# Gio auto-detects X11 as fallback
pier
```

### Build fails with CGo errors

Gio requires a C compiler. Install one:

```bash
# Arch/CachyOS
sudo pacman -S gcc

# Ubuntu/Debian
sudo apt install gcc

# macOS (usually already present)
xcode-select --install
```
