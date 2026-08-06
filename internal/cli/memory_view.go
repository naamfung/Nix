package cli

import (
	"time"

	"inx/internal/control"
	"inx/internal/memory"
)

func renderMemory(width int, set *memory.Set) string {
	return viewProtectLines(control.RenderMemorySummary(set, time.Now().UTC()), width)
}
