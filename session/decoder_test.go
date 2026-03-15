package session

import (
	"strings"
	"testing"
)

func TestDecodeAgentLifecycle(t *testing.T) {
	input := `{"type":"agent_start"}
{"type":"turn_start"}
{"type":"message_start","message":{"role":"assistant","content":[],"timestamp":1234}}
{"type":"message_update","message":{"role":"assistant","content":[],"timestamp":1234},"assistantMessageEvent":{"type":"text_delta","delta":"Hello","contentIndex":0}}
{"type":"message_end","message":{"role":"assistant","content":[],"timestamp":1234}}
{"type":"turn_end","message":{"role":"assistant","content":[],"timestamp":1234},"toolResults":[]}
{"type":"agent_end","messages":[]}
`
	ch := DecodeEvents(strings.NewReader(input))

	expected := []string{
		"agent_start", "turn_start", "message_start",
		"message_update", "message_end", "turn_end", "agent_end",
	}

	for _, want := range expected {
		evt, ok := <-ch
		if !ok {
			t.Fatalf("channel closed early, expected %s", want)
		}
		got := evt.EventType()
		if got != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	}
}

func TestDecodeToolEvents(t *testing.T) {
	input := `{"type":"tool_execution_start","toolCallId":"tc1","toolName":"read","args":{"path":"foo.go"}}
{"type":"tool_execution_update","toolCallId":"tc1","toolName":"read","args":{},"partialResult":"line1\nline2"}
{"type":"tool_execution_end","toolCallId":"tc1","toolName":"read","result":{"content":"full"},"isError":false}
`
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.ToolStart == nil || evt.ToolStart.ToolName != "read" || evt.ToolStart.ToolCallID != "tc1" {
		t.Error("tool_execution_start mismatch")
	}

	evt = <-ch
	if evt.ToolUpdate == nil || evt.ToolUpdate.ToolCallID != "tc1" {
		t.Error("tool_execution_update mismatch")
	}

	evt = <-ch
	if evt.ToolEnd == nil || evt.ToolEnd.ToolCallID != "tc1" || evt.ToolEnd.IsError {
		t.Error("tool_execution_end mismatch")
	}
}

func TestDecodeStreamingDeltas(t *testing.T) {
	input := `{"type":"message_update","message":{"role":"assistant","content":[],"timestamp":1},"assistantMessageEvent":{"type":"text_start","contentIndex":0}}
{"type":"message_update","message":{"role":"assistant","content":[],"timestamp":1},"assistantMessageEvent":{"type":"text_delta","contentIndex":0,"delta":"Hello "}}
{"type":"message_update","message":{"role":"assistant","content":[],"timestamp":1},"assistantMessageEvent":{"type":"text_delta","contentIndex":0,"delta":"world"}}
{"type":"message_update","message":{"role":"assistant","content":[],"timestamp":1},"assistantMessageEvent":{"type":"text_end","contentIndex":0,"content":"Hello world"}}
`
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.MsgUpdate == nil || evt.MsgUpdate.AssistantMessageEvent.Type != "text_start" {
		t.Error("expected text_start")
	}

	evt = <-ch
	if evt.MsgUpdate.AssistantMessageEvent.Delta != "Hello " {
		t.Errorf("expected delta 'Hello ', got %q", evt.MsgUpdate.AssistantMessageEvent.Delta)
	}

	evt = <-ch
	if evt.MsgUpdate.AssistantMessageEvent.Delta != "world" {
		t.Errorf("expected delta 'world', got %q", evt.MsgUpdate.AssistantMessageEvent.Delta)
	}

	evt = <-ch
	if evt.MsgUpdate.AssistantMessageEvent.Content != "Hello world" {
		t.Errorf("expected content 'Hello world', got %q", evt.MsgUpdate.AssistantMessageEvent.Content)
	}
}

