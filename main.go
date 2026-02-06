package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tab struct {
	title       string
	cmd         []string
	disabled    bool
	disabledMsg string
}

type tickMsg time.Time

type cmdResultMsg struct {
	output string
	err    error
}

type spinnerMsg time.Time

type metricsMsg struct {
	metrics metricsSample
}

type systemMsg struct {
	info systemInfo
}

type metricsSample struct {
	load   float64
	cpu    float64
	mem    float64
	netKB  float64
	okLoad bool
	okCPU  bool
	okMem  bool
	okNet  bool
}

type metricHistory struct {
	load []float64
	cpu  []float64
	mem  []float64
	net  []float64
}

type systemInfo struct {
	snapshot string
	disk     string
	net      string
}

type model struct {
	tabs       []tab
	active     int
	viewport   viewport.Model
	content    string
	statusLine string
	metrics    metricHistory
	system     systemInfo
	themeIndex int
	spinnerIdx int
	width      int
	height     int
}

const refreshInterval = 5 * time.Second
const historyLength = 30
const spinnerInterval = 200 * time.Millisecond
const fixedRows = 9
const version = "0.1.0"

func main() {
	if printVersion() {
		return
	}
	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printVersion() bool {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&showVersion, "v", false, "print version and exit")
	flag.Parse()
	if showVersion {
		fmt.Printf("perfmon %s\n", version)
		return true
	}
	return false
}

func newModel() model {
	vp := viewport.New(0, 0)
	vp.SetContent("Loading...")

	applyTheme(0)
	return model{
		tabs:       loadTabs(),
		active:     0,
		viewport:   vp,
		themeIndex: 0,
	}
}

type configFile struct {
	Tabs []configTab
}

type configTab struct {
	Title string
	Cmd   []string
}

func loadTabs() []tab {
	if cfgTabs, ok := loadTabsFromConfig(); ok {
		validated := make([]tab, 0, len(cfgTabs))
		for _, t := range cfgTabs {
			validated = append(validated, validateTab(t))
		}
		if len(validated) > 0 {
			return validated
		}
	}
	return buildDefaultTabs()
}

func loadTabsFromConfig() ([]tab, bool) {
	paths := configPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		cfg, err := parseTomlConfig(data)
		if err != nil {
			continue
		}
		if len(cfg.Tabs) == 0 {
			continue
		}
		tabs := make([]tab, 0, len(cfg.Tabs))
		for _, ct := range cfg.Tabs {
			if ct.Title == "" || len(ct.Cmd) == 0 {
				continue
			}
			tabs = append(tabs, tab{title: ct.Title, cmd: ct.Cmd})
		}
		if len(tabs) > 0 {
			return tabs, true
		}
	}
	return nil, false
}

func parseTomlConfig(data []byte) (configFile, error) {
	var cfg configFile
	var current *configTab

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[[tab]]" {
			cfg.Tabs = append(cfg.Tabs, configTab{})
			current = &cfg.Tabs[len(cfg.Tabs)-1]
			continue
		}
		if current == nil {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "title":
			if s, ok := parseTomlString(val); ok {
				current.Title = s
			}
		case "cmd":
			if arr, ok := parseTomlStringArray(val); ok {
				current.Cmd = arr
			}
		}
	}
	return cfg, scanner.Err()
}

func parseTomlString(val string) (string, bool) {
	val = strings.TrimSpace(val)
	if len(val) < 2 || val[0] != '"' || val[len(val)-1] != '"' {
		return "", false
	}
	return strings.Trim(val, `"`), true
}

