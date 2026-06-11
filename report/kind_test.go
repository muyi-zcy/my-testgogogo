package report

import "testing"

func TestKindFromPackage(t *testing.T) {
	if got := KindFromPackage("examples/demo/api/book"); got != KindAPI {
		t.Fatalf("got %q want api", got)
	}
	if got := KindFromPackage("examples/demo/flow/example"); got != KindFlow {
		t.Fatalf("got %q want flow", got)
	}
}

func TestConfigOutputDirs(t *testing.T) {
	cfg := &Config{BaseDir: "/tmp/reports"}
	if got := cfg.OutputDir(KindAPI); got != "/tmp/reports/api" {
		t.Fatalf("api dir=%q", got)
	}
	if got := cfg.StagingDir(KindFlow); got != "/tmp/reports/flow/staging" {
		t.Fatalf("flow staging=%q", got)
	}
}

func TestReportFileName(t *testing.T) {
	if got := ReportFileName(KindAPI, "20260611-120000"); got != "api-report-20260611-120000.md" {
		t.Fatalf("got %q", got)
	}
}
