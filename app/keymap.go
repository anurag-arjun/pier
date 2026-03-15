package app

import (
	"gioui.org/io/key"
)

// Action represents what Pier should do in response to a key press.
// Some actions translate to pi RPC commands, others are Pier-only UI actions.
type Action int

const (
	ActionNone Action = iota

	// --- Pi passthrough (translated to RPC commands) ---

	ActionAbort              // Escape → send "abort" to pi
	ActionCycleModel         // Ctrl+P → send "cycle_model"
	ActionCycleModelBackward // Ctrl+Shift+P → send "cycle_model" backward
	ActionCycleThinking      // Shift+Tab → send "cycle_thinking_level"
	ActionSelectModel        // Ctrl+L → get_available_models + show picker

	// --- Session lifecycle (pi process management) ---

	ActionClearEditor   // Ctrl+C (first) → clear prompt bar
	ActionKillSession   // Ctrl+C (second within 1s) → stop pi process
	ActionCloseSession  // Ctrl+W → stop pi process + close session
	ActionNewSession    // Ctrl+Shift+N → spawn new pi session in workspace

	// --- Pier-only UI actions ---

	ActionCollapseTools    // Ctrl+O → collapse/expand all tool blocks
	ActionToggleThinking   // Ctrl+T → collapse/expand thinking blocks
	ActionToggleTaskPanel  // Ctrl+B → show/hide task panel
	ActionRefreshTasks     // Ctrl+R → refresh br task list
	ActionFocusPrompt      // Ctrl+/ → focus prompt bar
	ActionCopyLastMessage  // Ctrl+Shift+C → copy last assistant text
	ActionNewWorkspace     // Ctrl+N → create new workspace
	ActionCycleWorkspace   // Ctrl+Tab → cycle workspaces
	ActionWorkspace1       // Ctrl+1..9 → switch workspace
	ActionWorkspace2
	ActionWorkspace3
	ActionWorkspace4
	ActionWorkspace5
	ActionWorkspace6
	ActionWorkspace7
	ActionWorkspace8
	ActionWorkspace9
)

// KeyFilters returns the key filters to register with Gio.
func KeyFilters() []key.Filter {
	return []key.Filter{
		// Pi passthrough
		{Name: key.NameEscape},
		{Name: "P", Required: key.ModShortcut},
		{Name: "P", Required: key.ModShortcut | key.ModShift},
		{Name: key.NameTab, Required: key.ModShift},
		{Name: "L", Required: key.ModShortcut},

		// Session lifecycle
		{Name: "C", Required: key.ModShortcut},
		{Name: "W", Required: key.ModShortcut},
		{Name: "N", Required: key.ModShortcut | key.ModShift},

		// Pier UI
		{Name: "O", Required: key.ModShortcut},
		{Name: "T", Required: key.ModShortcut},
		{Name: "B", Required: key.ModShortcut},
		{Name: "R", Required: key.ModShortcut},
		{Name: "/", Required: key.ModShortcut},
		{Name: "C", Required: key.ModShortcut | key.ModShift},
		{Name: "N", Required: key.ModShortcut},
		{Name: key.NameTab, Required: key.ModShortcut},
		{Name: "1", Required: key.ModShortcut},
		{Name: "2", Required: key.ModShortcut},
		{Name: "3", Required: key.ModShortcut},
		{Name: "4", Required: key.ModShortcut},
		{Name: "5", Required: key.ModShortcut},
		{Name: "6", Required: key.ModShortcut},
		{Name: "7", Required: key.ModShortcut},
		{Name: "8", Required: key.ModShortcut},
		{Name: "9", Required: key.ModShortcut},
	}
}

// MatchAction converts a key event to an Action.
func MatchAction(e key.Event) Action {
	if e.State != key.Press {
		return ActionNone
	}

	ctrl := e.Modifiers.Contain(key.ModShortcut)
	shift := e.Modifiers.Contain(key.ModShift)

	switch {
	// Pi passthrough
	case e.Name == key.NameEscape:
		return ActionAbort
	case ctrl && shift && e.Name == "P":
		return ActionCycleModelBackward
	case ctrl && e.Name == "P":
		return ActionCycleModel
	case shift && e.Name == key.NameTab:
		return ActionCycleThinking
	case ctrl && e.Name == "L":
		return ActionSelectModel

	// Session lifecycle
	case ctrl && shift && e.Name == "C":
		return ActionCopyLastMessage
	case ctrl && e.Name == "C":
		return ActionClearEditor // caller tracks double-tap → ActionKillSession
	case ctrl && e.Name == "W":
		return ActionCloseSession
	case ctrl && shift && e.Name == "N":
		return ActionNewSession

	// Pier UI
	case ctrl && e.Name == "O":
		return ActionCollapseTools
	case ctrl && e.Name == "T":
		return ActionToggleThinking
	case ctrl && e.Name == "B":
		return ActionToggleTaskPanel
	case ctrl && e.Name == "R":
		return ActionRefreshTasks
	case ctrl && e.Name == "/":
		return ActionFocusPrompt
	case ctrl && e.Name == "N":
		return ActionNewWorkspace
	case ctrl && e.Name == key.NameTab:
		return ActionCycleWorkspace
	case ctrl && e.Name >= "1" && e.Name <= "9":
		return ActionWorkspace1 + Action(e.Name[0]-'1')
	}

	return ActionNone
}

// IsPiCommand returns true if the action should be sent to pi as an RPC command.
func IsPiCommand(a Action) bool {
	switch a {
	case ActionAbort, ActionCycleModel, ActionCycleModelBackward,
		ActionCycleThinking, ActionSelectModel:
		return true
	}
	return false
}