func parseTomlStringArray(val string) ([]string, bool) {
	val = strings.TrimSpace(val)
	if !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
		return nil, false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(val, "["), "]"))
	if inner == "" {
		return []string{}, true
	}
	parts := strings.Split(inner, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s, ok := parseTomlString(strings.TrimSpace(p))
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
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

func buildDefaultTabs() []tab {
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

	tabs := []tab{
		{title: "uptime", cmd: []string{"uptime"}},
		{title: "vmstat", cmd: []string{"vmstat"}},
		{title: "mpstat -P ALL", cmd: []string{"mpstat", "-P", "ALL"}},
		{title: "pidstat -p ALL", cmd: []string{"pidstat", "-p", "ALL"}},
		{title: "iostat", cmd: []string{"iostat"}},
		{title: freeTitle, cmd: freeCmd},
		{title: "sar -n DEV", cmd: []string{"sar", "-n", "DEV"}},
		{title: "sar -n TCP,ETCP", cmd: []string{"sar", "-n", "TCP,ETCP"}},
		{title: topTitle, cmd: topCmd},
		{title: fetchTitle, cmd: fetchCmd},
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

func validateTab(t tab) tab {
	if len(t.cmd) == 0 {
		t.disabled = true
		t.disabledMsg = "No command configured for this tab."
		return t
	}

	// If fetch is already handled by a safe echo command, leave it enabled.
	if t.cmd[0] == "echo" {
		return t
	}

	if _, err := exec.LookPath(t.cmd[0]); err == nil {
		return t
	}

	t.disabled = true
	t.disabledMsg = missingHint(t.cmd[0], t.title)
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

func (m model) Init() tea.Cmd {
	if m.tabs[m.active].disabled {
		m.content = m.tabs[m.active].disabledMsg
		m.viewport.SetContent(m.content)
		return tea.Batch(tick(), spinnerTick(), sampleMetricsCmd(), sampleSystemCmd())
	}
	return tea.Batch(runCommandCmd(m.tabs[m.active]), tick(), spinnerTick(), sampleMetricsCmd(), sampleSystemCmd())
}

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func spinnerTick() tea.Cmd {
	return tea.Tick(spinnerInterval, func(t time.Time) tea.Msg { return spinnerMsg(t) })
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
		if isQuitKey(msg) {
			return m, tea.Quit
		}
		switch msg.String() {
		case "ctrl+c":
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
		case "t":
			m.themeIndex = (m.themeIndex + 1) % len(themes)
			applyTheme(m.themeIndex)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = clampMin(msg.Width-2, 0)
		m.viewport.Height = clampMin(msg.Height-fixedRows, 0)
		m.viewport.SetContent(m.content)
	case tickMsg:
		if m.tabs[m.active].disabled {
			return m, tea.Batch(tick(), sampleMetricsCmd(), sampleSystemCmd())
		}
		return m, tea.Batch(runCommandCmd(m.tabs[m.active]), tick(), sampleMetricsCmd(), sampleSystemCmd())
	case spinnerMsg:
		m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
		return m, spinnerTick()
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
	case metricsMsg:
		m.metrics = updateHistory(m.metrics, msg.metrics)
	case systemMsg:
		m.system = msg.info
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) onTabSelected() tea.Cmd {
	if m.tabs[m.active].disabled {
		m.content = m.tabs[m.active].disabledMsg
		m.viewport.SetContent(m.content)
		m.statusLine = "disabled"
		return nil
	}
	m.content = "Loading..."
	m.viewport.SetContent(m.content)
	return runCommandCmd(m.tabs[m.active])
}

func (m model) View() string {
	header := renderTabs(m.tabs, m.active, m.width)
	summary := renderSummary(m.metrics, m.width)
	snapshot := renderInfoLine(m.system.snapshot, m.width)
	disk := renderInfoLine(m.system.disk, m.width)
	net := renderInfoLine(m.system.net, m.width)
	title := renderContentTitle(m.tabs[m.active].title, m.width)
	content := contentBoxStyle.Width(m.width).Render(m.viewport.View())
	footer := renderFooter(m.statusLine, spinnerFrames[m.spinnerIdx], m.width)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		summary,
		snapshot,
		disk,
		net,
		title,
		content,
		footer,
	)
}

var (
	accent          lipgloss.Color
	accentDark      lipgloss.Color
	ink             lipgloss.Color
	muted           lipgloss.Color
	background      lipgloss.Color
	headerStyle     lipgloss.Style
	activeTab       lipgloss.Style
	inactiveTab     lipgloss.Style
	disabledTab     lipgloss.Style
	footerStyle     lipgloss.Style
	summaryStyle    lipgloss.Style
	infoStyle       lipgloss.Style
	contentBoxStyle lipgloss.Style
	overflowStyle   lipgloss.Style
)

type theme struct {
	name       string
	accent     string
	accentDark string
	ink        string
	muted      string
	background string
}

var themes = []theme{
	{
		name:       "Ocean",
		accent:     "#34B3A0",
		accentDark: "#0F2E2B",
		ink:        "#E6EDF3",
		muted:      "#8AA1A8",
		background: "#0B1115",
	},
	{
		name:       "Sand",
		accent:     "#D7A86E",
		accentDark: "#332819",
		ink:        "#F2E8D5",
		muted:      "#B8A387",
		background: "#1A140D",
	},
	{
		name:       "Day",
		accent:     "#3B82F6",
		accentDark: "#E6EEF9",
		ink:        "#0B1220",
		muted:      "#506072",
		background: "#F7FAFF",
	},
}

var spinnerFrames = []string{"|", "/", "-", "\\"}

func applyTheme(index int) {
	if index < 0 || index >= len(themes) {
		index = 0
	}
	t := themes[index]
	accent = lipgloss.Color(t.accent)
	accentDark = lipgloss.Color(t.accentDark)
	ink = lipgloss.Color(t.ink)
	muted = lipgloss.Color(t.muted)
	background = lipgloss.Color(t.background)

	headerStyle = lipgloss.NewStyle().Foreground(ink).Background(background).Padding(0, 1)
	activeTab = lipgloss.NewStyle().Foreground(background).Background(accent).Bold(true).Padding(0, 1)
	inactiveTab = lipgloss.NewStyle().Foreground(muted).Background(background).Padding(0, 1)
	disabledTab = lipgloss.NewStyle().Foreground(muted).Background(background).Faint(true).Padding(0, 1)
	footerStyle = lipgloss.NewStyle().Foreground(muted).Background(background).Padding(0, 1)
	summaryStyle = lipgloss.NewStyle().Foreground(ink).Background(accentDark).Padding(0, 1)
	infoStyle = lipgloss.NewStyle().Foreground(ink).Background(background).Padding(0, 1)
	contentBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(muted).
		Padding(0, 1)
	overflowStyle = lipgloss.NewStyle().Foreground(muted).Background(background).Padding(0, 1)
}

func renderTabs(tabs []tab, active, width int) string {
	if width <= 0 {
		return ""
	}
	rendered := make([]string, 0, len(tabs))
	renderedWidths := make([]int, 0, len(tabs))
	for i, t := range tabs {
		var cell string
		if i == active {
			cell = activeTab.Render(" " + t.title + " ")
		} else if t.disabled {
			cell = disabledTab.Render(" " + t.title + " ")
		} else {
			cell = inactiveTab.Render(" " + t.title + " ")
		}
		rendered = append(rendered, cell)
		renderedWidths = append(renderedWidths, lipgloss.Width(cell))
	}

	total := 0
	for _, w := range renderedWidths {
		total += w
	}
	if total <= width {
		row := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
		return headerStyle.Width(width).Render(row)
	}

	left := active
	right := active
	used := renderedWidths[active]
	for {
		grew := false
		if left > 0 && used+renderedWidths[left-1] <= width {
			left--
			used += renderedWidths[left]
			grew = true
		}
		if right < len(tabs)-1 && used+renderedWidths[right+1] <= width {
			right++
			used += renderedWidths[right]
			grew = true
		}
		if !grew {
			break
		}
	}

	leftOverflow := left > 0
	rightOverflow := right < len(tabs)-1
	overflowWidth := 0
	if leftOverflow {
		overflowWidth += lipgloss.Width(overflowStyle.Render(" … "))
	}
	if rightOverflow {
		overflowWidth += lipgloss.Width(overflowStyle.Render(" … "))
	}

	for used+overflowWidth > width && (left < active || right > active) {
		if right > active && used+overflowWidth-renderedWidths[right] >= 0 {
			used -= renderedWidths[right]
			right--
		} else if left < active && used+overflowWidth-renderedWidths[left] >= 0 {
			used -= renderedWidths[left]
			left++
		} else {
			break
		}
		leftOverflow = left > 0
		rightOverflow = right < len(tabs)-1
		overflowWidth = 0
		if leftOverflow {
			overflowWidth += lipgloss.Width(overflowStyle.Render(" … "))
		}
		if rightOverflow {
			overflowWidth += lipgloss.Width(overflowStyle.Render(" … "))
		}
	}

	parts := make([]string, 0, (right-left)+3)
	if leftOverflow {
		parts = append(parts, overflowStyle.Render(" … "))
	}
	for i := left; i <= right; i++ {
		parts = append(parts, rendered[i])
	}
	if rightOverflow {
		parts = append(parts, overflowStyle.Render(" … "))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return headerStyle.Width(width).Render(row)
}

func renderFooter(status, spinner string, width int) string {
	help := "q:quit  tab/shift+tab:next/prev  up/down/pgup/pgdn:scroll  t:theme"
	if status != "" {
		help = spinner + "  " + status + "  |  " + help
	} else if spinner != "" {
		help = spinner + "  " + help
	}
	return footerStyle.Width(width).Render(help)
}

func isQuitKey(msg tea.KeyMsg) bool {
	if msg.Type == tea.KeyEsc {
		return true
	}
	if msg.Type == tea.KeyCtrlC {
		return true
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case 'q', 'Q':
			return true
		}
	}
	switch msg.String() {
	case "q", "Q", "esc", "ctrl+c":
		return true
	}
	return false
}

func renderSummary(history metricHistory, width int) string {
	parts := make([]string, 0, 4)
	if len(history.load) > 0 {
		max := maxFloat(history.load)
		if max < 1 {
			max = 1
		}
		parts = append(parts, fmt.Sprintf("LOAD %s %0.2f", sparkline(history.load, 0, max), history.load[len(history.load)-1]))
	}
	if len(history.cpu) > 0 {
		parts = append(parts, fmt.Sprintf("CPU %s %0.0f%%", sparkline(history.cpu, 0, 100), history.cpu[len(history.cpu)-1]))
	}
	if len(history.mem) > 0 {
		parts = append(parts, fmt.Sprintf("MEM %s %0.0f%%", sparkline(history.mem, 0, 100), history.mem[len(history.mem)-1]))
	}
	if len(history.net) > 0 {
		max := maxFloat(history.net)
		if max < 1 {
			max = 1
		}
		parts = append(parts, fmt.Sprintf("NET %s %s", sparkline(history.net, 0, max), formatRate(history.net[len(history.net)-1])))
	}
	row := strings.Join(parts, "  |  ")
	if row == "" {
		row = "METRICS unavailable (missing commands)"
	}
	return summaryStyle.Width(width).Render(row)
}

func renderInfoLine(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if strings.TrimSpace(text) == "" {
		text = " "
	}
	return infoStyle.Width(width).Render(text)
}

func renderContentTitle(title string, width int) string {
	if width <= 0 {
		return ""
	}
	label := fmt.Sprintf(" %s ", title)
	return summaryStyle.Width(width).Render(label)
}

func sparkline(values []float64, min, max float64) string {
	if len(values) == 0 {
		return ""
	}
	if max <= min {
		max = min + 1
	}
	levels := []rune(" .:-=+*#%@")
	var b strings.Builder
	for _, v := range values {
		if v < min {
			v = min
		}
		if v > max {
			v = max
		}
		n := int(((v - min) / (max - min)) * float64(len(levels)-1))
		if n < 0 {
			n = 0
		}
		if n >= len(levels) {
			n = len(levels) - 1
		}
		b.WriteRune(levels[n])
	}
	return b.String()
}

func updateHistory(history metricHistory, sample metricsSample) metricHistory {
	if sample.okLoad {
		history.load = append(history.load, sample.load)
		history.load = trimHistory(history.load, historyLength)
	}
	if sample.okCPU {
		history.cpu = append(history.cpu, sample.cpu)
		history.cpu = trimHistory(history.cpu, historyLength)
	}
	if sample.okMem {
		history.mem = append(history.mem, sample.mem)
		history.mem = trimHistory(history.mem, historyLength)
	}
	if sample.okNet {
		history.net = append(history.net, sample.netKB)
		history.net = trimHistory(history.net, historyLength)
	}
	return history
}

func trimHistory(values []float64, maxLen int) []float64 {
	if len(values) <= maxLen {
		return values
	}
	return values[len(values)-maxLen:]
}

func sampleMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		return metricsMsg{metrics: sampleMetrics()}
	}
}

func sampleSystemCmd() tea.Cmd {
	return func() tea.Msg {
		return systemMsg{info: sampleSystem()}
	}
}

func sampleMetrics() metricsSample {
	var sample metricsSample
	if load, ok := getLoadAvg(); ok {
		sample.load = load
		sample.okLoad = true
	}
	if cpu, ok := getCPUUsage(); ok {
		sample.cpu = cpu
		sample.okCPU = true
	}
	if mem, ok := getMemUsage(); ok {
		sample.mem = mem
		sample.okMem = true
	}
	if netKB, ok := getNetRateKB(); ok {
		sample.netKB = netKB
		sample.okNet = true
	}
	return sample
}

func sampleSystem() systemInfo {
	var info systemInfo
	load, _ := getLoadAvg()
	cpu, _ := getCPUUsage()
	mem, _ := getMemUsage()
	uptime := getUptimeShort()
	info.snapshot = fmt.Sprintf("Snapshot: CPU %0.0f%%  MEM %0.0f%%  LOAD %0.2f  UPTIME %s", cpu, mem, load, uptime)

	if disk := getDiskSummary(); disk != "" {
		info.disk = "Disk: " + disk
	}
	if net := getNetSummary(); net != "" {
		info.net = "Net: " + net
	}
	return info
}

func getUptimeShort() string {
	if _, err := exec.LookPath("uptime"); err != nil {
		return "unknown"
	}
	out, err := runQuickCmd([]string{"uptime"}, 2*time.Second)
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(out)
	idx := strings.Index(line, " up ")
	if idx == -1 {
		return "unknown"
	}
	part := line[idx+4:]
	if cut := strings.Index(part, "load average"); cut != -1 {
		part = part[:cut]
	}
	if cut := strings.Index(part, "load averages"); cut != -1 {
		part = part[:cut]
	}
	if cut := strings.Index(part, " user"); cut != -1 {
		part = part[:cut]
	}
	return strings.Trim(part, " ,")
}

func getDiskSummary() string {
	if _, err := exec.LookPath("df"); err != nil {
		return ""
	}
	out, err := runQuickCmd([]string{"df", "-h", "/"}, 2*time.Second)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return ""
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return ""
	}
	size := fields[1]
	used := fields[2]
	usePct := fields[4]
	return fmt.Sprintf("/ %s used %s (%s)", size, used, usePct)
}

func getNetSummary() string {
	rate, ok := getNetRateKB()
	if !ok {
		return ""
	}
	iface := getPrimaryIface()
	if iface == "" {
		iface = "iface"
	}
	return fmt.Sprintf("%s %s", iface, formatRate(rate))
}

func getPrimaryIface() string {
	if data, err := os.ReadFile("/proc/net/dev"); err == nil {
		if iface := firstIfaceLinux(data); iface != "" {
			return iface
		}
	}
	if _, err := exec.LookPath("netstat"); err == nil {
		if iface := firstIfaceDarwin(); iface != "" {
			return iface
		}
	}
	return ""
}

func firstIfaceLinux(data []byte) string {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" || strings.HasPrefix(iface, "lo") {
			continue
		}
		return iface
	}
	return ""
}

