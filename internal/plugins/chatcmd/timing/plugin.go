package timingplugin

import (
	"context"

	"github.com/wmentor/go-magnetar/internal/plugin"
)

func init() {
	plugin.Register("chatcmd.timing", &Plugin{})
}

// Plugin registers the /timing chat command.
type Plugin struct {
	state *plugin.State
}

func (p *Plugin) Init(s *plugin.State, hub plugin.Hub) error {
	p.state = s

	hub.RegisterChatCommand(plugin.ChatCommand{
		Name:    "timing",
		Help:    "Show or toggle the execution time display (on|off).",
		Execute: p.execute,
	})
	return nil
}

func (p *Plugin) execute(_ context.Context, params *plugin.ChatCommandInput) (*plugin.ChatCommandResponse, error) {
	switch params.Args {
	case "on":
		p.state.TimingEnabled.Store(true)
		return message("Execution time measurement enabled."), nil
	case "off":
		p.state.TimingEnabled.Store(false)
		return message("Execution time measurement disabled."), nil
	default:
		if p.state.TimingEnabled.Load() {
			return message("Execution time measurement is currently ON."), nil
		}
		return message("Execution time measurement is currently OFF."), nil
	}
}

// message is a status line for the user, kept out of the response store used by /copy and /revert.
func message(text string) *plugin.ChatCommandResponse {
	return &plugin.ChatCommandResponse{Stdout: text + "\n", NoLog: true}
}
