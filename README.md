# Perfmon
Perfmon is TUI that shows essential performance monitoring commands in one place. It uses tabbed approach for showing the results. Its quick way to find the stats in one place

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