func firstIfaceDarwin() string {
	out, err := runQuickCmd([]string{"netstat", "-ib"}, 2*time.Second)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return ""
	}
	header := strings.Fields(lines[0])
	nIdx := indexOf(header, "Name")
	if nIdx == -1 {
		return ""
	}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) <= nIdx {
			continue
		}
		iface := fields[nIdx]
		if iface == "lo0" || strings.HasPrefix(iface, "lo") {
			continue
		}
		return iface
	}
	return ""
}

func getLoadAvg() (float64, bool) {
	if _, err := exec.LookPath("uptime"); err != nil {
		return 0, false
	}
	out, err := runQuickCmd([]string{"uptime"}, 2*time.Second)
	if err != nil {
		return 0, false
	}
	line := strings.TrimSpace(out)
	idx := strings.Index(line, "load average")
	if idx == -1 {
		idx = strings.Index(line, "load averages")
	}
	if idx == -1 {
		return 0, false
	}
	part := line[idx:]
	parts := strings.FieldsFunc(part, func(r rune) bool {
		return r == ':' || r == ','
	})
	if len(parts) < 2 {
		return 0, false
	}
	loadStr := strings.TrimSpace(parts[1])
	loadStr = strings.TrimSuffix(loadStr, ",")
	load, err := parseFloat(loadStr)
	if err != nil {
		return 0, false
	}
	return load, true
}

