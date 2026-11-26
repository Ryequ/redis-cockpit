package command

import (
	"fmt"
	"strings"
)

// Parse splits a command line into Redis arguments with shell-like quoting.
func Parse(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	hasToken := false

	flush := func() {
		if hasToken {
			args = append(args, current.String())
			current.Reset()
			hasToken = false
		}
	}

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\\' {
			if i+1 >= len(runes) {
				return nil, fmt.Errorf("dangling escape at position %d", i)
			}
			i++
			current.WriteRune(runes[i])
			hasToken = true
			continue
		}

		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
			hasToken = true
			continue
		}

		switch r {
		case '\'', '"':
			quote = r
			hasToken = true
		case ' ', '\t', '\n', '\r':
			flush()
		default:
			current.WriteRune(r)
			hasToken = true
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote %q", string(quote))
	}
	flush()
	return args, nil
}
