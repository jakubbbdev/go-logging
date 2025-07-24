package logging

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dashboard represents a real-time logging dashboard
type Dashboard struct {
	logger        Logger
	metrics       *MetricsCollector
	healthMonitor *HealthMonitor
	mu            sync.RWMutex
	stats         DashboardStats
	started       time.Time
}

// DashboardStats holds statistics for the dashboard
type DashboardStats struct {
	TotalLogs    int64
	LogsByLevel  map[Level]int64
	ErrorRate    float64
	AvgDuration  time.Duration
	LastLogTime  time.Time
	RecentLogs   []LogEntry
	HealthStatus HealthStatus
	Uptime       time.Duration
}

// LogEntry represents a recent log entry for the dashboard
type LogEntry struct {
	Timestamp time.Time
	Level     Level
	Message   string
	Fields    Fields
}

// NewDashboard creates a new dashboard
func NewDashboard(logger Logger, metrics *MetricsCollector, health *HealthMonitor) *Dashboard {
	return &Dashboard{
		logger:        logger,
		metrics:       metrics,
		healthMonitor: health,
		stats: DashboardStats{
			LogsByLevel: make(map[Level]int64),
			RecentLogs:  make([]LogEntry, 0, 10),
		},
		started: time.Now(),
	}
}

// Update updates dashboard statistics
func (d *Dashboard) Update() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update from metrics
	if d.metrics != nil {
		stats := d.metrics.GetStats()
		d.stats.LogsByLevel = stats.LogCount
		d.stats.LastLogTime = stats.LastLogTime

		// Calculate total logs
		total := int64(0)
		for _, count := range stats.LogCount {
			total += count
		}
		d.stats.TotalLogs = total

		// Calculate error rate
		errors := stats.LogCount[ErrorLevel] + stats.LogCount[FatalLevel] + stats.LogCount[PanicLevel]
		if total > 0 {
			d.stats.ErrorRate = float64(errors) / float64(total) * 100
		}
	}

	// Update from health monitor
	if d.healthMonitor != nil {
		d.stats.HealthStatus = d.healthMonitor.GetOverallHealth()
	}

	// Update uptime
	d.stats.Uptime = time.Since(d.started)
}

// AddRecentLog adds a recent log entry
func (d *Dashboard) AddRecentLog(level Level, message string, fields Fields) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	// Add to recent logs (keep only last 10)
	d.stats.RecentLogs = append([]LogEntry{entry}, d.stats.RecentLogs...)
	if len(d.stats.RecentLogs) > 10 {
		d.stats.RecentLogs = d.stats.RecentLogs[:10]
	}
}

// GetStats returns current dashboard statistics
func (d *Dashboard) GetStats() DashboardStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Make a copy
	stats := d.stats
	stats.LogsByLevel = make(map[Level]int64)
	for k, v := range d.stats.LogsByLevel {
		stats.LogsByLevel[k] = v
	}

	recentCopy := make([]LogEntry, len(d.stats.RecentLogs))
	copy(recentCopy, d.stats.RecentLogs)
	stats.RecentLogs = recentCopy

	return stats
}

// DashboardModel implements the Bubble Tea model for the dashboard
type DashboardModel struct {
	dashboard *Dashboard
	table     table.Model
	width     int
	height    int
	styles    DashboardStyles
}

// DashboardStyles defines the styling for the dashboard
type DashboardStyles struct {
	Base        lipgloss.Style
	Header      lipgloss.Style
	StatusOK    lipgloss.Style
	StatusWarn  lipgloss.Style
	StatusError lipgloss.Style
	LogDebug    lipgloss.Style
	LogInfo     lipgloss.Style
	LogWarn     lipgloss.Style
	LogError    lipgloss.Style
	LogFatal    lipgloss.Style
	Metric      lipgloss.Style
	Border      lipgloss.Style
}

// NewDashboardStyles creates default styles for the dashboard
func NewDashboardStyles() DashboardStyles {
	return DashboardStyles{
		Base: lipgloss.NewStyle().
			Padding(1, 2),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1),
		StatusOK: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true),
		StatusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true),
		LogDebug: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),
		LogInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		LogWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")),
		LogError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")),
		LogFatal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true),
		Metric: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),
		Border: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")),
	}
}

// NewDashboardModel creates a new dashboard model for Bubble Tea
func NewDashboardModel(dashboard *Dashboard) DashboardModel {
	columns := []table.Column{
		{Title: "Time", Width: 12},
		{Title: "Level", Width: 8},
		{Title: "Message", Width: 50},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(7),
	)

	return DashboardModel{
		dashboard: dashboard,
		table:     t,
		styles:    NewDashboardStyles(),
	}
}

// Init implements the Bubble Tea model interface
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		updateTick(),
		tea.EnterAltScreen,
	)
}