func getCPUUsage() (float64, bool) {
	if _, err := exec.LookPath("vmstat"); err == nil {
		if cpu, ok := cpuFromVmstat(); ok {
			return cpu, true
		}
	}
	if _, err := exec.LookPath("mpstat"); err == nil {
		if cpu, ok := cpuFromMpstat(); ok {
			return cpu, true
		}
	}
	return 0, false
}

func cpuFromVmstat() (float64, bool) {
	out, err := runQuickCmd([]string{"vmstat"}, 2*time.Second)
	if err != nil {
		return 0, false
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 3 {
		return 0, false
	}
	header := ""
	values := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if values == "" {
			values = lines[i]
			continue
		}
		header = lines[i]
		break
	}
	if header == "" || values == "" {
		return 0, false
	}
	hFields := strings.Fields(header)
	vFields := strings.Fields(values)
	if len(hFields) != len(vFields) {
		return 0, false
	}
	idx := indexOf(hFields, "id")
	if idx == -1 {
		return 0, false
	}
	idle, err := parseFloat(vFields[idx])
	if err != nil {
		return 0, false
	}
	cpu := 100 - idle
	if cpu < 0 {
		cpu = 0
	}
	if cpu > 100 {
		cpu = 100
	}
	return cpu, true
}

func cpuFromMpstat() (float64, bool) {
	out, err := runQuickCmd([]string{"mpstat", "1", "1"}, 3*time.Second)
	if err != nil {
		return 0, false
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "Linux") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if strings.ToLower(fields[1]) != "all" {
			continue
		}
		idleStr := fields[len(fields)-1]
		idle, err := parseFloat(idleStr)
		if err != nil {
			continue
		}
		cpu := 100 - idle
		if cpu < 0 {
			cpu = 0
		}
		if cpu > 100 {
			cpu = 100
		}
		return cpu, true
	}
	return 0, false
}

