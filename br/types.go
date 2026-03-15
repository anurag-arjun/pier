package br

// Task represents a br issue.
type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	Status      string   `json:"status"`       // "open"|"in_progress"|"closed"|"deferred"
	Type        string   `json:"type"`          // "task"|"feature"|"bug"|"epic"
	Tags        []string `json:"tags,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
	ClosedAt    string   `json:"closed_at,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}
