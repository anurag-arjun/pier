package session

import (
	"encoding/json"
	"testing"
)

func TestStateTransitions(t *testing.T) {
	s := NewSessionState()

	if s.Status != StatusIdle {
		t.Errorf("initial status: got %v, want idle", s.Status)
	}

	// agent_start → thinking
	s.ProcessEvent(Event{AgentStart: &AgentStartEvent{Type: "agent_start"}})
	if s.Status != StatusThinking {
		t.Errorf("after agent_start: got %v, want thinking", s.Status)
	}

	// message_start (assistant) → still thinking, entry created
	s.ProcessEvent(Event{MsgStart: &MessageStartEvent{
		Type: "message_start",
		Message: AgentMessage{Role: "assistant", Model: "claude-4"},
	}})
	if len(s.Timeline) != 1 || s.Timeline[0].Kind != EntryAssistantMessage {
		t.Error("expected assistant message timeline entry")
	}
	if !s.Timeline[0].Streaming {
		t.Error("assistant message should be streaming")
	}

	// text_start → streaming
	s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
		Type: "message_update",
		AssistantMessageEvent: AssistantMessageEvent{Type: "text_start"},
	}})
	if s.Status != StatusStreaming {
		t.Errorf("after text_start: got %v, want streaming", s.Status)
	}

	// text_delta → append text
	s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
		Type: "message_update",
		AssistantMessageEvent: AssistantMessageEvent{Type: "text_delta", Delta: "Hello "},
	}})
	s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
		Type: "message_update",
		AssistantMessageEvent: AssistantMessageEvent{Type: "text_delta", Delta: "world"},
	}})
	if s.Timeline[0].AssistantText != "Hello world" {
		t.Errorf("text: got %q, want 'Hello world'", s.Timeline[0].AssistantText)
	}

	// message_end → finalize message
	s.ProcessEvent(Event{MsgEnd: &MessageEndEvent{
		Type: "message_end",
		Message: AgentMessage{
			Role:    "assistant",
			Content: json.RawMessage(`[{"type":"text","text":"Hello world"}]`),
		},
	}})
	if s.Timeline[0].Streaming {
		t.Error("message should no longer be streaming")
	}

	// agent_end → waiting
	s.ProcessEvent(Event{AgentEnd: &AgentEndEvent{Type: "agent_end"}})
	if s.Status != StatusWaiting {
		t.Errorf("after agent_end: got %v, want waiting", s.Status)
	}
}

func TestToolCallTracking(t *testing.T) {
	s := NewSessionState()

	// Start a tool
	s.ProcessEvent(Event{ToolStart: &ToolExecutionStartEvent{
		Type:       "tool_execution_start",
		ToolCallID: "tc1",
		ToolName:   "read",
		Args:       json.RawMessage(`{"path":"main.go"}`),
	}})
	if len(s.Timeline) != 1 || s.Timeline[0].ToolName != "read" {
		t.Error("expected tool call entry")
	}
	if s.Timeline[0].ToolArgs != "main.go" {
		t.Errorf("tool args: got %q, want 'main.go'", s.Timeline[0].ToolArgs)
	}

	// Partial update
	s.ProcessEvent(Event{ToolUpdate: &ToolExecutionUpdateEvent{
		Type:          "tool_execution_update",
		ToolCallID:    "tc1",
		PartialResult: json.RawMessage(`"partial content"`),
	}})
	if s.Timeline[0].ToolPartial != "partial content" {
		t.Errorf("partial: got %q", s.Timeline[0].ToolPartial)
	}

	// Complete
	s.ProcessEvent(Event{ToolEnd: &ToolExecutionEndEvent{
		Type:       "tool_execution_end",
		ToolCallID: "tc1",
		ToolName:   "read",
		Result:     json.RawMessage(`"full content"`),
		IsError:    false,
	}})
	if s.Timeline[0].ToolResult != "full content" {
		t.Errorf("result: got %q", s.Timeline[0].ToolResult)
	}
	if s.Timeline[0].ToolPartial != "" {
		t.Error("partial should be cleared after end")
	}
}