func getMemUsage() (float64, bool) {
	if _, err := exec.LookPath("free"); err != nil {
		return 0, false
	}
	out, err := runQuickCmd([]string{"free", "-m"}, 2*time.Second)
	if err != nil {
		return 0, false
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				return 0, false
			}
			total, err := parseFloat(fields[1])
			if err != nil || total == 0 {
				return 0, false
			}
			used, err := parseFloat(fields[2])
			if err != nil {
				return 0, false
			}
			return (used / total) * 100, true
		}
	}
	return 0, false
}

var netPrevTotal uint64
var netPrevAt time.Time

func getNetRateKB() (float64, bool) {
	total, ok := readNetBytes()
	if !ok {
		return 0, false
	}
	now := time.Now()
	if netPrevAt.IsZero() {
		netPrevAt = now
		netPrevTotal = total
		return 0, false
	}
	if total < netPrevTotal {
		netPrevAt = now
		netPrevTotal = total
		return 0, false
	}
	secs := now.Sub(netPrevAt).Seconds()
	if secs <= 0 {
		netPrevAt = now
		netPrevTotal = total
		return 0, false
	}
	delta := total - netPrevTotal
	netPrevAt = now
	netPrevTotal = total
	return float64(delta) / 1024.0 / secs, true
}

func readNetBytes() (uint64, bool) {
	if data, err := os.ReadFile("/proc/net/dev"); err == nil {
		if total, ok := sumNetBytesLinux(data); ok {
			return total, true
		}
	}
	if _, err := exec.LookPath("netstat"); err == nil {
		if total, ok := sumNetBytesDarwin(); ok {
			return total, true
		}
	}
	return 0, false
}

