package cli

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)


func RunTUI(res Result) {
	pretty, _ := json.MarshalIndent(res, "", "  ")
	m := newModel(res, string(pretty))
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("TUI error:", err)
	}
}


type model struct {
	res      Result
	json     string
	showJSON bool
	ready    bool
	width    int
	height   int
	vp       viewport.Model
}

func newModel(res Result, jsonPretty string) model { return model{res: res, json: jsonPretty} }
func (m model) Init() tea.Cmd                      { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j":
			m.showJSON = !m.showJSON
			if m.showJSON {
				m.vp.SetContent(m.json)
			}
			return m, nil
		case "up", "k":
			if m.showJSON {
				m.vp.LineUp(1)
			}
			return m, nil
		case "down":
			if m.showJSON {
				m.vp.LineDown(1)
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.ready {
			m.ready = true
			w, h := max(m.width-10, 10), max(m.height-10, 5)
			m.vp = viewport.New(w, h)
			m.vp.SetContent(m.json)
		} else {
			m.vp.Width = max(m.width-10, 10)
			m.vp.Height = max(m.height-10, 5)
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Loading…"
	}

	title := styleTitle().Render("ValorantSecureCheck")
	sub := styleSub().Render("Press J to toggle JSON • Q to quit")
	author := styleAuthor().Render("Author: ImElio")

	status := func(b bool) string {
		if b {
			return styleOK().Render("✓ OK")
		}
		return styleNO().Render("✗ Not OK")
	}

	// Checks panel
	lines := []string{
		fmt.Sprintf("%-14s %s", "TPM 2.0", status(m.res.Checks["TPM2"])),
		fmt.Sprintf("%-14s %s", "Secure Boot", status(m.res.Checks["SecureBoot"])),
		fmt.Sprintf("%-14s %s", "CPU", status(m.res.Checks["CPU"])),
		fmt.Sprintf("%-14s %s", "GPU", status(m.res.Checks["GPU"])),
		fmt.Sprintf("%-14s %s", "RAM ≥ 4 GiB", status(m.res.Checks["RAM>=4GiB"])),
		fmt.Sprintf("%-14s %s", "Motherboard", status(m.res.Checks["Motherboard"])),
	}
	checks := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Details panel
	tpmVer := m.res.TPM.Version
	if tpmVer == "" && m.res.TPM.IsV2 {
		tpmVer = "2.0"
	}
	details := []string{
		fmt.Sprintf("%s  %v / %v", styleKey().Render("TPM Present/Ready"), m.res.TPM.Present, m.res.TPM.Ready),
		fmt.Sprintf("%s  %s", styleKey().Render("TPM Version"), styleVal().Render(tpmVer)),
		fmt.Sprintf("%s  %s", styleKey().Render("TPM Vendor"), styleVal().Render(m.res.TPM.Vendor)),
		fmt.Sprintf("%s  %v", styleKey().Render("Secure Boot Enabled"), m.res.SecureBoot.Enabled),
		fmt.Sprintf("%s  %s", styleKey().Render("CPU"), styleVal().Render(m.res.System.CPU)),
		fmt.Sprintf("%s  %s", styleKey().Render("GPU"), styleVal().Render(m.res.System.GPU)),
		fmt.Sprintf("%s  %d", styleKey().Render("RAM (GiB)"), m.res.System.RAMGiB),
		fmt.Sprintf("%s  %s", styleKey().Render("Motherboard"), styleVal().Render(m.res.System.Motherboard)),
		fmt.Sprintf("%s  %s", styleKey().Render("OS"), styleVal().Render(m.res.System.OS)),
	}
	det := lipgloss.JoinVertical(lipgloss.Left, details...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(min(70, m.width-10))

	mainCol := lipgloss.JoinVertical(lipgloss.Left, box.Render(checks), "", box.Render(det))

	content := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, sub, author, "", mainCol),
	)

	if m.showJSON {
		jsonTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).
			Render("JSON (VIEW)  —  ↑/↓ scroll, J to close")
		overlay := lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				jsonTitle,
				lipgloss.NewStyle().
					Border(lipgloss.NormalBorder()).
					Padding(0, 1).
					Render(m.vp.View()),
			),
		)
		return overlay
	}
	return content
}


func styleTitle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) }
func styleSub() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) }
func styleAuthor() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("14")) }
func styleOK() lipgloss.Style     { return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) }
func styleNO() lipgloss.Style     { return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) }
func styleKey() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Bold(true) }
func styleVal() lipgloss.Style    { return lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true) }


func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