func TestDecodeSessionEvents(t *testing.T) {
	input := `{"type":"auto_compaction_start","reason":"threshold"}
{"type":"auto_compaction_end","result":{"summary":"compacted"},"aborted":false,"willRetry":false}
{"type":"auto_retry_start","attempt":1,"maxAttempts":3,"delayMs":5000,"errorMessage":"overloaded"}
{"type":"auto_retry_end","success":true,"attempt":1}
`
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.CompactStart == nil || evt.CompactStart.Reason != "threshold" {
		t.Error("compaction_start mismatch")
	}

	evt = <-ch
	if evt.CompactEnd == nil || evt.CompactEnd.Aborted {
		t.Error("compaction_end mismatch")
	}

	evt = <-ch
	if evt.RetryStart == nil || evt.RetryStart.Attempt != 1 {
		t.Error("retry_start mismatch")
	}

	evt = <-ch
	if evt.RetryEnd == nil || !evt.RetryEnd.Success {
		t.Error("retry_end mismatch")
	}
}

func TestDecodeExtensionEvents(t *testing.T) {
	input := `{"type":"extension_ui_request","id":"ext1","method":"confirm","title":"Delete?","message":"Are you sure?"}
{"type":"extension_error","extensionPath":"/foo/bar.ts","event":"onLoad","error":"boom"}
`
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.ExtUIReq == nil || evt.ExtUIReq.Method != "confirm" || evt.ExtUIReq.Title != "Delete?" {
		t.Error("extension_ui_request mismatch")
	}

	evt = <-ch
	if evt.ExtError == nil || evt.ExtError.Error != "boom" {
		t.Error("extension_error mismatch")
	}
}

func TestDecodeRpcResponse(t *testing.T) {
	input := `{"id":"cmd1","type":"response","command":"get_state","success":true,"data":{"thinkingLevel":"medium","isStreaming":false,"isCompacting":false,"steeringMode":"all","followUpMode":"all","sessionId":"abc","autoCompactionEnabled":true,"messageCount":5,"pendingMessageCount":0}}
{"id":"cmd2","type":"response","command":"prompt","success":false,"error":"agent is busy"}
`
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.Response == nil || evt.Response.Command != "get_state" || !evt.Response.Success {
		t.Error("response success mismatch")
	}

	evt = <-ch
	if evt.Response == nil || evt.Response.Success || evt.Response.Error != "agent is busy" {
		t.Error("response error mismatch")
	}
}

func TestDecodeUnicodeLineseparator(t *testing.T) {
	// U+2028 LINE SEPARATOR inside a JSON string value — must NOT split the line
	input := "{\"type\":\"message_update\",\"message\":{\"role\":\"assistant\",\"content\":[],\"timestamp\":1},\"assistantMessageEvent\":{\"type\":\"text_delta\",\"delta\":\"before\\u2028after\",\"contentIndex\":0}}\n"
	ch := DecodeEvents(strings.NewReader(input))

	evt, ok := <-ch
	if !ok {
		t.Fatal("channel closed unexpectedly")
	}
	if evt.MsgUpdate == nil {
		t.Fatal("expected message_update")
	}
	if evt.MsgUpdate.AssistantMessageEvent.Delta != "before\u2028after" {
		t.Errorf("delta mismatch: got %q", evt.MsgUpdate.AssistantMessageEvent.Delta)
	}
}

func TestDecodeMalformedLines(t *testing.T) {
	input := "not json at all\n{\"type\":\"agent_start\"}\n{broken\n"
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.AgentStart == nil {
		t.Error("expected agent_start after malformed lines")
	}

	// Channel should close after all lines processed
	_, ok := <-ch
	if ok {
		t.Error("expected channel to close")
	}
}

func TestDecodeEmptyLines(t *testing.T) {
	input := "\n\n{\"type\":\"agent_start\"}\n\n"
	ch := DecodeEvents(strings.NewReader(input))

	evt := <-ch
	if evt.AgentStart == nil {
		t.Error("expected agent_start")
	}
}
