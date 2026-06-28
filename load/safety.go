package load

import (
	"fmt"
	"strings"

	"github.com/muyi-zcy/my-testgogogo/config"
)

// ValidateSafety 校验压测是否允许在当前环境执行。
func ValidateSafety(cfg *Config, confirmed bool) error {
	if cfg == nil {
		return nil
	}

	appCfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load app config: %w", err)
	}

	active := appCfg.Active
	if len(cfg.AllowedActive) > 0 && !containsIgnoreCase(cfg.AllowedActive, active) {
		return fmt.Errorf("load blocked: active env %q not in load.allowed_active %v", active, cfg.AllowedActive)
	}

	if cfg.RequireConfirm && !confirmed {
		return fmt.Errorf("load blocked: set load.require_confirm or pass --confirm to run against %q (%s)", active, appCfg.BaseURL)
	}

	return nil
}

func containsIgnoreCase(list []string, val string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(val)) {
			return true
		}
	}
	return false
}
