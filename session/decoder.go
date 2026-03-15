package session

import (
	"encoding/json"
	"io"
	"log"
)

// DecodeEvents reads JSONL from r and sends parsed Events to the returned channel.
// Splits strictly on \n (0x0A) — not on Unicode line separators (U+2028, U+2029)
// which are valid inside JSON strings. Closes the channel when r returns io.EOF or an error.
func DecodeEvents(r io.Reader) <-chan Event {
	ch := make(chan Event, 64)
	go func() {
		defer close(ch)
		scanner := newByteScanner(r)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			evt, err := parseEvent(line)
			if err != nil {
				log.Printf("jsonl: skip malformed line: %v (first 200 bytes: %s)", err, truncate(line, 200))
				continue
			}
			ch <- evt
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			log.Printf("jsonl: reader error: %v", err)
		}
	}()
	return ch
}

// parseEvent decodes a single JSON line into a typed Event.
func parseEvent(line []byte) (Event, error) {
	// First pass: extract the type field.
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &raw); err != nil {
		return Event{}, err
	}

	var evt Event
	var err error

	switch raw.Type {
	case "agent_start":
		var e AgentStartEvent
		err = json.Unmarshal(line, &e)
		evt.AgentStart = &e
	case "agent_end":
		var e AgentEndEvent
		err = json.Unmarshal(line, &e)
		evt.AgentEnd = &e
	case "turn_start":
		var e TurnStartEvent
		err = json.Unmarshal(line, &e)
		evt.TurnStart = &e
	case "turn_end":
		var e TurnEndEvent
		err = json.Unmarshal(line, &e)
		evt.TurnEnd = &e
	case "message_start":
		var e MessageStartEvent
		err = json.Unmarshal(line, &e)
		evt.MsgStart = &e
	case "message_update":
		var e MessageUpdateEvent
		err = json.Unmarshal(line, &e)
		evt.MsgUpdate = &e
	case "message_end":
		var e MessageEndEvent
		err = json.Unmarshal(line, &e)
		evt.MsgEnd = &e
	case "tool_execution_start":
		var e ToolExecutionStartEvent
		err = json.Unmarshal(line, &e)
		evt.ToolStart = &e
	case "tool_execution_update":
		var e ToolExecutionUpdateEvent
		err = json.Unmarshal(line, &e)
		evt.ToolUpdate = &e
	case "tool_execution_end":
		var e ToolExecutionEndEvent
		err = json.Unmarshal(line, &e)
		evt.ToolEnd = &e
	case "auto_compaction_start":
		var e AutoCompactionStartEvent
		err = json.Unmarshal(line, &e)
		evt.CompactStart = &e
	case "auto_compaction_end":
		var e AutoCompactionEndEvent
		err = json.Unmarshal(line, &e)
		evt.CompactEnd = &e
	case "auto_retry_start":
		var e AutoRetryStartEvent
		err = json.Unmarshal(line, &e)
		evt.RetryStart = &e
	case "auto_retry_end":
		var e AutoRetryEndEvent
		err = json.Unmarshal(line, &e)
		evt.RetryEnd = &e
	case "extension_error":
		var e ExtensionErrorEvent
		err = json.Unmarshal(line, &e)
		evt.ExtError = &e
	case "extension_ui_request":
		var e ExtensionUIRequest
		err = json.Unmarshal(line, &e)
		evt.ExtUIReq = &e
	case "response":
		var e RpcResponse
		err = json.Unmarshal(line, &e)
		evt.Response = &e
	default:
		// Unknown event type — log but don't fail
		log.Printf("jsonl: unknown event type %q", raw.Type)
		return Event{}, nil
	}

	return evt, err
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

// byteScanner splits on \n (0x0A) only, ignoring Unicode line separators.
// bufio.Scanner's default ScanLines splits on \r\n and \n, which is fine,
// but this implementation is explicit about not splitting on U+2028/U+2029.
type byteScanner struct {
	r    io.Reader
	buf  []byte
	start int
	end  int
	err  error
	line []byte
}

const initialBufSize = 64 * 1024

func newByteScanner(r io.Reader) *byteScanner {
	buf := make([]byte, initialBufSize)
	return &byteScanner{
		r:   r,
		buf: buf,
	}
}

func (s *byteScanner) Scan() bool {
	for {
		// Look for \n in buffered data
		for i := s.start; i < s.end; i++ {
			if s.buf[i] == '\n' {
				s.line = s.buf[s.start:i]
				// Strip trailing \r if present
				if len(s.line) > 0 && s.line[len(s.line)-1] == '\r' {
					s.line = s.line[:len(s.line)-1]
				}
				s.start = i + 1
				return true
			}
		}

		// No newline found — need more data
		if s.err != nil {
			// Return remaining data as final line if non-empty
			if s.start < s.end {
				s.line = s.buf[s.start:s.end]
				if len(s.line) > 0 && s.line[len(s.line)-1] == '\r' {
					s.line = s.line[:len(s.line)-1]
				}
				s.start = s.end
				return len(s.line) > 0
			}
			return false
		}

		// Compact buffer: move unprocessed data to front
		if s.start > 0 {
			n := copy(s.buf, s.buf[s.start:s.end])
			s.end = n
			s.start = 0
		}

		// Grow buffer if full
		if s.end >= len(s.buf) {
			newBuf := make([]byte, len(s.buf)*2)
			copy(newBuf, s.buf[:s.end])
			s.buf = newBuf
		}

		// Read more data into remaining space
		n, err := s.r.Read(s.buf[s.end:])
		s.end += n
		if err != nil {
			s.err = err
		}
	}
}

func (s *byteScanner) Bytes() []byte {
	return s.line
}

func (s *byteScanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
