package report

import (
	"strings"
	"testing"
)

func TestCollectorSetResponseOnFailure(t *testing.T) {
	t.Parallel()

	r := Enable(t, Meta{Generate: true, Standalone: true, Title: "failure response"})
	r.Step("call api", func(t *testing.T) {
		r.SetInput(map[string]any{"id": "1"})
		r.SetResponse(map[string]any{
			"code":    500,
			"message": "internal error",
			"success": false,
		})
		t.Fatalf("assert failed")
	})

	c := r.(*Collector)
	if len(c.steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(c.steps))
	}
	step := c.steps[0]
	if step.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", step.Status)
	}
	if step.Response["code"] != 500 {
		t.Fatalf("expected response code 500, got %v", step.Response["code"])
	}
	if len(step.Result) != 0 {
		t.Fatalf("expected empty result on failure-only capture, got %v", step.Result)
	}
}

func TestCollectorSetResponseSkippedOnPass(t *testing.T) {
	t.Parallel()

	r := Enable(t, Meta{Generate: true, Standalone: true, Title: "pass response"})
	r.Step("call api", func(t *testing.T) {
		r.SetResponse(map[string]any{"code": 200, "success": true})
		r.SetResult(map[string]any{"ok": true})
	})

	c := r.(*Collector)
	step := c.steps[0]
	if step.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", step.Status)
	}
	if len(step.Response) != 0 {
		t.Fatalf("expected no response on pass, got %v", step.Response)
	}
}

func TestRenderStepDetailsShowsResponseOnFailure(t *testing.T) {
	steps := []StepRecord{
		{
			Index:  1,
			Name:   "login",
			Status: "FAIL",
			Detail: "step failed",
			Input: map[string]any{
				"username": "admin",
			},
			Response: map[string]any{
				"code":    401,
				"message": "bad credentials",
				"success": false,
			},
		},
	}

	out := renderStepDetails(steps)
	if !strings.Contains(out, "**接口响应：**") {
		t.Fatalf("expected response section, got: %s", out)
	}
	if !strings.Contains(out, "bad credentials") {
		t.Fatalf("expected response body in output, got: %s", out)
	}
}
