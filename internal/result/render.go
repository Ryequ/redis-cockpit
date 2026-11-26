package result

import (
	"strconv"
	"strings"
)

// Format controls how a Value is rendered in the UI.
type Format string

const (
	FormatRaw   Format = "raw"
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func Render(value Value, format Format) string {
	switch format {
	case FormatJSON:
		return renderJSON(value)
	case FormatTable:
		return renderTable(value)
	default:
		return renderRaw(value, 0)
	}
}

func renderRaw(value Value, depth int) string {
	indent := strings.Repeat("  ", depth)
	switch value.Kind {
	case KindArray:
		var lines []string
		for _, item := range value.Items {
			lines = append(lines, indent+"- "+strings.TrimPrefix(renderRaw(item, depth+1), indent+"  "))
		}
		return strings.Join(lines, "\n")
	case KindMap:
		var lines []string
		for _, pair := range value.Pairs {
			lines = append(lines, indent+pair.Key+": "+strings.TrimPrefix(renderRaw(pair.Value, depth+1), indent+"  "))
		}
		return strings.Join(lines, "\n")
	default:
		return value.Display()
	}
}

func renderJSON(value Value) string {
	switch value.Kind {
	case KindArray:
		parts := make([]string, len(value.Items))
		for i, item := range value.Items {
			parts[i] = renderJSON(item)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case KindMap:
		parts := make([]string, len(value.Pairs))
		for i, pair := range value.Pairs {
			parts[i] = strconv.Quote(pair.Key) + ":" + renderJSON(pair.Value)
		}
		return "{" + strings.Join(parts, ",") + "}"
	case KindString:
		return strconv.Quote(value.Text)
	case KindInt:
		return strconv.FormatInt(value.Int, 10)
	case KindFloat:
		return strconv.FormatFloat(value.Float, 'f', -1, 64)
	case KindBool:
		return strconv.FormatBool(value.Bool)
	default:
		return "null"
	}
}

func renderTable(value Value) string {
	if value.Kind != KindArray {
		return "VALUE\n" + value.Display()
	}

	headers := []string{"VALUE"}
	rows := make([][]string, 0, len(value.Items))
	for _, item := range value.Items {
		if item.Kind == KindMap {
			headers = headers[:0]
			for _, pair := range item.Pairs {
				headers = append(headers, pair.Key)
			}
			row := make([]string, len(item.Pairs))
			for i, pair := range item.Pairs {
				row[i] = pair.Value.Display()
			}
			rows = append(rows, row)
			continue
		}
		rows = append(rows, []string{item.Display()})
	}

	var b strings.Builder
	b.WriteString(strings.Join(headers, " | "))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", maxTableWidth(headers)))
	for _, row := range rows {
		b.WriteString("\n")
		b.WriteString(strings.Join(row, " | "))
	}
	return b.String()
}

func maxTableWidth(headers []string) int {
	width := 0
	for _, header := range headers {
		width += len(header) + 3
	}
	if width == 0 {
		return 5
	}
	return width - 3
}
