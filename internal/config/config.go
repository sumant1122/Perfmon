package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

type Tab struct {
	Title       string   `toml:"title"`
	Cmd         []string `toml:"cmd"`
	Disabled    bool     `toml:"-"`
	DisabledMsg string   `toml:"-"`
}

type Config struct {
	Tabs []Tab `toml:"tab"`
}

func Load() []Tab {
	if cfgTabs, ok := loadFromConfig(); ok {
		validated := make([]Tab, 0, len(cfgTabs))
		for _, t := range cfgTabs {
			validated = append(validated, validateTab(t))
		}
		if len(validated) > 0 {
			return validated
		}
	}
	return buildDefaultTabs()
}

func loadFromConfig() ([]Tab, bool) {
	paths := configPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg Config
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			// In a real app we might want to log this error
			continue
		}
		if len(cfg.Tabs) == 0 {
			continue
		}
		
		// Filter invalid tabs
		validTabs := make([]Tab, 0, len(cfg.Tabs))
		for _, t := range cfg.Tabs {
			if t.Title != "" && len(t.Cmd) > 0 {
				validTabs = append(validTabs, t)
			}
		}
		
		if len(validTabs) > 0 {
			return validTabs, true
		}
	}
	return nil, false
}

func configPaths() []string {
	var paths []string
	if env := strings.TrimSpace(os.Getenv("PERFMON_CONFIG")); env != "" {
		paths = append(paths, env)
	}
	if cfgDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(cfgDir, "perfmon", "config.toml"))
	}
	paths = append(paths, "perfmon.toml")
	return paths
}

func validateTab(t Tab) Tab {
	if len(t.Cmd) == 0 {
		t.Disabled = true
		t.DisabledMsg = "No command configured for this tab."
		return t
	}

	// If fetch is already handled by a safe echo command, leave it enabled.
	if t.Cmd[0] == "echo" {
		return t
	}

	if _, err := exec.LookPath(t.Cmd[0]); err == nil {
		return t
	}

	t.Disabled = true
	t.DisabledMsg = missingHint(t.Cmd[0], t.Title)
	return t
}

func missingHint(cmd, title string) string {
	switch cmd {
	case "mpstat", "pidstat", "sar", "iostat":
		return fmt.Sprintf("Missing %s. Install sysstat to enable this tab.", title)
	case "vm_stat":
		return "Missing vm_stat. This tab requires macOS."
	case "vmstat":
		return "Missing vmstat. Install procps/sysstat to enable this tab."
	case "free":
		return "Missing free. Install procps to enable this tab."
	case "top":
		return "Missing top. Install procps to enable this tab."
	case "uptime":
		return "Missing uptime. Install coreutils to enable this tab."
	}
	return fmt.Sprintf("Missing %s. Install the command to enable this tab.", cmd)
}

func buildDefaultTabs() []Tab {
	freeCmd := []string{"free", "-m"}
	freeTitle := "free -m"
	if runtime.GOOS == "darwin" {
		freeCmd = []string{"vm_stat"}
		freeTitle = "vm_stat (free)"
	}

	topCmd := []string{"top", "-b", "-n", "1"}
	topTitle := "top -b -n 1"
	if runtime.GOOS == "darwin" {
		topCmd = []string{"top", "-l", "1"}
		topTitle = "top -l 1"
	}

	fetchTitle, fetchCmd := detectFetchCmd()

	tabs := []Tab{
		{Title: "uptime", Cmd: []string{"uptime"}},
		{Title: "vmstat", Cmd: []string{"vmstat"}},
		{Title: "mpstat -P ALL", Cmd: []string{"mpstat", "-P", "ALL"}},
		{Title: "pidstat -p ALL", Cmd: []string{"pidstat", "-p", "ALL"}},
		{Title: "iostat", Cmd: []string{"iostat"}},
		{Title: freeTitle, Cmd: freeCmd},
		{Title: "sar -n DEV", Cmd: []string{"sar", "-n", "DEV"}},
		{Title: "sar -n TCP,ETCP", Cmd: []string{"sar", "-n", "TCP,ETCP"}},
		{Title: topTitle, Cmd: topCmd},
		{Title: fetchTitle, Cmd: fetchCmd},
	}

	for i := range tabs {
		tabs[i] = validateTab(tabs[i])
	}

	return tabs
}

func detectFetchCmd() (string, []string) {
	if _, err := exec.LookPath("fastfetch"); err == nil {
		return "fastfetch", []string{"fastfetch"}
	}
	if _, err := exec.LookPath("neofetch"); err == nil {
		return "neofetch", []string{"neofetch"}
	}
	if _, err := exec.LookPath("screenfetch"); err == nil {
		return "screenfetch", []string{"screenfetch"}
	}
	return "fetch (missing)", []string{"echo", "No fetch tool found. Install fastfetch to enable this tab."}
}
