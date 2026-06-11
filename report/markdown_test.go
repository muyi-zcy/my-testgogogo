package report

import (
	"strings"
	"testing"
	"time"
)

func TestRenderStepDetailsShowsInput(t *testing.T) {
	steps := []StepRecord{
		{
			Index:      1,
			Name:       "login",
			Status:     "PASS",
			StartedAt:  time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2026, 6, 11, 9, 0, 0, 29_000_000, time.UTC),
			Duration:   "29ms",
			DurationMs: 29,
			Input: map[string]any{
				"username": "admin",
				"type":     "app",
			},
			Result: map[string]any{
				"code": 0,
			},
		},
	}

	out := renderStepDetails(steps)
	if !strings.Contains(out, "**入参：**") {
		t.Fatalf("expected input section, got: %s", out)
	}
	if !strings.Contains(out, "username") || !strings.Contains(out, "admin") {
		t.Fatalf("expected input fields in output, got: %s", out)
	}
	if !strings.Contains(out, "**结构化结果：**") {
		t.Fatalf("expected result section, got: %s", out)
	}
}
