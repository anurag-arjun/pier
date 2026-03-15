package session

// RPC commands sent to pi's stdin as JSON lines.

// PromptCmd sends a user prompt.
type PromptCmd struct {
	ID                string `json:"id,omitempty"`
	Type              string `json:"type"` // "prompt"
	Message           string `json:"message"`
	StreamingBehavior string `json:"streamingBehavior,omitempty"` // "steer"|"followUp"
}

// SteerCmd queues an interrupt message.
type SteerCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "steer"
	Message string `json:"message"`
}

// FollowUpCmd queues a message for after agent finishes.
type FollowUpCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "follow_up"
	Message string `json:"message"`
}

// AbortCmd aborts the current operation.
type AbortCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "abort"
}

// GetStateCmd requests current session state.
type GetStateCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_state"
}

// GetMessagesCmd requests full conversation history.
type GetMessagesCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_messages"
}

// SetModelCmd switches the model.
type SetModelCmd struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"` // "set_model"
	Provider string `json:"provider"`
	ModelID  string `json:"modelId"`
}

// CycleModelCmd cycles to next model.
type CycleModelCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "cycle_model"
}

// GetAvailableModelsCmd lists all configured models.
type GetAvailableModelsCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_available_models"
}

// SetThinkingLevelCmd sets thinking level.
type SetThinkingLevelCmd struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type"` // "set_thinking_level"
	Level string `json:"level"` // "off"|"minimal"|"low"|"medium"|"high"|"xhigh"
}

// CycleThinkingLevelCmd cycles thinking level.
type CycleThinkingLevelCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "cycle_thinking_level"
}

// SetSteeringModeCmd sets steering mode.
type SetSteeringModeCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "set_steering_mode"
	Mode string `json:"mode"` // "all"|"one-at-a-time"
}

// SetFollowUpModeCmd sets follow-up mode.
type SetFollowUpModeCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "set_follow_up_mode"
	Mode string `json:"mode"` // "all"|"one-at-a-time"
}

// GetCommandsCmd gets available slash commands.
type GetCommandsCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_commands"
}

// NewSessionCmd starts a fresh session.
type NewSessionCmd struct {
	ID            string `json:"id,omitempty"`
	Type          string `json:"type"` // "new_session"
	ParentSession string `json:"parentSession,omitempty"`
}

// SwitchSessionCmd loads a different session file.
type SwitchSessionCmd struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"` // "switch_session"
	SessionPath string `json:"sessionPath"`
}

// ForkCmd branches from a previous user message.
type ForkCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "fork"
	EntryID string `json:"entryId"`
}

// GetForkMessagesCmd gets messages available for forking.
type GetForkMessagesCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_fork_messages"
}

// CompactCmd triggers manual compaction.
type CompactCmd struct {
	ID                 string `json:"id,omitempty"`
	Type               string `json:"type"` // "compact"
	CustomInstructions string `json:"customInstructions,omitempty"`
}

// SetAutoCompactionCmd enables/disables auto-compaction.
type SetAutoCompactionCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "set_auto_compaction"
	Enabled bool   `json:"enabled"`
}

// SetAutoRetryCmd enables/disables auto-retry.
type SetAutoRetryCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "set_auto_retry"
	Enabled bool   `json:"enabled"`
}

// AbortRetryCmd cancels in-progress retry.
type AbortRetryCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "abort_retry"
}

// BashCmd executes a shell command.
type BashCmd struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"` // "bash"
	Command string `json:"command"`
}

// AbortBashCmd cancels running bash command.
type AbortBashCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "abort_bash"
}

// GetSessionStatsCmd gets token usage statistics.
type GetSessionStatsCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_session_stats"
}

// ExportHTMLCmd exports session to HTML.
type ExportHTMLCmd struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type"` // "export_html"
	OutputPath string `json:"outputPath,omitempty"`
}

// GetLastAssistantTextCmd gets last assistant message text.
type GetLastAssistantTextCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "get_last_assistant_text"
}

// SetSessionNameCmd sets a display name for the session.
type SetSessionNameCmd struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "set_session_name"
	Name string `json:"name"`
}

// ExtensionUIResponse is sent back to pi for extension dialog results.
type ExtensionUIResponse struct {
	Type      string `json:"type"` // "extension_ui_response"
	ID        string `json:"id"`
	Value     string `json:"value,omitempty"`
	Confirmed *bool  `json:"confirmed,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
}

// NewPromptCmd creates a prompt command.
func NewPromptCmd(message string) PromptCmd {
	return PromptCmd{Type: "prompt", Message: message}
}

// NewGetStateCmd creates a get_state command.
func NewGetStateCmd() GetStateCmd {
	return GetStateCmd{Type: "get_state"}
}

// NewGetMessagesCmd creates a get_messages command.
func NewGetMessagesCmd() GetMessagesCmd {
	return GetMessagesCmd{Type: "get_messages"}
}

// NewGetCommandsCmd creates a get_commands command.
func NewGetCommandsCmd() GetCommandsCmd {
	return GetCommandsCmd{Type: "get_commands"}
}

// NewAbortCmd creates an abort command.
func NewAbortCmd() AbortCmd {
	return AbortCmd{Type: "abort"}
}

// NewSetModelCmd creates a set_model command.
func NewSetModelCmd(provider, modelID string) SetModelCmd {
	return SetModelCmd{Type: "set_model", Provider: provider, ModelID: modelID}
}

// NewGetAvailableModelsCmd creates a get_available_models command.
func NewGetAvailableModelsCmd() GetAvailableModelsCmd {
	return GetAvailableModelsCmd{Type: "get_available_models"}
}
