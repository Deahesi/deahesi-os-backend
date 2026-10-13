package aihandlers

// Keep the existing SSE wire format independent of the model provider.
type streamChoice struct {
	Delta        streamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
	Index        int         `json:"index"`
}

type streamDelta struct {
	Content   string `json:"content,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
	Role      string `json:"role,omitempty"`
}

func finishReason(reason string) string {
	switch reason {
	case "blocked":
		return "content_filter"
	case "interrupted":
		return "tool_calls"
	default:
		return reason
	}
}
