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

- **Sidebar** (left) — your workspaces and sessions
- **Main area** (center) — session timeline, prompts, plan, discover
- **Task panel** (right) — br tasks, toggleable with Ctrl+B

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

- **Your messages** — labeled "You" in accent color
- **Assistant messages** — labeled with the model name (e.g., "claude-sonnet-4-5"), rendered as markdown when complete, plain text while streaming
- **Tool calls** — collapsible blocks showing `[read] path/to/file`, `[write] path/to/file`, `[bash] command`. Click the header to expand and see the full output
- **Notices** — compaction and retry events appear as inline notices

### Sending Prompts

Type in the prompt bar at the bottom and press **Enter** to send. The prompt bar is disabled while the agent is working — you'll see "Agent is working..." as placeholder text.

While the agent is running:
- **Escape** sends an `abort` command (stops the current run)
- The status dot changes: thinking (amber) → streaming (blue) → waiting (green)

### Slash Commands

Type `/` in the prompt bar to see all available pi commands. The autocomplete popup filters as you type:

```
/model          Open model selector
/new            Start a new session
/session        Show session info
/compact        Compact context
/tree           Open session tree
...
```

Click a command or keep typing to filter, then press Enter.

### Model and Thinking Control

These key shortcuts are forwarded directly to pi as RPC commands:

| Key | What happens |
|-----|-------------|
| **Ctrl+P** | Cycle to the next model |
| **Shift+Tab** | Cycle thinking level (off → low → medium → high) |
| **Ctrl+L** | Select model (currently cycles) |

The model name updates in the timeline header and sidebar after switching.

### Exiting a Session

| Method | What happens |
|--------|-------------|
| **Ctrl+C** | Clears the prompt bar |
| **Ctrl+C twice** (within 1 second) | Kills the pi process, returns to start screen |
| **Ctrl+W** | Closes the session immediately |
| Close the window | Stops all sessions and exits |

## Multiple Sessions

Press **Ctrl+Shift+N** to start another session in the same workspace. When you have multiple sessions:

- **Session tabs** appear above the timeline to switch between them
- The **sidebar** shows session cards under the workspace with model and status
- Only the **focused session** renders its full timeline (others update status only — saves CPU)
- Each session has its own prompt bar and independent state

## Task Panel

The task panel shows [br](https://github.com/Dicklesworthstone/beads_rust) tasks for the workspace, grouped into three sections:

### Ready
Tasks with no blockers, sorted by priority. These are what you should work on next.

### In Progress
Tasks currently being worked on.

### Blocked
Tasks waiting on other tasks to complete.

Each task card shows:
- **Priority badge** — P1 (red), P2 (amber), P3 (gray)
- **Task ID** — short br identifier
- **Title** — truncated to 2 lines

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

### 2. Generate Plan

When your thinking is clear, click **Generate Plan →**. Pier:
1. Takes the full discovery conversation transcript
2. Spawns a new pi session (with tools enabled)
3. Sends a structured prompt to write `plan.md` in your workspace root

The plan appears in the Plan panel as rendered markdown.

### 3. Plan Panel

Shows `plan.md` as formatted markdown. You can:
- **Read it** directly in Pier
- Click **Open in $EDITOR** to edit in your preferred editor (uses `$EDITOR`, `$VISUAL`, or `xdg-open`)
- The plan re-renders when you switch back to the workspace

### 4. Create Tasks

Click **Create br Tasks** below the plan to decompose it into trackable work. Pier sends the `/create-tasks` prompt to pi, which reads `plan.md` and creates br tasks with proper titles, descriptions, priorities, and dependencies.

Tasks appear in the task panel as they're created. You can re-run task creation after editing the plan.

### 5. Execute

Start sessions from task cards and work through them. The task panel shows your progress at all times.

## Extension Dialogs

Some pi extensions request user input (e.g., confirmation before running a dangerous command). When this happens, Pier shows a native dialog:

- **Select** — pick from a list of options
- **Confirm** — Yes/No question
- **Input** — text entry field
- **Editor** — multi-line text entry

The dialog appears as a modal overlay. Your response is sent back to pi and the extension continues.

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

## Workspace Persistence

Workspaces are saved to `~/.config/pier/workspaces/<id>.json` and restored on restart. Persisted state includes:

- Discovery conversation messages
- Plan file path
- Session metadata (including session file paths for resume)
- br initialization status

Pi's own session files (stored in `~/.pi/agent/sessions/`) are referenced by path — Pier never reads or writes them directly.

## Performance Notes

Pier is designed to stay smooth even with multiple sessions streaming simultaneously:

- **Streaming text** — plain text append during streaming, no markdown parsing until message completes
- **Tool output** — only new bytes rendered (diffed against previous partial result)
- **Background sessions** — update status badges only, no timeline layout
- **Long timelines** — lazy list rendering, only visible entries laid out
- **Collapsed tool blocks** — single-line header, content not rendered

## Troubleshooting

### "pi not found"

Make sure pi is in your PATH:

```bash
which pi
pi --version
```

### "br not found"

The task panel shows this when br isn't installed. Install it:

```bash
curl -fsSL "https://raw.githubusercontent.com/Dicklesworthstone/beads_rust/main/install.sh" | bash
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
