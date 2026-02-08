# Perfmon
Perfmon is a TUI that shows essential performance monitoring commands in one place. It uses a tabbed approach for showing the results, making it quick to find stats in one place.

## Perfmon in Action with default settings

![Perfmon_without_toml](https://github.com/user-attachments/assets/7a94f63d-02ee-4992-b66d-9adf42a16603)

## Perfmon in Action with Toml file using user defined commands

![Perfmon_with_toml](https://github.com/user-attachments/assets/053be0f1-d3ea-4b8f-8b10-8797c7103cb2)

## Features
- Tabbed command output with keyboard navigation
- Summary strip with sparklines (load, CPU, memory, network)
- Disabled tab detection with install hints
- Configurable tabs via TOML
- Theme toggle (`t`)


[![Go Reference](https://pkg.go.dev/badge/perfmon.svg)](https://pkg.go.dev/perfmon)
[![Go Report Card](https://goreportcard.com/badge/github.com/sumant1122/Perfmon)](https://goreportcard.com/report/github.com/sumant1122/Perfmon)
[![CI](https://github.com/sumant1122/Perfmon/actions/workflows/ci.yml/badge.svg)](https://github.com/sumant1122/Perfmon/actions/workflows/ci.yml)

## Usage
```bash
go run .
```

Quit with `q`, `Q`, `Esc`, or `Ctrl+C`.

### Version
```bash
go run . --version
```

### Configuration
Perfmon can load a custom tab list from a TOML file. Search order:
1. `PERFMON_CONFIG` env var (full path)
2. `$XDG_CONFIG_HOME/perfmon/config.toml` (or `~/.config/perfmon/config.toml`)
3. `./perfmon.toml`

Example `perfmon.toml`:
```toml
# Global refresh rate (default: 5s)
global_refresh_interval = "5s"

[[tab]]
title = "uptime"
cmd = ["uptime"]

[[tab]]
title = "top (fast)"
cmd = ["top", "-b", "-n", "1"]
refresh_interval = "1s" # Override global rate
```

### Default Commands
If no configuration file is found, Perfmon falls back to a sensible default set:
- **Load**: `uptime`
- **CPU**: `vmstat` (or `mpstat`)
- **Memory**: `free -m` (Linux) or `vm_stat` (macOS)
- **Net**: `/proc/net/dev` (Linux) or `netstat` (macOS)
- **Top**: `top`


## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to set up your development environment and contribute to the project.

