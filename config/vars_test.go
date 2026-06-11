package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeVars(t *testing.T) {
	t.Parallel()

	merged := mergeVars(
		map[string]any{"shared": "root", "only_root": 1},
		map[string]any{"shared": "env", "only_env": 2},
	)
	assert.Equal(t, map[string]any{
		"shared":    "env",
		"only_root": 1,
		"only_env":  2,
	}, merged)
}

func TestConfigExpand(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		User: UserConfig{
			Username: "alice",
			Password: "secret",
			Type:     "app",
		},
		Vars: map[string]any{
			"user_id": 1001,
			"enabled": true,
		},
	}

	assert.Equal(t, "alice", cfg.Expand("{{user.username}}"))
	assert.Equal(t, "1001", cfg.Expand("uid={{vars.user_id}}"))
	assert.Equal(t, "alice/1001", cfg.Expand("{{user.username}}/{{vars.user_id}}"))
	assert.Equal(t, "true", cfg.Expand("{{vars.enabled}}"))
}

func TestConfigVarsInto(t *testing.T) {
	t.Parallel()

	type testVars struct {
		UserID   string `yaml:"user_id"`
		PageSize int    `yaml:"page_size"`
		Enabled  bool   `yaml:"enabled"`
	}

	cfg := &Config{
		Vars: map[string]any{
			"user_id":   "42",
			"page_size": 20,
			"enabled":   true,
		},
	}

	var vars testVars
	require.NoError(t, cfg.VarsInto(&vars))
	assert.Equal(t, testVars{
		UserID:   "42",
		PageSize: 20,
		Enabled:  true,
	}, vars)
}

func TestConfigVarAccessors(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Vars: map[string]any{
			"user_id":   "42",
			"page_size": 20,
			"ratio":     1.5,
		},
	}

	val, ok := cfg.Var("user_id")
	require.True(t, ok)
	assert.Equal(t, "42", val)

	got, err := cfg.VarString("user_id")
	require.NoError(t, err)
	assert.Equal(t, "42", got)

	gotInt, err := cfg.VarInt("page_size")
	require.NoError(t, err)
	assert.Equal(t, 20, gotInt)

	_, err = cfg.VarInt("ratio")
	assert.Error(t, err)

	_, err = cfg.VarString("missing")
	assert.Error(t, err)
}

func TestLoadVarsFromExampleWMS(t *testing.T) {
	t.Chdir("examples/wms")

	cfg, err := Load()
	require.NoError(t, err)

	var vars struct {
		PageSize int    `yaml:"page_size"`
		UserID   string `yaml:"user_id"`
	}
	require.NoError(t, cfg.VarsInto(&vars))
	assert.Equal(t, 10, vars.PageSize)
	assert.Equal(t, "1", vars.UserID)
}
