package timingplugin

import (
	"context"
	"testing"

	"github.com/wmentor/go-magnetar/internal/plugin"
)

type captureHub struct {
	plugin.Hub
	command plugin.ChatCommand
}

func (h *captureHub) RegisterChatCommand(command plugin.ChatCommand) {
	h.command = command
}

func registeredCommand(t *testing.T, state *plugin.State) plugin.ChatCommand {
	t.Helper()
	hub := &captureHub{}
	if err := (&Plugin{}).Init(state, hub); err != nil {
		t.Fatal(err)
	}
	return hub.command
}

func run(t *testing.T, command plugin.ChatCommand, args string) string {
	t.Helper()
	resp, err := command.Execute(context.Background(), &plugin.ChatCommandInput{Args: args})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.NoLog {
		t.Errorf("/timing %s: a status line must not go to the response store", args)
	}
	return resp.Stdout
}

func TestTimingOnOffAndStatus(t *testing.T) {
	t.Parallel()
	state := &plugin.State{}
	command := registeredCommand(t, state)
	if command.Name != "timing" {
		t.Fatalf("unexpected command name %q", command.Name)
	}

	steps := []struct {
		args    string
		stdout  string
		enabled bool
	}{
		{"", "Execution time measurement is currently OFF.\n", false},
		{"on", "Execution time measurement enabled.\n", true},
		{"", "Execution time measurement is currently ON.\n", true},
		{"off", "Execution time measurement disabled.\n", false},
	}
	for _, step := range steps {
		if got := run(t, command, step.args); got != step.stdout {
			t.Errorf("/timing %s: got %q, want %q", step.args, got, step.stdout)
		}
		if got := state.TimingEnabled.Load(); got != step.enabled {
			t.Errorf("after /timing %s: TimingEnabled = %v, want %v", step.args, got, step.enabled)
		}
	}
}

func TestTimingOffTurnsOffTimingStartedByTheCLIFlag(t *testing.T) {
	t.Parallel()
	// --enable-timing only sets the starting value; /timing off still applies.
	state := &plugin.State{}
	state.TimingEnabled.Store(true)
	command := registeredCommand(t, state)

	if got, want := run(t, command, "off"), "Execution time measurement disabled.\n"; got != want {
		t.Errorf("/timing off: got %q, want %q", got, want)
	}
	if state.TimingEnabled.Load() {
		t.Error("/timing off must turn off timing that --enable-timing turned on")
	}
}
