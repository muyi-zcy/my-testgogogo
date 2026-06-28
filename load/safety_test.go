package load

import "testing"

func TestValidateSafetyAllowedActive(t *testing.T) {
	cfg := &Config{
		AllowedActive: []string{"local"},
	}
	if err := ValidateSafety(cfg, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSafetyRequireConfirm(t *testing.T) {
	cfg := &Config{
		RequireConfirm: true,
		AllowedActive:  []string{"local"},
	}
	if err := ValidateSafety(cfg, false); err == nil {
		t.Fatal("expected confirm error")
	}
	if err := ValidateSafety(cfg, true); err != nil {
		t.Fatalf("unexpected error with confirm: %v", err)
	}
}
