package session

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Status represents the derived UI state of a session.
type Status int

const (
	StatusIdle       Status = iota // no session running
	StatusThinking                 // between agent_start and first output
	StatusStreaming                 // between text_start and text_end
	StatusWaiting                  // after agent_end, ready for input
	StatusError                    // error occurred
	StatusCompacting               // auto-compaction in progress
	StatusRetrying                 // auto-retry in progress
)

func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusThinking:
		return "thinking"
	case StatusStreaming:
		return "streaming"
	case StatusWaiting:
		return "waiting"
	case StatusError:
		return "error"
	case StatusCompacting:
		return "compacting"
	case StatusRetrying:
		return "retrying"
	default:
		return "unknown"
	}
}

// TimelineEntryKind discriminates timeline entries.
type TimelineEntryKind int

const (
	EntryUserMessage TimelineEntryKind = iota
	EntryAssistantMessage
	EntryToolCall
	EntryCompactionNotice
	EntryRetryNotice
)

// TimelineEntry is one item in the session timeline.
type TimelineEntry struct {
	Kind TimelineEntryKind

	// EntryUserMessage
	UserText string

	// EntryAssistantMessage
	AssistantText string
	ThinkingText  string
	Streaming     bool   // true while receiving text_delta
	Model         string // model that generated this message

	// EntryToolCall
	ToolCallID    string
	ToolName      string
	ToolArgs      string // human-readable args summary
	ToolResult    string // final result text
	ToolIsError   bool
	ToolPartial   string // accumulated partial result during streaming
	ToolExpanded  bool   // UI state: is the tool block expanded?

	// EntryCompactionNotice
	CompactReason string

	// EntryRetryNotice
	RetryAttempt    int
	RetryMaxAttempts int
	RetryError      string
}

// SessionState holds derived UI state from RPC events.
type SessionState struct {
	mu sync.Mutex

	Status        Status
	prevStatus    Status // status before compacting/retrying
	Model         string
	SessionFile   string
	SessionID     string
	Timeline      []TimelineEntry
	activeMessage *TimelineEntry // pointer into Timeline for streaming message
	toolCalls     map[string]int // toolCallId → Timeline index
}

// Lock acquires the state mutex for external readers (e.g., UI rendering).
func (s *SessionState) Lock() { s.mu.Lock() }

// Unlock releases the state mutex.
func (s *SessionState) Unlock() { s.mu.Unlock() }

// NewSessionState creates an empty session state.
func NewSessionState() *SessionState {
	return &SessionState{
		Status:    StatusIdle,
		toolCalls: make(map[string]int),
	}
}

// ProcessEvent updates state from an RPC event. Returns true if UI should redraw.
func (s *SessionState) ProcessEvent(evt Event) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case evt.AgentStart != nil:
		s.Status = StatusThinking
		return true

	case evt.AgentEnd != nil:
		s.Status = StatusWaiting
		s.activeMessage = nil
		return true

	case evt.TurnStart != nil:
		// No state change needed
		return false

	case evt.TurnEnd != nil:
		return false

	case evt.MsgStart != nil:
		msg := evt.MsgStart.Message
		if msg.Role == "assistant" {
			entry := TimelineEntry{
				Kind:      EntryAssistantMessage,
				Streaming: true,
				Model:     msg.Model,
			}
			s.Timeline = append(s.Timeline, entry)
			s.activeMessage = &s.Timeline[len(s.Timeline)-1]
		}
		return true

	case evt.MsgUpdate != nil:
		return s.processMessageUpdate(evt.MsgUpdate)

	case evt.MsgEnd != nil:
		if s.activeMessage != nil {
			s.activeMessage.Streaming = false
			// Extract final text from the message content
			s.activeMessage.AssistantText = extractText(evt.MsgEnd.Message.Content)
			s.activeMessage.ThinkingText = extractThinking(evt.MsgEnd.Message.Content)
			s.activeMessage = nil
		}
		return true

	case evt.ToolStart != nil:
		entry := TimelineEntry{
			Kind:       EntryToolCall,
			ToolCallID: evt.ToolStart.ToolCallID,
			ToolName:   evt.ToolStart.ToolName,
			ToolArgs:   summarizeArgs(evt.ToolStart.Args),
		}
		s.Timeline = append(s.Timeline, entry)
		s.toolCalls[evt.ToolStart.ToolCallID] = len(s.Timeline) - 1
		return true

	case evt.ToolUpdate != nil:
		if idx, ok := s.toolCalls[evt.ToolUpdate.ToolCallID]; ok {
			s.Timeline[idx].ToolPartial = stringifyRaw(evt.ToolUpdate.PartialResult)
		}
		return true

	case evt.ToolEnd != nil:
		if idx, ok := s.toolCalls[evt.ToolEnd.ToolCallID]; ok {
			s.Timeline[idx].ToolResult = stringifyRaw(evt.ToolEnd.Result)
			s.Timeline[idx].ToolIsError = evt.ToolEnd.IsError
			s.Timeline[idx].ToolPartial = "" // clear partial
		}
		return true

	case evt.CompactStart != nil:
		s.prevStatus = s.Status
		s.Status = StatusCompacting
		s.Timeline = append(s.Timeline, TimelineEntry{
			Kind:          EntryCompactionNotice,
			CompactReason: evt.CompactStart.Reason,
		})
		return true

	case evt.CompactEnd != nil:
		s.Status = s.prevStatus
		return true

	case evt.RetryStart != nil:
		s.prevStatus = s.Status
		s.Status = StatusRetrying
		s.Timeline = append(s.Timeline, TimelineEntry{
			Kind:             EntryRetryNotice,
			RetryAttempt:     evt.RetryStart.Attempt,
			RetryMaxAttempts: evt.RetryStart.MaxAttempts,
			RetryError:       evt.RetryStart.ErrorMessage,
		})
		return true

	case evt.RetryEnd != nil:
		if evt.RetryEnd.Success {
			s.Status = s.prevStatus
		} else {
			s.Status = StatusError
		}
		return true

	case evt.ExtError != nil:
		return true

	case evt.ExtUIReq != nil:
		return true

	case evt.Response != nil:
		return s.processResponse(evt.Response)
	}

	return false
}

