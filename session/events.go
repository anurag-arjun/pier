// Package session manages pi process lifecycle and RPC event streams.
package session

import "encoding/json"

// --- Top-level event envelope ---

// RawEvent is the initial parse target. The Type field determines
// which concrete struct the payload should be decoded into.
type RawEvent struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"` // full line for re-parsing
}

// --- Agent lifecycle events (from pi-agent-core AgentEvent) ---

type AgentStartEvent struct {
	Type string `json:"type"` // "agent_start"
}

type AgentEndEvent struct {
	Type     string         `json:"type"` // "agent_end"
	Messages []AgentMessage `json:"messages"`
}

type TurnStartEvent struct {
	Type string `json:"type"` // "turn_start"
}

type TurnEndEvent struct {
	Type        string           `json:"type"` // "turn_end"
	Message     AgentMessage     `json:"message"`
	ToolResults []ToolResultMsg  `json:"toolResults"`
}

type MessageStartEvent struct {
	Type    string       `json:"type"` // "message_start"
	Message AgentMessage `json:"message"`
}

type MessageUpdateEvent struct {
	Type                  string                `json:"type"` // "message_update"
	Message               AgentMessage          `json:"message"`
	AssistantMessageEvent AssistantMessageEvent `json:"assistantMessageEvent"`
}

type MessageEndEvent struct {
	Type    string       `json:"type"` // "message_end"
	Message AgentMessage `json:"message"`
}

// --- Tool execution events ---

type ToolExecutionStartEvent struct {
	Type       string          `json:"type"` // "tool_execution_start"
	ToolCallID string          `json:"toolCallId"`
	ToolName   string          `json:"toolName"`
	Args       json.RawMessage `json:"args"`
}

type ToolExecutionUpdateEvent struct {
	Type          string          `json:"type"` // "tool_execution_update"
	ToolCallID    string          `json:"toolCallId"`
	ToolName      string          `json:"toolName"`
	Args          json.RawMessage `json:"args"`
	PartialResult json.RawMessage `json:"partialResult"`
}

type ToolExecutionEndEvent struct {
	Type       string          `json:"type"` // "tool_execution_end"
	ToolCallID string          `json:"toolCallId"`
	ToolName   string          `json:"toolName"`
	Result     json.RawMessage `json:"result"`
	IsError    bool            `json:"isError"`
}

// --- Session-level events (from AgentSessionEvent) ---

type AutoCompactionStartEvent struct {
	Type   string `json:"type"` // "auto_compaction_start"
	Reason string `json:"reason"` // "threshold" | "overflow"
}

type AutoCompactionEndEvent struct {
	Type         string           `json:"type"` // "auto_compaction_end"
	Result       *CompactionResult `json:"result"`
	Aborted      bool             `json:"aborted"`
	WillRetry    bool             `json:"willRetry"`
	ErrorMessage string           `json:"errorMessage,omitempty"`
}

type AutoRetryStartEvent struct {
	Type         string `json:"type"` // "auto_retry_start"
	Attempt      int    `json:"attempt"`
	MaxAttempts  int    `json:"maxAttempts"`
	DelayMs      int    `json:"delayMs"`
	ErrorMessage string `json:"errorMessage"`
}

type AutoRetryEndEvent struct {
	Type       string `json:"type"` // "auto_retry_end"
	Success    bool   `json:"success"`
	Attempt    int    `json:"attempt"`
	FinalError string `json:"finalError,omitempty"`
}

// --- Extension events ---

type ExtensionErrorEvent struct {
	Type          string `json:"type"` // "extension_error"
	ExtensionPath string `json:"extensionPath"`
	Event         string `json:"event"`
	Error         string `json:"error"`
}

