package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zhengda-lu/pstop/internal/process"
)

// Tab represents the active view tab.
type Tab int

const (
	TabAll Tab = iota
	TabTop
	TabDev
)

// SortColumn represents the column to sort by.
type SortColumn int

const (
	SortCPU SortColumn = iota
	SortMem
	SortPID
	SortName
)

type tickMsg time.Time

type processMsg struct {
	processes []process.Info
	err       error
}

type devGroupMsg struct {
	groups []process.DevGroup
	err    error
}

type detailMsg struct {
	info *process.DetailedInfo
	err  error
}

type killResultMsg struct {
	pid int
	err error
}

// keyMap defines key bindings for the TUI.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Quit     key.Binding
	Kill     key.Binding
	Info     key.Binding
	Search   key.Binding
	Tab      key.Binding
	Help     key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Sort1    key.Binding
	Sort2    key.Binding
	Sort3    key.Binding
	Sort4    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "up")),
		Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "down")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Kill:     key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "kill")),
		Info:     key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "info")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch tab")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Confirm:  key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
		Cancel:   key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "cancel")),
		Sort1:    key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "sort CPU")),
		Sort2:    key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "sort MEM")),
		Sort3:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "sort PID")),
		Sort4:    key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "sort Name")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("PgUp", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("PgDn", "page down")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Kill, k.Info, k.Search, k.Tab, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Sort1, k.Sort2, k.Sort3, k.Sort4},
		{k.Kill, k.Info, k.Search, k.Tab},
		{k.Quit, k.Help},
	}
}

// Model is the Bubble Tea model for pstop.
type Model struct {
	version     string
	keys        keyMap
	help        help.Model
	width       int
	height      int
	tab         Tab
	sort        SortColumn
	cursor      int
	offset      int
	processes   []process.Info
	devGroups   []process.DevGroup
	filtered    []process.Info
	searching   bool
	searchInput textinput.Model
	filter      string
	confirming  bool
	confirmPID  int
	showDetail  bool
	detail      *process.DetailedInfo
	showHelp    bool
	err         error
	statusMsg   string
}

