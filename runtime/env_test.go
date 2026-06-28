package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/stretchr/testify/assert"
)

func TestNewVarsUsesSeed(t *testing.T) {
	vars := NewVars(nil)
	assert.Equal(t, 10, vars.MustInt("pageSize"))
}

func TestRunStepInvokesRecorder(t *testing.T) {
	var called bool
	var stepName string
	base := New(&config.Config{}, client.New("http://localhost", 0), nil, context.Background())
	base.SetStepRecorder(func(name string, d time.Duration, err error) {
		called = true
		stepName = name
	})
	err := base.RunStep("demo", func() error { return nil })
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "demo", stepName)
}

func TestCloneForWorkerIsolatesVars(t *testing.T) {
	base := New(&config.Config{}, client.New("http://localhost", 0), nil, context.Background())
	base.Vars.Set("key", 1)

	clone := base.CloneForWorker()
	if clone.Vars == base.Vars {
		t.Fatal("expected isolated vars")
	}
	clone.Vars.Set("key", 2)

	assert.Equal(t, 1, base.Vars.MustInt("key"))
	assert.Equal(t, 2, clone.Vars.MustInt("key"))
}
