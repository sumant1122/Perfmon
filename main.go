package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tab struct {
	title string
	cmd   []string
}

type tickMsg time.Time

type cmdResultMsg struct {
	output string
	err    error
}

type model struct {
	tabs       []tab
	active     int
	viewport   viewport.Model
	content    string
	statusLine string
	width      int
	height     int
}

const refreshInterval = 5 * time.Second

func main() {
	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newModel() model {
	vp := viewport.New(0, 0)
	vp.SetContent("Loading...")

	return model{
		tabs:     buildTabs(),
		active:   0,
		viewport: vp,
	}
}

func buildTabs() []tab {
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

	return []tab{
		{title: "uptime", cmd: []string{"uptime"}},
		{title: "vmstat", cmd: []string{"vmstat"}},
		{title: "mpstat -P ALL", cmd: []string{"mpstat", "-P", "ALL"}},
		{title: "pidstat -P ALL", cmd: []string{"pidstat", "-P", "ALL"}},
		{title: "iostat", cmd: []string{"iostat"}},
		{title: freeTitle, cmd: freeCmd},
		{title: "sar -n DEV", cmd: []string{"sar", "-n", "DEV"}},
		{title: "sar -n TCP,ETCP", cmd: []string{"sar", "-n", "TCP,ETCP"}},
		{title: topTitle, cmd: topCmd},
		{title: fetchTitle, cmd: fetchCmd},
	}
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

func (m model) Init() tea.Cmd {
	return tea.Batch(runCommandCmd(m.tabs[m.active]), tick())
}

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func runCommandCmd(t tab) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, t.cmd[0], t.cmd[1:]...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		return cmdResultMsg{output: out.String(), err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "tab":
			m.active = (m.active + 1) % len(m.tabs)
			return m, m.onTabSelected()
		case "left", "h", "shift+tab":
			m.active--
			if m.active < 0 {
				m.active = len(m.tabs) - 1
			}
			return m, m.onTabSelected()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.viewport.SetContent(m.content)
	case tickMsg:
		return m, tea.Batch(runCommandCmd(m.tabs[m.active]), tick())
	case cmdResultMsg:
		m.content = strings.TrimSpace(msg.output)
		if m.content == "" {
			m.content = "(no output)"
		}
		m.viewport.SetContent(m.content)
		if msg.err != nil {
			m.statusLine = fmt.Sprintf("error: %v", msg.err)
		} else {
			m.statusLine = fmt.Sprintf("updated %s", time.Now().Format("15:04:05"))
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) onTabSelected() tea.Cmd {
	m.content = "Loading..."
	m.viewport.SetContent(m.content)
	return runCommandCmd(m.tabs[m.active])
}

func (m model) View() string {
	header := renderTabs(m.tabs, m.active, m.width)
	footer := renderFooter(m.statusLine, m.width)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

var (
	accent      = lipgloss.Color("#34B3A0")
	accentDark  = lipgloss.Color("#0F2E2B")
	ink         = lipgloss.Color("#E6EDF3")
	muted       = lipgloss.Color("#8AA1A8")
	background  = lipgloss.Color("#0B1115")
	headerStyle = lipgloss.NewStyle().Foreground(ink).Background(background).Padding(0, 1)
	activeTab   = lipgloss.NewStyle().Foreground(background).Background(accent).Bold(true).Padding(0, 1)
	inactiveTab = lipgloss.NewStyle().Foreground(muted).Background(background).Padding(0, 1)
	footerStyle = lipgloss.NewStyle().Foreground(muted).Background(background).Padding(0, 1)
)

func renderTabs(tabs []tab, active, width int) string {
	parts := make([]string, 0, len(tabs))
	for i, t := range tabs {
		if i == active {
			parts = append(parts, activeTab.Render(" "+t.title+" "))
		} else {
			parts = append(parts, inactiveTab.Render(" "+t.title+" "))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return headerStyle.Width(width).Render(row)
}

func renderFooter(status string, width int) string {
	help := "q:quit  tab/shift+tab:next/prev  up/down/pgup/pgdn:scroll"
	if status != "" {
		help = status + "  |  " + help
	}
	return footerStyle.Width(width).Render(help)
}
