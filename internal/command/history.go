package command

import "strings"

// History stores recent commands with a bounded size.
type History struct {
	limit   int
	entries []string
}

func NewHistory(limit int) *History {
	if limit <= 0 {
		limit = 100
	}
	return &History{limit: limit}
}

func (h *History) Add(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	for i, entry := range h.entries {
		if entry == line {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}

	h.entries = append(h.entries, line)
	if len(h.entries) > h.limit {
		h.entries = h.entries[len(h.entries)-h.limit:]
	}
}

func (h *History) All() []string {
	result := make([]string, len(h.entries))
	copy(result, h.entries)
	return result
}

func (h *History) Clear() {
	h.entries = nil
}
