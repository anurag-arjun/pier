package session

import (
	"strings"
	"sync"
	"testing"
)

// TestConcurrentDecoders verifies 5 concurrent JSONL decoders don't race or leak.
func TestConcurrentDecoders(t *testing.T) {
	input := `{"type":"agent_start"}
{"type":"message_start","message":{"role":"assistant","content":[],"timestamp":1}}
{"type":"message_update","message":{"role":"assistant","content":[],"timestamp":1},"assistantMessageEvent":{"type":"text_delta","delta":"Hello","contentIndex":0}}
{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}],"timestamp":1}}
{"type":"tool_execution_start","toolCallId":"tc1","toolName":"read","args":{"path":"main.go"}}
{"type":"tool_execution_end","toolCallId":"tc1","toolName":"read","result":"content","isError":false}
{"type":"agent_end","messages":[]}
`

	const numDecoders = 5
	var wg sync.WaitGroup
	wg.Add(numDecoders)

	for d := 0; d < numDecoders; d++ {
		go func() {
			defer wg.Done()
			ch := DecodeEvents(strings.NewReader(input))
			count := 0
			for range ch {
				count++
			}
			if count != 7 {
				t.Errorf("expected 7 events, got %d", count)
			}
		}()
	}

	wg.Wait()
}

// TestLongTimeline verifies the state machine handles 100+ entries without issues.
func TestLongTimeline(t *testing.T) {
	s := NewSessionState()

	// Simulate 50 prompt/response cycles
	for i := 0; i < 50; i++ {
		s.AddUserMessage("prompt")
		s.ProcessEvent(Event{AgentStart: &AgentStartEvent{Type: "agent_start"}})
		s.ProcessEvent(Event{MsgStart: &MessageStartEvent{
			Type:    "message_start",
			Message: AgentMessage{Role: "assistant"},
		}})
		s.ProcessEvent(Event{MsgUpdate: &MessageUpdateEvent{
			Type: "message_update",
			AssistantMessageEvent: AssistantMessageEvent{Type: "text_delta", Delta: "response "},
		}})
		s.ProcessEvent(Event{MsgEnd: &MessageEndEvent{
			Type:    "message_end",
			Message: AgentMessage{Role: "assistant", Content: []byte(`[{"type":"text","text":"response"}]`)},
		}})
		s.ProcessEvent(Event{AgentEnd: &AgentEndEvent{Type: "agent_end"}})
	}

	s.Lock()
	defer s.Unlock()

	// 50 user + 50 assistant = 100 entries
	if len(s.Timeline) != 100 {
		t.Errorf("expected 100 timeline entries, got %d", len(s.Timeline))
	}
	if s.Status != StatusWaiting {
		t.Errorf("expected waiting, got %v", s.Status)
	}
}

// TestConcurrentStateAccess verifies Lock/Unlock works under concurrent access.
func TestConcurrentStateAccess(t *testing.T) {
	s := NewSessionState()

	var wg sync.WaitGroup

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.ProcessEvent(Event{AgentStart: &AgentStartEvent{Type: "agent_start"}})
			s.ProcessEvent(Event{AgentEnd: &AgentEndEvent{Type: "agent_end"}})
		}
	}()

	// Reader goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.Lock()
			_ = s.Status
			_ = len(s.Timeline)
			s.Unlock()
		}
	}()

	wg.Wait()
}