// ExtensionUIRequest covers all extension UI dialog methods.
// Method determines which fields are populated.
type ExtensionUIRequest struct {
	Type    string `json:"type"` // "extension_ui_request"
	ID      string `json:"id"`
	Method  string `json:"method"` // "select"|"confirm"|"input"|"editor"|"notify"|"setStatus"|"setWidget"|"setTitle"|"set_editor_text"

	// select
	Title   string   `json:"title,omitempty"`
	Options []string `json:"options,omitempty"`

	// confirm
	Message string `json:"message,omitempty"`

	// input
	Placeholder string `json:"placeholder,omitempty"`

	// editor
	Prefill string `json:"prefill,omitempty"`

	// notify
	NotifyType string `json:"notifyType,omitempty"` // "info"|"warning"|"error"

	// setStatus
	StatusKey  string `json:"statusKey,omitempty"`
	StatusText string `json:"statusText,omitempty"`

	// setWidget
	WidgetKey       string   `json:"widgetKey,omitempty"`
	WidgetLines     []string `json:"widgetLines,omitempty"`
	WidgetPlacement string   `json:"widgetPlacement,omitempty"`

	// set_editor_text
	Text string `json:"text,omitempty"`

	// shared
	Timeout *int `json:"timeout,omitempty"`
}

// --- RPC response (for command acknowledgements) ---

type RpcResponse struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"` // "response"
	Command string          `json:"command"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// --- Parsed response data types ---

type RpcSessionState struct {
	Model                  *ModelInfo `json:"model,omitempty"`
	ThinkingLevel          string     `json:"thinkingLevel"`
	IsStreaming            bool       `json:"isStreaming"`
	IsCompacting           bool       `json:"isCompacting"`
	SteeringMode           string     `json:"steeringMode"`
	FollowUpMode           string     `json:"followUpMode"`
	SessionFile            string     `json:"sessionFile,omitempty"`
	SessionID              string     `json:"sessionId"`
	SessionName            string     `json:"sessionName,omitempty"`
	AutoCompactionEnabled  bool       `json:"autoCompactionEnabled"`
	MessageCount           int        `json:"messageCount"`
	PendingMessageCount    int        `json:"pendingMessageCount"`
}

type RpcSlashCommand struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source"` // "extension"|"prompt"|"skill"
	Location    string `json:"location,omitempty"`
	Path        string `json:"path,omitempty"`
}

// --- Shared data structures ---

// ModelInfo matches pi-ai's Model interface (subset we care about).
type ModelInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	API           string    `json:"api"`
	Provider      string    `json:"provider"`
	BaseURL       string    `json:"baseUrl"`
	Reasoning     bool      `json:"reasoning"`
	Input         []string  `json:"input"`
	Cost          ModelCost `json:"cost"`
	ContextWindow int       `json:"contextWindow"`
	MaxTokens     int       `json:"maxTokens"`
}

type ModelCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
}

