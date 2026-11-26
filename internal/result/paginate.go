package result

// Paginate returns a 0-based page and whether more items follow it.
func Paginate(items []Value, page, size int) ([]Value, bool) {
	if size <= 0 {
		size = 100
	}
	if page < 0 {
		page = 0
	}

	start := page * size
	if start >= len(items) {
		return nil, false
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], end < len(items)
}