func sumNetBytesLinux(data []byte) (uint64, bool) {
	lines := strings.Split(string(data), "\n")
	var total uint64
	var found bool
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" || strings.HasPrefix(iface, "lo") {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}
		rx, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		tx, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			continue
		}
		total += rx + tx
		found = true
	}
	return total, found
}

func sumNetBytesDarwin() (uint64, bool) {
	out, err := runQuickCmd([]string{"netstat", "-ib"}, 2*time.Second)
	if err != nil {
		return 0, false
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return 0, false
	}
	header := strings.Fields(lines[0])
	iIdx := indexOf(header, "Ibytes")
	oIdx := indexOf(header, "Obytes")
	nIdx := indexOf(header, "Name")
	if iIdx == -1 || oIdx == -1 || nIdx == -1 {
		return 0, false
	}
	var total uint64
	var found bool
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) <= oIdx || len(fields) <= iIdx || len(fields) <= nIdx {
			continue
		}
		iface := fields[nIdx]
		if iface == "lo0" || strings.HasPrefix(iface, "lo") {
			continue
		}
		ib, err := strconv.ParseUint(fields[iIdx], 10, 64)
		if err != nil {
			continue
		}
		ob, err := strconv.ParseUint(fields[oIdx], 10, 64)
		if err != nil {
			continue
		}
		total += ib + ob
		found = true
	}
	return total, found
}

func runQuickCmd(cmd []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = &out
	if err := c.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func indexOf(fields []string, target string) int {
	for i, f := range fields {
		if f == target {
			return i
		}
	}
	return -1
}

func maxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	out := values[0]
	for _, v := range values[1:] {
		if v > out {
			out = v
		}
	}
	return out
}

func clampMin(value, min int) int {
	if value < min {
		return min
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func formatRate(kbPerSec float64) string {
	if kbPerSec < 1024 {
		return fmt.Sprintf("%0.0fKB/s", kbPerSec)
	}
	return fmt.Sprintf("%0.1fMB/s", kbPerSec/1024.0)
}