func TestCompactionTransition(t *testing.T) {
	s := NewSessionState()
	s.Status = StatusStreaming

	s.ProcessEvent(Event{CompactStart: &AutoCompactionStartEvent{
		Type:   "auto_compaction_start",
		Reason: "threshold",
	}})
	if s.Status != StatusCompacting {
		t.Errorf("got %v, want compacting", s.Status)
	}

	s.ProcessEvent(Event{CompactEnd: &AutoCompactionEndEvent{Type: "auto_compaction_end"}})
	if s.Status != StatusStreaming {
		t.Errorf("got %v, want streaming (restored)", s.Status)
	}
}

func TestRetryTransition(t *testing.T) {
	s := NewSessionState()
	s.Status = StatusThinking

	s.ProcessEvent(Event{RetryStart: &AutoRetryStartEvent{
		Type:    "auto_retry_start",
		Attempt: 1, MaxAttempts: 3, DelayMs: 5000,
		ErrorMessage: "overloaded",
	}})
	if s.Status != StatusRetrying {
		t.Errorf("got %v, want retrying", s.Status)
	}

	// Success → restore
	s.ProcessEvent(Event{RetryEnd: &AutoRetryEndEvent{
		Type: "auto_retry_end", Success: true, Attempt: 1,
	}})
	if s.Status != StatusThinking {
		t.Errorf("got %v, want thinking (restored)", s.Status)
	}
}

func TestRetryFailure(t *testing.T) {
	s := NewSessionState()
	s.Status = StatusThinking

	s.ProcessEvent(Event{RetryStart: &AutoRetryStartEvent{
		Type: "auto_retry_start", Attempt: 3, MaxAttempts: 3,
	}})
	s.ProcessEvent(Event{RetryEnd: &AutoRetryEndEvent{
		Type: "auto_retry_end", Success: false, Attempt: 3,
	}})
	if s.Status != StatusError {
		t.Errorf("got %v, want error", s.Status)
	}
}

func TestThinkingDelta(t *testing.T) {
	s := NewSessionState()

	// Create an active message
	s.ProcessEvent(Event{MsgStart: &MessageStartEvent{
		Type:    "message_start",
		Message: AgentMessage{Role: "assistant"},
	}})

	s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
		Type: "message_update",
		AssistantMessageEvent: AssistantMessageEvent{Type: "thinking_start"},
	}})
	s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
		Type: "message_update",
		AssistantMessageEvent: AssistantMessageEvent{Type: "thinking_delta", Delta: "Let me think..."},
	}})
	if s.Timeline[0].ThinkingText != "Let me think..." {
		t.Errorf("thinking: got %q", s.Timeline[0].ThinkingText)
	}
}

func TestUserMessage(t *testing.T) {
	s := NewSessionState()
	s.AddUserMessage("Fix the bug")
	if len(s.Timeline) != 1 || s.Timeline[0].UserText != "Fix the bug" {
		t.Error("user message not added correctly")
	}
}

func TestResponseUpdatesModel(t *testing.T) {
	s := NewSessionState()
	s.ProcessEvent(Event{Response: &RpcResponse{
		Type:    "response",
		Command: "get_state",
		Success: true,
		Data: json.RawMessage(`{
			"model":{"id":"claude-4","name":"Claude 4","api":"anthropic","provider":"anthropic","baseUrl":"","reasoning":true,"input":["text"],"cost":{},"contextWindow":200000,"maxTokens":8192},
			"thinkingLevel":"medium",
			"isStreaming":false,
			"isCompacting":false,
			"steeringMode":"all",
			"followUpMode":"all",
			"sessionId":"s1",
			"autoCompactionEnabled":true,
			"messageCount":3,
			"pendingMessageCount":0
		}`),
	}})
	if s.Model != "Claude 4" {
		t.Errorf("model: got %q, want 'Claude 4'", s.Model)
	}
	if s.SessionID != "s1" {
		t.Errorf("sessionID: got %q", s.SessionID)
	}
}

func TestErrorTransition(t *testing.T) {
	s := NewSessionState()

	s.ProcessEvent(Event{MsgStart: &MessageStartEvent{
		Type:    "message_start",
		Message: AgentMessage{Role: "assistant"},
	}})
	s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
		Type: "message_update",
		AssistantMessageEvent: AssistantMessageEvent{
			Type:   "error",
			Reason: "error",
		},
	}})
	if s.Status != StatusError {
		t.Errorf("got %v, want error", s.Status)
	}
}