// New creates a new TUI model.
func New(version string) Model {
	ti := textinput.New()
	ti.Placeholder = "Search processes..."
	ti.CharLimit = 64

	return Model{
		version:     version,
		keys:        newKeyMap(),
		help:        help.New(),
		sort:        SortCPU,
		searchInput: ti,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchProcesses(tab Tab, sort SortColumn) tea.Cmd {
	return func() tea.Msg {
		var procs []process.Info
		var err error

		switch tab {
		case TabTop:
			procs, err = process.Top(20)
		default:
			procs, err = process.List()
		}
		if err != nil {
			return processMsg{err: err}
		}
		process.Sort(procs, sortColumnToField(sort))
		return processMsg{processes: procs}
	}
}

func fetchDevGroups() tea.Cmd {
	return func() tea.Msg {
		groups, err := process.GroupByStack()
		return devGroupMsg{groups: groups, err: err}
	}
}

func fetchDetail(pid int) tea.Cmd {
	return func() tea.Msg {
		info, err := process.GetInfo(pid)
		return detailMsg{info: info, err: err}
	}
}

func killProcess(pid int) tea.Cmd {
	return func() tea.Msg {
		err := process.Kill(pid, false)
		return killResultMsg{pid: pid, err: err}
	}
}

func sortColumnToField(s SortColumn) string {
	switch s {
	case SortMem:
		return "mem"
	case SortPID:
		return "pid"
	case SortName:
		return "name"
	default:
		return "cpu"
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchProcesses(m.tab, m.sort), tickCmd())
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tickMsg:
		if m.tab == TabDev {
			return m, tea.Batch(fetchDevGroups(), tickCmd())
		}
		return m, tea.Batch(fetchProcesses(m.tab, m.sort), tickCmd())

	case processMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.processes = msg.processes
		m.applyFilter()
		return m, nil

	case devGroupMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.devGroups = msg.groups
		return m, nil

	case detailMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.detail = msg.info
		m.showDetail = true
		return m, nil

	case killResultMsg:
		m.confirming = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Failed to kill PID %d: %v", msg.pid, msg.err)
		} else {
			m.statusMsg = fmt.Sprintf("Killed PID %d", msg.pid)
		}
		return m, fetchProcesses(m.tab, m.sort)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If searching, handle search input first.
	if m.searching {
		switch msg.String() {
		case "enter":
			m.searching = false
			m.filter = m.searchInput.Value()
			m.applyFilter()
			m.cursor = 0
			m.offset = 0
			return m, nil
		case "esc":
			m.searching = false
			m.filter = ""
			m.searchInput.SetValue("")
			m.applyFilter()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.filter = m.searchInput.Value()
			m.applyFilter()
			return m, cmd
		}
	}

	// If confirming kill.
	if m.confirming {
		switch {
		case key.Matches(msg, m.keys.Confirm):
			return m, killProcess(m.confirmPID)
		case key.Matches(msg, m.keys.Cancel):
			m.confirming = false
			return m, nil
		}
		return m, nil
	}

	// If showing detail.
	if m.showDetail {
		m.showDetail = false
		return m, nil
	}

	// If showing help.
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}

	case key.Matches(msg, m.keys.Down):
		max := m.listLen() - 1
		if m.cursor < max {
			m.cursor++
			viewHeight := m.tableHeight()
			if m.cursor >= m.offset+viewHeight {
				m.offset = m.cursor - viewHeight + 1
			}
		}

	case key.Matches(msg, m.keys.PageUp):
		viewHeight := m.tableHeight()
		m.cursor -= viewHeight
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.offset = m.cursor

	case key.Matches(msg, m.keys.PageDown):
		viewHeight := m.tableHeight()
		max := m.listLen() - 1
		m.cursor += viewHeight
		if m.cursor > max {
			m.cursor = max
		}
		if m.cursor >= m.offset+viewHeight {
			m.offset = m.cursor - viewHeight + 1
		}

	case key.Matches(msg, m.keys.Kill):
		if m.tab != TabDev && m.listLen() > 0 {
			proc := m.filtered[m.cursor]
			m.confirming = true
			m.confirmPID = proc.PID
		}

	case key.Matches(msg, m.keys.Info):
		if m.tab != TabDev && m.listLen() > 0 {
			proc := m.filtered[m.cursor]
			return m, fetchDetail(proc.PID)
		}

	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.searchInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Tab):
		m.tab = (m.tab + 1) % 3
		m.cursor = 0
		m.offset = 0
		if m.tab == TabDev {
			return m, fetchDevGroups()
		}
		return m, fetchProcesses(m.tab, m.sort)

	case key.Matches(msg, m.keys.Help):
		m.showHelp = true

	case key.Matches(msg, m.keys.Sort1):
		m.sort = SortCPU
		m.applyFilter()
	case key.Matches(msg, m.keys.Sort2):
		m.sort = SortMem
		m.applyFilter()
	case key.Matches(msg, m.keys.Sort3):
		m.sort = SortPID
		m.applyFilter()
	case key.Matches(msg, m.keys.Sort4):
		m.sort = SortName
		m.applyFilter()
	}

	return m, nil
}

func (m *Model) applyFilter() {
	if m.filter == "" {
		m.filtered = m.processes
		return
	}
	query := strings.ToLower(m.filter)
	var result []process.Info
	for _, p := range m.processes {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Command), query) ||
			strings.Contains(strings.ToLower(p.User), query) ||
			strings.Contains(fmt.Sprintf("%d", p.PID), query) {
			result = append(result, p)
		}
	}
	m.filtered = result
	process.Sort(m.filtered, sortColumnToField(m.sort))
}

func (m Model) listLen() int {
	if m.tab == TabDev {
		count := 0
		for _, g := range m.devGroups {
			count += 1 + len(g.Processes)
		}
		return count
	}
	return len(m.filtered)
}

func (m Model) tableHeight() int {
	// Header(1) + tab bar(1) + status(1) + help(2) + border padding(2)
	overhead := 7
	h := m.height - overhead
	if h < 1 {
		h = 10
	}
	return h
}

