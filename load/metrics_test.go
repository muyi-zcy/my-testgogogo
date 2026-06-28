package load

import (
	"strings"
	"testing"
	"time"
)

func TestMetricsTimeSeries(t *testing.T) {
	m := NewMetrics(time.Second)
	m.Record(10*time.Millisecond, nil)
	m.Record(20*time.Millisecond, nil)
	m.RecordCustom("list_total", 100)
	m.RecordCustom("list_total", 110)
	m.MarkEnded()

	s := m.Snapshot()
	if len(s.TimeSeries) == 0 {
		t.Fatal("expected time series buckets")
	}
	if s.Total != 2 {
		t.Fatalf("total=%d want 2", s.Total)
	}
	if len(s.CustomSeries["list_total"]) == 0 {
		t.Fatal("expected list_total custom series")
	}
}

func TestMetricsRecordStep(t *testing.T) {
	m := NewMetrics(time.Second)
	m.RecordStep("list", 10*time.Millisecond, nil)
	m.RecordStep("list", 20*time.Millisecond, nil)
	m.RecordStep("queryById", 5*time.Millisecond, errTest("fail"))
	m.MarkEnded()

	s := m.Snapshot()
	if len(s.Steps) != 2 {
		t.Fatalf("steps=%d want 2", len(s.Steps))
	}
	if s.Steps[0].Name != "list" || s.Steps[0].Success != 2 {
		t.Fatalf("list step=%+v", s.Steps[0])
	}
	if s.Steps[1].Name != "queryById" || s.Steps[1].Failed != 1 {
		t.Fatalf("query step=%+v", s.Steps[1])
	}
}

func TestRenderStepSection(t *testing.T) {
	out := renderMarkdown(ReportInput{
		Meta:    ScenarioMeta{Name: "test", Type: TypeFlow, Title: "Test"},
		Options: Options{Duration: time.Second, Rate: 10, Concurrency: 1, Warmup: 0, Timeout: time.Second},
		Metrics: Snapshot{
			Steps: []StepSummary{
				{Name: "list", Total: 10, Success: 10, Latency: LatencySummary{P50: 5 * time.Millisecond, P95: 8 * time.Millisecond, Max: 10 * time.Millisecond}},
			},
		},
	}, "local", "http://localhost", "user")
	if !strings.Contains(out, "## Flow 步骤延迟") {
		t.Fatal("missing step section")
	}
}

func TestMetricsSnapshot(t *testing.T) {
	m := NewMetrics(time.Second)
	m.MarkStarted()
	m.Record(10*time.Millisecond, nil)
	m.Record(20*time.Millisecond, nil)
	m.Record(30*time.Millisecond, nil)
	m.Record(0, errTest("timeout"))
	m.MarkEnded()

	s := m.Snapshot()
	if s.Total != 4 {
		t.Fatalf("total=%d want 4", s.Total)
	}
	if s.Success != 3 || s.Failed != 1 {
		t.Fatalf("success=%d failed=%d", s.Success, s.Failed)
	}
	if s.Latency.P50 != 20*time.Millisecond {
		t.Fatalf("p50=%v want 20ms", s.Latency.P50)
	}
	if s.Errors["timeout"] != 1 {
		t.Fatalf("errors=%v", s.Errors)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func TestRenderTimeSeriesSection(t *testing.T) {
	snap := Snapshot{
		TimeSeries: []TimePoint{
			{OffsetSec: 0, Requests: 20, Success: 20, QPS: 20, SuccessPct: 100, LatencyP95: 50 * time.Millisecond},
			{OffsetSec: 1, Requests: 18, Success: 17, Failed: 1, QPS: 18, SuccessPct: 94.4, LatencyP95: 80 * time.Millisecond},
		},
		CustomSeries: map[string][]CustomPoint{
			"list_total": {{OffsetSec: 0, Avg: 100}, {OffsetSec: 1, Avg: 98}},
		},
	}

	out := renderTimeSeriesSection(snap, ReportOptions{Charts: true, ChartMetrics: []string{"list_total"}})
	if !strings.Contains(out, "## 压测过程") {
		t.Fatal("missing section")
	}
	if !strings.Contains(out, "```mermaid") {
		t.Fatal("missing mermaid chart")
	}
	if !strings.Contains(out, "list_total") {
		t.Fatal("missing custom metric chart")
	}
}

func TestParseDefaults(t *testing.T) {
	opts, err := parseDefaults(defaultsYAML{
		Duration:       "15s",
		Rate:           50,
		Concurrency:    5,
		Warmup:         "2s",
		Timeout:        "10s",
		BucketInterval: "2s",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opts.BucketInterval != 2*time.Second {
		t.Fatalf("bucket=%v", opts.BucketInterval)
	}
}
