package security

import "strings"

const redacted = "***"

// RedactCommand masks authentication and password-related arguments.
func RedactCommand(argv []string) []string {
	if len(argv) == 0 {
		return nil
	}
	out := append([]string(nil), argv...)
	name := strings.ToUpper(out[0])

	switch name {
	case "AUTH":
		if len(out) > 1 {
			out[len(out)-1] = redacted
		}
	case "CONFIG":
		if len(out) >= 4 && isPasswordSetting(out[2]) {
			out[3] = redacted
		}
	case "HELLO":
		for i := 1; i < len(out)-1; i++ {
			if strings.EqualFold(out[i], "AUTH") {
				out[len(out)-1] = redacted
				break
			}
		}
	case "ACL":
		if len(out) >= 3 && strings.EqualFold(out[1], "SETUSER") {
			maskACLTokens(out, 3)
		}
	case "SETUSER":
		maskACLTokens(out, 2)
	}
	return out
}

func maskACLTokens(argv []string, start int) {
	for i := start; i < len(argv); i++ {
		token := strings.ToUpper(argv[i])
		if strings.HasPrefix(token, ">") || strings.HasPrefix(token, "#") {
			argv[i] = redacted
		}
	}
}

// RedactText masks password assignments in human-readable text.
func RedactText(text string) string {
	fields := strings.Fields(text)
	for i, field := range fields {
		if strings.HasPrefix(strings.ToLower(field), "password=") {
			fields[i] = "password=" + redacted
		}
	}
	return strings.Join(fields, " ")
}

func isPasswordSetting(name string) bool {
	name = strings.ToLower(name)
	return name == "requirepass" || name == "masterauth" || strings.Contains(name, "password") || strings.Contains(name, "auth")
}
