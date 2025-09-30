package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"valorantsecurecheck/pkg/system"
)

type result struct {
	TPM        system.TPMInfo
	SecureBoot system.SecureBoot
	System     system.SystemInfo
	Checks     map[string]bool
	CanRun     bool
}

// CLI flags
var (
	flagJSON     = flag.Bool("json", false, "Print pretty JSON and exit")
	flagTable    = flag.Bool("table", false, "Print a simple table instead of TUI")
	flagExitCode = flag.Bool("exit-code", false, "Set process exit code: 0 if OK for Valorant, 1 otherwise")
	flagVerbose  = flag.Bool("v", false, "Print warnings to stderr (TUI shows none)")
)

// TUI model
type model struct {
	res      result
	width    int
	height   int
	ready    bool
	showJSON bool
	json     string
	vp       viewport.Model
}

func newModel(res result) model { return model{res: res} }
func (m model) Init() tea.Cmd   { return nil }

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
			m.vp = viewport.New(m.width-10, m.height-10)
			m.vp.SetContent(m.json)
		} else {
			m.vp.Width = m.width - 10
			m.vp.Height = m.height - 10
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Loading…"
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Render("ValorantSecureCheck")

	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Press J to toggle JSON • Q to quit")

	author := lipgloss.NewStyle().
		   Foreground(lipgloss.Color("14")).
		   Render("Author: ImElio")

	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	noStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Bold(true)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)

	status := func(b bool) string {
		if b {
			return okStyle.Render("✓ OK")
		}
		return noStyle.Render("✗ Not OK")
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
		fmt.Sprintf("%s  %v / %v", keyStyle.Render("TPM Present/Ready"), m.res.TPM.Present, m.res.TPM.Ready),
		fmt.Sprintf("%s  %s", keyStyle.Render("TPM Version"), valStyle.Render(tpmVer)),
		fmt.Sprintf("%s  %s", keyStyle.Render("TPM Vendor"), valStyle.Render(m.res.TPM.Vendor)),
		fmt.Sprintf("%s  %v", keyStyle.Render("Secure Boot Enabled"), m.res.SecureBoot.Enabled),
		fmt.Sprintf("%s  %s", keyStyle.Render("CPU"), valStyle.Render(m.res.System.CPU)),
		fmt.Sprintf("%s  %s", keyStyle.Render("GPU"), valStyle.Render(m.res.System.GPU)),
		fmt.Sprintf("%s  %d", keyStyle.Render("RAM (GiB)"), m.res.System.RAMGiB),
		fmt.Sprintf("%s  %s", keyStyle.Render("Motherboard"), valStyle.Render(m.res.System.Motherboard)),
		fmt.Sprintf("%s  %s", keyStyle.Render("OS"), valStyle.Render(m.res.System.OS)),
	}
	det := lipgloss.JoinVertical(lipgloss.Left, details...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(min(70, m.width-10))

	checkBox := box.Render(checks)
	detBox := box.Render(det)
	mainCol := lipgloss.JoinVertical(lipgloss.Left, checkBox, "", detBox)

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	flag.Parse()

	// Gather data once
	tpm, tpmErr := system.GetTPMInfo()
	sb, sbErr := system.CheckSecureBoot()
	sys, sysErr := system.GetSystemInfo()

	// Compute checks
	checks := map[string]bool{
		"TPM2":        tpm.Present && tpm.Ready && tpm.IsV2,
		"SecureBoot":  sb.Enabled,
		"CPU":         sys.CPU != "",
		"GPU":         sys.GPU != "",
		"RAM>=4GiB":   sys.RAMGiB >= 4,
		"Motherboard": sys.Motherboard != "",
	}
	canRun := checks["TPM2"] && checks["SecureBoot"] && checks["CPU"] && checks["GPU"] && checks["RAM>=4GiB"]

	res := result{TPM: tpm, SecureBoot: sb, System: sys, Checks: checks, CanRun: canRun}

	if *flagVerbose {
		if tpmErr != nil {
			fmt.Fprintf(os.Stderr, "[warn] TPM check error: %v\n", tpmErr)
		}
		if sbErr != nil {
			fmt.Fprintf(os.Stderr, "[warn] SecureBoot check error: %v\n", sbErr)
		}
		if sysErr != nil {
			fmt.Fprintf(os.Stderr, "[warn] System info error: %v\n", sysErr)
		}
	}

	if *flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		if *flagExitCode {
			if canRun {
				os.Exit(0)
			}
			os.Exit(1)
		}
		return
	}

	if *flagTable {
		fmt.Println("+--------------+---------+")
		fmt.Printf("| %-12s | %-7s |\n", "CHECK", "STATUS")
		fmt.Println("+--------------+---------+")
		printRow("TPM 2.0", checks["TPM2"])
		printRow("Secure Boot", checks["SecureBoot"])
		printRow("CPU", checks["CPU"])
		printRow("GPU", checks["GPU"])
		printRow("RAM ≥ 4 GiB", checks["RAM>=4GiB"])
		printRow("Motherboard", checks["Motherboard"])
		fmt.Println("+--------------+---------+")
		if *flagExitCode {
			if canRun {
				os.Exit(0)
			}
			os.Exit(1)
		}
		return
	}

	prettyJSON, _ := json.MarshalIndent(res, "", "  ")
	m := newModel(res)
	m.json = string(prettyJSON)
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "TUI error:", err)
		os.Exit(1)
	}

	if *flagExitCode {
		if canRun {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func printRow(label string, ok bool) {
	status := "✗ Not OK"
	if ok {
		status = "✓ OK"
	}
	fmt.Printf("| %-12s | %-7s |\n", label, status)
}