// View renders the TUI.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title bar.
	b.WriteString(titleStyle.Render(fmt.Sprintf("pstop %s", m.version)))
	b.WriteString("\n")

	// Tab bar.
	tabs := []string{"All", "Top", "Dev"}
	var tabParts []string
	for i, t := range tabs {
		if Tab(i) == m.tab {
			tabParts = append(tabParts, activeTabStyle.Render(fmt.Sprintf("[%s]", t)))
		} else {
			tabParts = append(tabParts, inactiveTabStyle.Render(fmt.Sprintf(" %s ", t)))
		}
	}
	b.WriteString(strings.Join(tabParts, " "))
	b.WriteString("\n")

	// Search bar.
	if m.searching {
		b.WriteString(m.searchInput.View())
		b.WriteString("\n")
	} else if m.filter != "" {
		b.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s", m.filter)))
		b.WriteString("\n")
	}

	// Confirm dialog.
	if m.confirming {
		b.WriteString(warnStyle.Render(fmt.Sprintf("Kill PID %d? (y/n)", m.confirmPID)))
		b.WriteString("\n")
		return b.String()
	}

	// Detail view.
	if m.showDetail && m.detail != nil {
		b.WriteString(m.renderDetail())
		return b.String()
	}

	// Help view.
	if m.showHelp {
		b.WriteString(m.help.View(m.keys))
		return b.String()
	}

	// Main content.
	if m.tab == TabDev {
		b.WriteString(m.renderDevView())
	} else {
		b.WriteString(m.renderProcessTable())
	}

	// Status bar.
	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Short help.
	b.WriteString(m.help.View(m.keys))

	return b.String()
}

func (m Model) renderProcessTable() string {
	var b strings.Builder

	// Header with sort indicator.
	sortIndicator := func(col SortColumn) string {
		if m.sort == col {
			return " v"
		}
		return ""
	}
	header := fmt.Sprintf("%-8s %-20s %-10s %8s%s %8s%s %-6s %-s",
		"PID", "NAME", "USER",
		"CPU%", sortIndicator(SortCPU),
		"MEM%", sortIndicator(SortMem),
		"STATE", "COMMAND",
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	viewHeight := m.tableHeight()
	end := m.offset + viewHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.offset; i < end; i++ {
		p := m.filtered[i]
		line := fmt.Sprintf("%-8d %-20s %-10s %8.1f %8.1f %-6s %-s",
			p.PID, truncate(p.Name, 20), truncate(p.User, 10),
			p.CPU, p.Mem, p.State, truncate(p.Command, 40),
		)

		switch {
		case i == m.cursor:
			line = selectedStyle.Render(line)
		case p.CPU > 50:
			line = highCPUStyle.Render(line)
		case p.CPU > 20:
			line = medCPUStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Footer with count.
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %d processes", len(m.filtered))))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderDevView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Developer Process Groups"))
	b.WriteString("\n\n")

	for _, g := range m.devGroups {
		b.WriteString(groupStyle.Render(fmt.Sprintf("%s (%d processes) - CPU: %.1f%% MEM: %.1f%%",
			g.Stack, len(g.Processes), g.TotalCPU, g.TotalMem)))
		b.WriteString("\n")

		for _, p := range g.Processes {
			b.WriteString(fmt.Sprintf("  %-8d %-20s %6.1f%% %6.1f%%\n",
				p.PID, truncate(p.Name, 20), p.CPU, p.Mem))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderDetail() string {
	var b strings.Builder

	d := m.detail
	b.WriteString(titleStyle.Render(fmt.Sprintf("Process Details: %s (PID %d)", d.Name, d.PID)))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("User:       "))
	b.WriteString(d.User)
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("CPU:        "))
	b.WriteString(fmt.Sprintf("%.1f%%", d.CPU))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Memory:     "))
	b.WriteString(fmt.Sprintf("%.1f%%", d.Mem))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Open Files: "))
	b.WriteString(fmt.Sprintf("%d", d.OpenFiles))
	b.WriteString("\n")

	if len(d.Ports) > 0 {
		b.WriteString(labelStyle.Render("Ports:      "))
		ports := make([]string, len(d.Ports))
		for i, p := range d.Ports {
			ports[i] = fmt.Sprintf("%d", p)
		}
		b.WriteString(strings.Join(ports, ", "))
		b.WriteString("\n")
	}

	if len(d.Children) > 0 {
		b.WriteString(labelStyle.Render("Children:   "))
		children := make([]string, len(d.Children))
		for i, c := range d.Children {
			children[i] = fmt.Sprintf("%d", c)
		}
		b.WriteString(strings.Join(children, ", "))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Press any key to return"))

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "~"
}
