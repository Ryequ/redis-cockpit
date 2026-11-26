package ui

import (
	"os"
	"runtime"
	"strings"
)

// DetectMouseSupport applies conservative terminal capability detection.
func DetectMouseSupport() bool {
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	if runtime.GOOS != "windows" {
		return true
	}

	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	if os.Getenv("WT_SESSION") != "" || termProgram == "windowsterminal" {
		return true
	}
	if strings.EqualFold(os.Getenv("ConEmuANSI"), "ON") {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	return term == "xterm" || term == "xterm-256color" || term == "screen" || term == "screen-256color"
}
