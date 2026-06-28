package commands

import (
	"fmt"
	"strings"

	"dockerToolBox/internal/engine"
)

// shortID returns the 12-character short form of a Docker ID.
func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// containerLabel renders a container for display.
func containerLabel(c engine.Container) string {
	if c.Name != "" {
		return fmt.Sprintf("%s (%s)", c.Name, c.Image)
	}
	return shortID(c.ID)
}

func containerLabels(cs []engine.Container) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = containerLabel(c)
	}
	return out
}

// imageLabel renders an image for display.
func imageLabel(im engine.Image) string {
	if len(im.Tags) > 0 {
		return strings.Join(im.Tags, ", ")
	}
	return shortID(im.ID)
}

func imageLabels(ims []engine.Image) []string {
	out := make([]string, len(ims))
	for i, im := range ims {
		out[i] = imageLabel(im)
	}
	return out
}

// nameOr returns name, or "-" when it is empty (for table cells).
func nameOr(name string) string {
	if name == "" {
		return "-"
	}
	return name
}

// humanSizeI formats a signed byte count, treating negatives as zero.
func humanSizeI(b int64) string {
	if b < 0 {
		b = 0
	}
	return humanSize(uint64(b))
}

// humanSize formats a byte count as a human-readable string.
func humanSize(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
