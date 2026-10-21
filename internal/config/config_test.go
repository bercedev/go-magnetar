package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSetProfileUpdatesProfileKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	testConfig := `version: "1.0"
profile: work

profiles:
  work:
    llm:
      model: work-model
  home:
    llm:
      model: home-model
`
	if err := os.WriteFile(path, []byte(testConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := cfg.SetProfile("home"); err != nil {
		t.Fatal(err)
	}
	if got := cfg.String("profile"); got != "home" {
		t.Errorf("String(\"profile\") = %q after SetProfile(\"home\"), want \"home\"", got)
	}
	if got := cfg.Profile(); got != "home" {
		t.Errorf("Profile() = %q, want \"home\"", got)
	}
	if got := cfg.ProfileParamString("llm.model"); got != "home-model" {
		t.Errorf("ProfileParamString(\"llm.model\") = %q, want \"home-model\"", got)
	}

	// an unknown profile changes nothing
	if err := cfg.SetProfile("missing"); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("SetProfile(\"missing\") error = %v, want ErrProfileNotFound", err)
	}
	if got := cfg.String("profile"); got != "home" {
		t.Errorf("String(\"profile\") = %q after a failed SetProfile, want \"home\"", got)
	}
}