func (s *SessionState) processMessageUpdate(evt *MessageUpdateEvent) bool {
	ame := evt.AssistantMessageEvent
	switch ame.Type {
	case "text_start":
		s.Status = StatusStreaming
		return true
	case "text_delta":
		if s.activeMessage != nil {
			s.activeMessage.AssistantText += ame.Delta
		}
		return true
	case "text_end":
		// Full text will be set from message_end
		return true
	case "thinking_start":
		s.Status = StatusThinking
		return true
	case "thinking_delta":
		if s.activeMessage != nil {
			s.activeMessage.ThinkingText += ame.Delta
		}
		return true
	case "thinking_end":
		return true
	case "toolcall_start", "toolcall_delta", "toolcall_end":
		return false // tool events handle this
	case "done":
		return true
	case "error":
		s.Status = StatusError
		return true
	}
	return false
}

func (s *SessionState) processResponse(resp *RpcResponse) bool {
	if !resp.Success {
		return false
	}
	switch resp.Command {
	case "get_state":
		var state RpcSessionState
		if err := json.Unmarshal(resp.Data, &state); err == nil {
			if state.Model != nil {
				s.Model = state.Model.Name
			}
			s.SessionFile = state.SessionFile
			s.SessionID = state.SessionID
		}
		return true
	}
	return false
}

// AddUserMessage appends a user message to the timeline.
func (s *SessionState) AddUserMessage(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Timeline = append(s.Timeline, TimelineEntry{
		Kind:     EntryUserMessage,
		UserText: text,
	})
}

// --- Helpers ---

// extractText pulls text content from an AgentMessage.Content JSON array.
func extractText(content json.RawMessage) string {
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(content, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// extractThinking pulls thinking content from an AgentMessage.Content JSON array.
func extractThinking(content json.RawMessage) string {
	var blocks []struct {
		Type     string `json:"type"`
		Thinking string `json:"thinking"`
	}
	if err := json.Unmarshal(content, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "thinking" && b.Thinking != "" {
			parts = append(parts, b.Thinking)
		}
	}
	return strings.Join(parts, "\n")
}

// summarizeArgs creates a short human-readable summary of tool args.
func summarizeArgs(raw json.RawMessage) string {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return string(raw)
	}
	// For common tools, show the most relevant arg
	if path, ok := m["path"]; ok {
		return fmt.Sprintf("%v", path)
	}
	if command, ok := m["command"]; ok {
		cmd := fmt.Sprintf("%v", command)
		if len(cmd) > 80 {
			cmd = cmd[:80] + "..."
		}
		return cmd
	}
	// Fallback: compact JSON
	b, _ := json.Marshal(m)
	s := string(b)
	if len(s) > 100 {
		s = s[:100] + "..."
	}
	return s
}

// stringifyRaw converts a json.RawMessage to a display string.
func stringifyRaw(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// Try as string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// Otherwise compact JSON
	return string(raw)
}
