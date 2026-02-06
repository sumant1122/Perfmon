# Perfmon
Perfmon is a TUI that shows essential performance monitoring commands in one place. It uses a tabbed approach for showing the results, making it quick to find stats in one place.

## Features
- Tabbed command output with keyboard navigation
- Summary strip with sparklines (load, CPU, memory, network)
- Disabled tab detection with install hints
- Configurable tabs via TOML
- Theme toggle (`t`)

## Usage
```bash
go run .
```

Quit with `q`, `Q`, `Esc`, or `Ctrl+C`.

### Version
```bash
go run . --version
```

## Config
Perfmon can load a custom tab list from a TOML file. Search order:
1. `PERFMON_CONFIG` env var (full path)
2. `$XDG_CONFIG_HOME/perfmon/config.toml` (or `~/.config/perfmon/config.toml`)
3. `./perfmon.toml`

Example `perfmon.toml`:
```toml
[[tab]]
title = "uptime"
cmd = ["uptime"]

[[tab]]
title = "top"
cmd = ["top", "-b", "-n", "1"]
```

## Command Notes
- Load: `uptime`
- CPU: `vmstat` (fallback `mpstat`)
- Memory: `free -m` (Linux)
- Net: `/proc/net/dev` (Linux) or `netstat -ib` (macOS)

If a command is missing, the tab is disabled and a hint is shown.

## Config
Perfmon can load a custom tab list from a TOML file. Search order:
1. `PERFMON_CONFIG` env var (full path)
2. `$XDG_CONFIG_HOME/perfmon/config.toml` (or `~/.config/perfmon/config.toml`)
3. `./perfmon.toml`

Example `perfmon.toml`:
```toml
[[tab]]
title = "uptime"
cmd = ["uptime"]

[[tab]]
title = "top"
cmd = ["top", "-b", "-n", "1"]
```
