// Package ui provides small, dependency-free helpers for styled
// terminal output: colors and the status symbols used across the tool.
//
// The interactive terminal UI (Bubble Tea / Lip Gloss) is planned for a
// later version; v0.1 keeps output to plain ANSI so the binary stays
// lean and builds with no UI dependencies.
package ui

import (
	"fmt"
	"os"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	gray   = "\033[90m"
)

// noColor honors the NO_COLOR convention (https://no-color.org).
var noColor = func() bool { _, ok := os.LookupEnv("NO_COLOR"); return ok }()

func colorize(code, s string) string {
	if noColor {
		return s
	}
	return code + s + reset
}

func render(format string, a []any) string {
	if len(a) == 0 {
		return format
	}
	return fmt.Sprintf(format, a...)
}

// Header prints a bold section heading.
func Header(format string, a ...any) {
	fmt.Println(colorize(bold, render(format, a)))
}

// Success prints a green check line.
func Success(format string, a ...any) {
	fmt.Println(colorize(green, "✓ ") + render(format, a))
}

// Warn prints a yellow warning line.
func Warn(format string, a ...any) {
	fmt.Println(colorize(yellow, "⚠ ") + render(format, a))
}

// Fail prints a red failure line to stderr.
func Fail(format string, a ...any) {
	fmt.Fprintln(os.Stderr, colorize(red, "✗ ")+render(format, a))
}

// Infof prints a blue information line.
func Infof(format string, a ...any) {
	fmt.Println(colorize(blue, "ℹ ") + render(format, a))
}

// Step prints an indented bullet, used for listing affected items and
// dry-run "would" lines.
func Step(format string, a ...any) {
	fmt.Println(colorize(gray, "  • ") + render(format, a))
}
