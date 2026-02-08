package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
)

func TestParseTomlConfig(t *testing.T) {
	data := []byte(`
global_refresh_interval = "10s"

[[tab]]
title = "uptime"
cmd = ["uptime"]

[[tab]]
title = "top"
cmd = ["top","-b","-n","1"]
refresh_interval = "1s"
`)
	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if cfg.GlobalRefreshInterval.Duration != 10*time.Second {
		t.Errorf("expected global 10s, got %v", cfg.GlobalRefreshInterval.Duration)
	}

	if len(cfg.Tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(cfg.Tabs))
	}

	// Note: Load() handles applying global defaults to tabs, so we just check raw parsing here
	if cfg.Tabs[1].RefreshInterval.Duration != 1*time.Second {
		t.Errorf("expected tab 1s, got %v", cfg.Tabs[1].RefreshInterval.Duration)
	}
}

func TestLoadTabsFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perfmon.toml")
	err := os.WriteFile(path, []byte(`
global_refresh_interval = "2s"
[[tab]]
title = "vmstat"
cmd = ["vmstat"]
`), 0o644)
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Setenv("PERFMON_CONFIG", path)
	_, tabs := Load() // Load now returns (Config, []Tab)

	if len(tabs) != 1 {
		t.Fatalf("expected 1 tab")
	}

	if tabs[0].RefreshInterval.Duration != 2*time.Second {
		t.Errorf("expected inherited 2s refresh, got %v", tabs[0].RefreshInterval.Duration)
	}
}
