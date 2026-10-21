package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/wmentor/go-magnetar/internal/config"
	"github.com/wmentor/go-magnetar/internal/plugin"
	"github.com/wmentor/go-magnetar/internal/plugin/testdata"
	skillplugin "github.com/wmentor/go-magnetar/internal/plugins/chatcmd/skill"
	"github.com/wmentor/go-magnetar/internal/printer"
)

type commandHub struct {
	commands []plugin.ChatCommand
}

func (h *commandHub) RegisterTool(plugin.LLMTool) {}
func (h *commandHub) RegisterChatCommand(cmd plugin.ChatCommand) {
	h.commands = append(h.commands, cmd)
}
func (h *commandHub) RegisterPreprocessor(plugin.PreprocessorFunc) {}
func (h *commandHub) Go(func(ctx context.Context))                 {}
func (h *commandHub) Stop()                                        {}

func TestApplySkill(t *testing.T) {
	t.Parallel()

	store := testdata.NewSkillStore(t)
	store.EXPECT().Get(mock.Anything, "review").
		Return(&plugin.Skill{Name: "review", Content: "Review the code.", Enable: true}, nil)
	store.EXPECT().Get(mock.Anything, "draft").
		Return(&plugin.Skill{Name: "draft", Content: "Not ready.", Enable: false}, nil)
	hub := &commandHub{}
	if err := (&skillplugin.Plugin{}).Init(&plugin.State{SkillStore: store}, hub); err != nil {
		t.Fatalf("skill plugin init: %v", err)
	}

	t.Run("skill alone, as at the start of a chat", func(t *testing.T) {
		got, err := applySkill(hub.commands, "review", "")
		if err != nil {
			t.Fatalf("applySkill returned error: %v", err)
		}
		if !strings.HasPrefix(got, "# SKILL: review\n\nReview the code.") || strings.Contains(got, "# Problem:") {
			t.Fatalf("applySkill = %q", got)
		}
	})

	t.Run("skill and input in one message, as with -f", func(t *testing.T) {
		got, err := applySkill(hub.commands, "review", "Check main.go\nplease")
		if err != nil {
			t.Fatalf("applySkill returned error: %v", err)
		}
		if !strings.HasPrefix(got, "# SKILL: review") || !strings.HasSuffix(got, "# Problem:\n\nCheck main.go\nplease\n") {
			t.Fatalf("applySkill = %q", got)
		}
	})

	t.Run("disabled skill is an error", func(t *testing.T) {
		if _, err := applySkill(hub.commands, "draft", ""); err == nil || !strings.Contains(err.Error(), "not enabled") {
			t.Fatalf("applySkill error = %v, want not enabled", err)
		}
	})

	t.Run("no skill command is an error", func(t *testing.T) {
		if _, err := applySkill(nil, "review", ""); err == nil {
			t.Fatal("applySkill returned no error without a skill command")
		}
	})
}

func TestPrintEnabledModules(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    []string
		notWant []string
	}{
		{
			name:    "nothing enabled",
			yaml:    "",
			notWant: []string{"ssh plugin is enabled", "guard plugin is enabled"},
		},
		{
			name:    "ssh enabled",
			yaml:    "ssh:\n  enable: true\n",
			want:    []string{"ssh plugin is enabled"},
			notWant: []string{"guard plugin is enabled"},
		},
		{
			name:    "guard enabled",
			yaml:    "guard:\n  enable: true\n",
			want:    []string{"guard plugin is enabled"},
			notWant: []string{"ssh plugin is enabled"},
		},
		{
			name: "ssh and guard enabled",
			yaml: "ssh:\n  enable: true\nguard:\n  enable: true\n",
			want: []string{"ssh plugin is enabled", "guard plugin is enabled"},
		},
		{
			name:    "ssh and guard explicitly disabled",
			yaml:    "ssh:\n  enable: false\nguard:\n  enable: false\n",
			notWant: []string{"ssh plugin is enabled", "guard plugin is enabled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(path, []byte("version: \"1.0\"\n"+tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}

			cfg, err := config.Load(path)
			if err != nil {
				t.Fatal(err)
			}

			// Use bytes.Buffer as the printer output
			var buf bytes.Buffer
			p := printer.New(&buf, false)

			printEnabledModules(p, cfg)

			out := buf.String()

			for _, s := range tt.want {
				if !strings.Contains(out, s) {
					t.Errorf("output %q does not contain %q", out, s)
				}
			}
			for _, s := range tt.notWant {
				if strings.Contains(out, s) {
					t.Errorf("output %q unexpectedly contains %q", out, s)
				}
			}
		})
	}
}

func TestReadFileInput(t *testing.T) {
	t.Parallel()

	t.Run("reads stdin when filename is dash", func(t *testing.T) {
		got, err := readFileInput("-", strings.NewReader("from stdin"))
		if err != nil {
			t.Fatalf("readFileInput returned error: %v", err)
		}
		if string(got) != "from stdin" {
			t.Fatalf("readFileInput = %q, want %q", got, "from stdin")
		}
	})

	t.Run("reads named file", func(t *testing.T) {
		path := filepath.Join("testdata", "prompt.txt")

		got, err := readFileInput(path, strings.NewReader("ignored stdin"))
		if err != nil {
			t.Fatalf("readFileInput returned error: %v", err)
		}
		if string(got) != "from file\n" {
			t.Fatalf("readFileInput = %q, want %q", got, "from file\n")
		}
	})
}