// AgentMessage is the union of user/assistant/toolResult messages.
// We keep it loosely typed since the full structure is complex.
type AgentMessage struct {
	Role      string          `json:"role"` // "user"|"assistant"|"toolResult"|"custom"
	Content   json.RawMessage `json:"content"`
	Timestamp float64         `json:"timestamp"`

	// assistant-specific
	API          string  `json:"api,omitempty"`
	Provider     string  `json:"provider,omitempty"`
	Model        string  `json:"model,omitempty"`
	StopReason   string  `json:"stopReason,omitempty"`
	ErrorMessage string  `json:"errorMessage,omitempty"`
	Usage        *Usage  `json:"usage,omitempty"`

	// toolResult-specific
	ToolCallID string `json:"toolCallId,omitempty"`
	ToolName   string `json:"toolName,omitempty"`
	IsError    bool   `json:"isError,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
}

type Usage struct {
	Input      int `json:"input"`
	Output     int `json:"output"`
	CacheRead  int `json:"cacheRead"`
	CacheWrite int `json:"cacheWrite"`
}

type ToolResultMsg = AgentMessage // alias for clarity in TurnEndEvent

// CompactionResult from pi
type CompactionResult struct {
	Summary string `json:"summary,omitempty"`
}

// --- AssistantMessageEvent sub-types ---

// AssistantMessageEvent is the streaming delta type inside message_update events.
type AssistantMessageEvent struct {
	Type         string        `json:"type"` // "start"|"text_start"|"text_delta"|"text_end"|"thinking_start"|"thinking_delta"|"thinking_end"|"toolcall_start"|"toolcall_delta"|"toolcall_end"|"done"|"error"
	ContentIndex int           `json:"contentIndex,omitempty"`
	Delta        string        `json:"delta,omitempty"`   // text_delta, thinking_delta, toolcall_delta
	Content      string        `json:"content,omitempty"` // text_end, thinking_end (full content)
	ToolCall     *ToolCallInfo `json:"toolCall,omitempty"` // toolcall_end
	Reason       string        `json:"reason,omitempty"`  // done: "stop"|"length"|"toolUse", error: "aborted"|"error"
	Partial      *AgentMessage `json:"partial,omitempty"` // partial assistant message
	Message      *AgentMessage `json:"message,omitempty"` // done: final message
	Error        *AgentMessage `json:"error,omitempty"`   // error: error message
}

type ToolCallInfo struct {
	Type string          `json:"type"` // "toolCall"
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// SessionStats from get_session_stats response
type SessionStats struct {
	SessionFile       string      `json:"sessionFile"`
	SessionID         string      `json:"sessionId"`
	UserMessages      int         `json:"userMessages"`
	AssistantMessages int         `json:"assistantMessages"`
	ToolCalls         int         `json:"toolCalls"`
	ToolResults       int         `json:"toolResults"`
	TotalMessages     int         `json:"totalMessages"`
	Tokens            TokenStats  `json:"tokens"`
	Cost              float64     `json:"cost"`
}

type TokenStats struct {
	Input      int `json:"input"`
	Output     int `json:"output"`
	CacheRead  int `json:"cacheRead"`
	CacheWrite int `json:"cacheWrite"`
	Total      int `json:"total"`
}

// Event is the union type returned by the decoder.
// Exactly one field is non-nil.
type Event struct {
	// Core agent events
	AgentStart  *AgentStartEvent
	AgentEnd    *AgentEndEvent
	TurnStart   *TurnStartEvent
	TurnEnd     *TurnEndEvent
	MsgStart    *MessageStartEvent
	MsgUpdate   *MessageUpdateEvent
	MsgEnd      *MessageEndEvent

	// Tool events
	ToolStart   *ToolExecutionStartEvent
	ToolUpdate  *ToolExecutionUpdateEvent
	ToolEnd     *ToolExecutionEndEvent

	// Session events
	CompactStart *AutoCompactionStartEvent
	CompactEnd   *AutoCompactionEndEvent
	RetryStart   *AutoRetryStartEvent
	RetryEnd     *AutoRetryEndEvent

	// Extension events
	ExtError    *ExtensionErrorEvent
	ExtUIReq    *ExtensionUIRequest

	// RPC responses
	Response    *RpcResponse
}

// EventType returns the event type string for dispatch.
func (e *Event) EventType() string {
	switch {
	case e.AgentStart != nil:
		return "agent_start"
	case e.AgentEnd != nil:
		return "agent_end"
	case e.TurnStart != nil:
		return "turn_start"
	case e.TurnEnd != nil:
		return "turn_end"
	case e.MsgStart != nil:
		return "message_start"
	case e.MsgUpdate != nil:
		return "message_update"
	case e.MsgEnd != nil:
		return "message_end"
	case e.ToolStart != nil:
		return "tool_execution_start"
	case e.ToolUpdate != nil:
		return "tool_execution_update"
	case e.ToolEnd != nil:
		return "tool_execution_end"
	case e.CompactStart != nil:
		return "auto_compaction_start"
	case e.CompactEnd != nil:
		return "auto_compaction_end"
	case e.RetryStart != nil:
		return "auto_retry_start"
	case e.RetryEnd != nil:
		return "auto_retry_end"
	case e.ExtError != nil:
		return "extension_error"
	case e.ExtUIReq != nil:
		return "extension_ui_request"
	case e.Response != nil:
		return "response"
	default:
		return "unknown"
	}
}