// Update implements the Bubble Tea model interface
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Refresh
			m.dashboard.Update()
		}

	case updateTickMsg:
		m.dashboard.Update()
		m.updateTable()
		return m, updateTick()
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implements the Bubble Tea model interface
func (m DashboardModel) View() string {
	stats := m.dashboard.GetStats()

	// Header
	header := m.styles.Header.Render("🚀 Go Logging Dashboard")

	// Status section
	statusStyle := m.styles.StatusOK
	statusText := "🟢 HEALTHY"
	switch stats.HealthStatus {
	case HealthStatusDegraded:
		statusStyle = m.styles.StatusWarn
		statusText = "🟡 DEGRADED"
	case HealthStatusUnhealthy:
		statusStyle = m.styles.StatusError
		statusText = "🔴 UNHEALTHY"
	}

	status := statusStyle.Render(statusText)

	// Metrics section
	metrics := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderMetric("Total Logs", fmt.Sprintf("%d", stats.TotalLogs)),
		m.renderMetric("Error Rate", fmt.Sprintf("%.1f%%", stats.ErrorRate)),
		m.renderMetric("Uptime", formatDuration(stats.Uptime)),
	)

	// Level breakdown
	levelBreakdown := m.renderLevelBreakdown(stats.LogsByLevel)

	// Recent logs table
	recentLogs := m.styles.Border.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.Header.Render("📋 Recent Logs"),
			m.table.View(),
		),
	)

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, "Status: ", status),
		"",
		metrics,
		"",
		levelBreakdown,
		"",
		recentLogs,
		"",
		m.styles.Base.Render("Press 'q' to quit, 'r' to refresh"),
	)

	return m.styles.Base.Render(content)
}

// renderMetric renders a metric with styling
func (m DashboardModel) renderMetric(label, value string) string {
	return m.styles.Border.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			m.styles.Header.Render(label),
			m.styles.Metric.Render(value),
		),
	)
}

// renderLevelBreakdown renders the log level breakdown
func (m DashboardModel) renderLevelBreakdown(levels map[Level]int64) string {
	var breakdown strings.Builder
	breakdown.WriteString(m.styles.Header.Render("📊 Log Levels"))
	breakdown.WriteString("\n")

	// Sort levels by count
	type levelCount struct {
		level Level
		count int64
	}

	var sorted []levelCount
	for level, count := range levels {
		sorted = append(sorted, levelCount{level, count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	for _, lc := range sorted {
		if lc.count == 0 {
			continue
		}

		style := m.styles.LogInfo
		switch lc.level {
		case DebugLevel:
			style = m.styles.LogDebug
		case WarnLevel:
			style = m.styles.LogWarn
		case ErrorLevel:
			style = m.styles.LogError
		case FatalLevel, PanicLevel:
			style = m.styles.LogFatal
		}

		breakdown.WriteString(fmt.Sprintf("  %s: %s\n",
			style.Render(strings.ToUpper(lc.level.String())),
			m.styles.Metric.Render(strconv.FormatInt(lc.count, 10))))
	}

	return m.styles.Border.Render(breakdown.String())
}

// updateTable updates the table with recent logs
func (m *DashboardModel) updateTable() {
	stats := m.dashboard.GetStats()

	var rows []table.Row
	for _, log := range stats.RecentLogs {
		// Truncate message if too long
		message := log.Message
		if len(message) > 47 {
			message = message[:47] + "..."
		}

		rows = append(rows, table.Row{
			log.Timestamp.Format("15:04:05"),
			strings.ToUpper(log.Level.String()),
			message,
		})
	}

	m.table.SetRows(rows)
}

// updateTickMsg represents a tick message for updates
type updateTickMsg struct{}

// updateTick returns a command that sends a tick message
func updateTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return updateTickMsg{}
	})
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// DashboardHook creates a hook that feeds data to the dashboard
func NewDashboardHook(dashboard *Dashboard) Hook {
	return func(entry *Entry) {
		dashboard.AddRecentLog(entry.Level, entry.Message, entry.Fields)
	}
}

// StartDashboard starts the dashboard UI
func StartDashboard(dashboard *Dashboard) error {
	model := NewDashboardModel(dashboard)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err := p.Run()
	return err
}

// DashboardHandler wraps another handler and feeds data to the dashboard
type DashboardHandler struct {
	handler   Handler
	dashboard *Dashboard
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(handler Handler, dashboard *Dashboard) Handler {
	return &DashboardHandler{
		handler:   handler,
		dashboard: dashboard,
	}
}

// Handle implements the Handler interface with dashboard integration
func (dh *DashboardHandler) Handle(entry *Entry) error {
	// Feed to dashboard
	dh.dashboard.AddRecentLog(entry.Level, entry.Message, entry.Fields)

	// Handle normally
	return dh.handler.Handle(entry)
}
