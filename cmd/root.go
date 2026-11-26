package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/Ryequ/redis-cockpit/internal/app"
)

// Options contains command-line startup settings.
type Options struct {
	ConfigPath string
	Profile    string
	ReadOnly   bool
	Debug      bool
}

// ParseArgs parses CLI flags without writing to stderr, which keeps tests and
// embedding callers predictable.
func ParseArgs(args []string) (Options, error) {
	var opts Options
	fs := flag.NewFlagSet("redis-cockpit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.ConfigPath, "config", "", "path to config YAML file")
	fs.StringVar(&opts.Profile, "profile", "", "connection profile name to open at startup")
	fs.BoolVar(&opts.ReadOnly, "readonly", false, "disable write and dangerous commands")
	fs.BoolVar(&opts.Debug, "debug", false, "enable debug logging")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	return opts, nil
}

// Execute parses arguments and runs the TUI.
func Execute(args []string) error {
	opts, err := ParseArgs(args)
	if err != nil {
		return fmt.Errorf("parse arguments: %w", err)
	}

	return app.Run(app.Options{
		ConfigPath: opts.ConfigPath,
		Profile:    opts.Profile,
		ReadOnly:   opts.ReadOnly,
		Debug:      opts.Debug,
	})
}
