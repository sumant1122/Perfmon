package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTomlConfig(t *testing.T) {
	data := []byte(`
[[tab]]
title = "uptime"
cmd = ["uptime"]

[[tab]]
title = "top"
cmd = ["top","-b","-n","1"]
`)
	cfg, err := parseTomlConfig(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(cfg.Tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(cfg.Tabs))
	}
	if cfg.Tabs[0].Title != "uptime" || len(cfg.Tabs[0].Cmd) != 1 {
		t.Fatalf("unexpected first tab: %+v", cfg.Tabs[0])
	}
	if cfg.Tabs[1].Cmd[0] != "top" || cfg.Tabs[1].Cmd[3] != "1" {
		t.Fatalf("unexpected second tab cmd: %+v", cfg.Tabs[1].Cmd)
	}
}

func TestLoadTabsFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perfmon.toml")
	err := os.WriteFile(path, []byte(`
[[tab]]
title = "vmstat"
cmd = ["vmstat"]
`), 0o644)
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Setenv("PERFMON_CONFIG", path)
	tabs, ok := loadTabsFromConfig()
	if !ok {
		t.Fatalf("expected config load")
	}
	if len(tabs) != 1 || tabs[0].title != "vmstat" || tabs[0].cmd[0] != "vmstat" {
		t.Fatalf("unexpected tabs: %+v", tabs)
	}
}

func TestRenderTabsOverflow(t *testing.T) {
	applyTheme(0)
	tabs := []tab{
		{title: "a", cmd: []string{"a"}},
		{title: "b", cmd: []string{"b"}},
		{title: "c", cmd: []string{"c"}},
		{title: "d", cmd: []string{"d"}},
		{title: "e", cmd: []string{"e"}},
		{title: "f", cmd: []string{"f"}},
	}
	row := renderTabs(tabs, 3, 10)
	if row == "" {
		t.Fatalf("expected rendered row")
	}
}
