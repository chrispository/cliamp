// Command cliamp-lite is a stripped-down cliamp: one radio station (9128.live)
// and the built-in visualizers. It has no playlists, providers, plugins, IPC
// or config. The only keys are the arrows to move, Enter to play or stop, and
// q or Ctrl+C to quit.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/bjarneo/cliamp/player"
	"github.com/bjarneo/cliamp/ui"
)

const (
	stationName = "9128.live"
	stationURL  = "https://streams.radio.co/s0aa1e6f4a/listen"
	listWidth   = 22
)

type tickMsg time.Time

type playResultMsg struct{ err error }

type model struct {
	p       *player.Player
	vis     *ui.Visualizer
	modes   []string // visualizer names, VisNone excluded
	cursor  int      // 0 is the station, 1.. are visualizers
	scroll  int
	width   int
	height  int
	loading bool
	err     error
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd { return tick(ui.TickAnim) }

func (m model) playing() bool { return m.p.IsPlaying() && !m.p.IsPaused() }

func (m model) togglePlay() (model, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	if m.p.IsPlaying() {
		m.p.Stop()
		return m, nil
	}
	m.loading, m.err = true, nil
	p := m.p
	return m, func() tea.Msg { return playResultMsg{err: p.PlayAt(stationURL, 0, 0)} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.vis.RequestRefresh()
	case playResultMsg:
		m.loading, m.err = false, msg.err
	case tickMsg:
		m.vis.Tick(m.tickContext(time.Time(msg)))
		return m, tick(ui.TickAnim)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m.cursor = max(0, m.cursor-1)
		case "down":
			m.cursor = min(len(m.modes), m.cursor+1)
		case "enter", "space":
			if m.cursor == 0 {
				return m.togglePlay()
			}
		}
		if m.cursor > 0 {
			m.vis.SetMode(ui.VisMode(m.cursor - 1))
		}
	}
	m.clampScroll()
	return m, nil
}

// headerRows is the number of lines above the visualizer list in View.
const headerRows = 7

func (m model) listRows() int { return max(1, m.height-headerRows-1) }

// clampScroll keeps the selected visualizer inside the visible list rows.
func (m *model) clampScroll() {
	rows, sel := m.listRows(), max(0, m.cursor-1)
	if sel < m.scroll {
		m.scroll = sel
	} else if sel >= m.scroll+rows {
		m.scroll = sel - rows + 1
	}
	m.scroll = max(0, min(m.scroll, len(m.modes)-rows))
}

func (m model) tickContext(now time.Time) ui.VisTickContext {
	playing := m.playing()
	return ui.VisTickContext{
		Now:     now,
		Playing: playing,
		Analyze: func(spec ui.VisAnalysisSpec) []float64 {
			spec = ui.NormalizeAnalysisSpec(spec)
			if !playing {
				return m.vis.Analyze(nil, spec)
			}
			buf := m.vis.EnsureSampleBuf(spec.FFTSize)
			var n int
			if spec.Tap == ui.VisTapAudible {
				n = m.p.WaveformSamplesInto(buf)
			} else {
				n = m.p.SamplesInto(buf)
			}
			return m.vis.Analyze(buf[max(0, n-spec.FFTSize):n], spec)
		},
		StereoSamplesInto: func(dst [][2]float64) int {
			if !playing {
				return 0
			}
			return m.p.StereoSamplesInto(dst)
		},
	}
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorTitle)
	dimStyle    = lipgloss.NewStyle().Foreground(ui.ColorDim)
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorAccent)
	activeStyle = lipgloss.NewStyle().Foreground(ui.ColorPlaying)
	errStyle    = lipgloss.NewStyle().Foreground(ui.ColorError)
)

func (m model) row(i int, label string, active bool) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if active {
		style = activeStyle
	}
	if i == m.cursor {
		prefix = "> "
		style = cursorStyle
	}
	return style.Render(prefix + label)
}

func (m model) View() tea.View {
	if m.width < listWidth+20 || m.height < 8 {
		v := tea.NewView("Terminal too small.")
		v.AltScreen = true
		return v
	}

	// Station block.
	state := "stopped"
	switch {
	case m.loading:
		state = "connecting..."
	case m.playing():
		state = "playing"
	}
	header := []string{
		titleStyle.Render("CLIAMP LITE"),
		"",
		dimStyle.Render("STATION"),
		m.row(0, stationName, m.p.IsPlaying()),
		dimStyle.Render("  " + state),
		"",
		dimStyle.Render("VISUALIZERS"),
	}

	// Scrollable visualizer list fills the rest of the height.
	lines := header
	for i := m.scroll; i < len(m.modes) && i < m.scroll+m.listRows(); i++ {
		lines = append(lines, m.row(i+1, m.modes[i], ui.VisMode(i) == m.vis.Mode))
	}
	left := lipgloss.NewStyle().Width(listWidth).Height(m.height - 1).Render(strings.Join(lines, "\n"))

	// Visualizer fills the right side, with the ICY title above it.
	visCols := m.width - listWidth - 2
	m.vis.Cols = visCols
	m.vis.Rows = max(1, m.height-4)
	nowPlaying := m.p.StreamTitle()
	if m.err != nil {
		nowPlaying = errStyle.Render(m.err.Error())
	}
	right := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(m.vis.ModeName()),
		ansiTrunc(nowPlaying, visCols),
		"",
		m.vis.Render(),
	)

	footer := dimStyle.Render("↑/↓ move  enter play/stop  q quit")
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	v := tea.NewView(ui.FitRect(body+"\n"+footer, m.width, m.height))
	v.AltScreen = true
	v.BackgroundColor = ui.ColorBackground
	return v
}

func ansiTrunc(s string, w int) string {
	return lipgloss.NewStyle().MaxWidth(w).Render(s)
}

func main() {
	sr := player.DeviceSampleRate()
	if sr <= 0 {
		sr = 44100
	}
	p, err := player.New(player.Quality{SampleRate: sr, BufferMs: 250, ResampleQuality: 4, BitDepth: 16})
	if err != nil {
		fmt.Fprintln(os.Stderr, "cliamp-lite:", err)
		os.Exit(1)
	}
	defer p.Close()

	vis := ui.NewVisualizer(float64(sr))
	modes := ui.VisModeNames()
	modes = modes[:len(modes)-1] // drop "None"

	m := model{p: p, vis: vis, modes: modes, cursor: 1}
	vis.SetMode(0)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "cliamp-lite:", err)
		os.Exit(1)
	}
}
