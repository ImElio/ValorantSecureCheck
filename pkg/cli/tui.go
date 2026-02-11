package cli

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type clearReportMsg struct{}

func RunTUI(res Result) {
	m := newModel(res)
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("TUI error:", err)
	}
}

type model struct {
	res        Result
	w, h       int
	reportPath string
}

func newModel(res Result) model { return model{res: res} }
func (m model) Init() tea.Cmd   { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j":
			if p, err := ExportJSONToFileAndOpen(m.res); err == nil {
				m.reportPath = p
			} else {
				m.reportPath = "Export failed"
			}
			return m, tea.Tick(4*time.Second, func(time.Time) tea.Msg { return clearReportMsg{} })
		}

	case clearReportMsg:
		m.reportPath = ""
		return m, nil

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.w == 0 || m.h == 0 {
		return "Loading…"
	}

	padV, padH := 1, 2
	if m.h < 28 {
		padV, padH = 0, 1
	}

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle().Render("ValorantSecureCheck"),
		subStyle().Render("J = export report (opens Notepad)  •  Q = quit"),
	)
	if m.reportPath != "" {
		header = lipgloss.JoinVertical(lipgloss.Left, header, hintStyle().Render("Report: "+m.reportPath))
	}

	body := m.renderBody(padV, padH)

	page := lipgloss.NewStyle().Padding(padV, padH).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, "", body),
	)

	return fitHeight(page, m.h)
}

func (m model) renderBody(padV, padH int) string {
	gap := 3

	safe := 2 // prevents right border clipping in conhost
	innerW := m.w - padH*2 - safe
	innerH := m.h - padV*2

	if innerW < 40 {
		innerW = 40
	}
	if innerH < 10 {
		innerH = 10
	}

	mode := "full"
	if innerH < 30 {
		mode = "compact"
	}
	if innerH < 20 {
		mode = "minimal"
	}

	minCol := 60
	colW := (innerW - gap) / 2
	stack := colW < minCol || innerW < (minCol*2+gap)

	if mode == "minimal" {
		boxW := max(innerW-1, 30) // -1 to avoid edge overflow
		checksBox := boxStyle().Width(boxW)
		checks := m.renderChecks(boxW - 6)

		hint := hintStyle().Render("Tip: press J to export the full report.")
		return checksBox.Render(sectionStyle().Render("Checks") + "\n\n" + checks + "\n\n" + hint)
	}

	if stack {
		boxW := max(innerW-1, 30) // -1 to avoid edge overflow
		checksBox := boxStyle().Width(boxW)
		detailsBox := boxStyle().Width(boxW)

		left := checksBox.Render(sectionStyle().Render("Checks") + "\n\n" + m.renderChecks(boxW-6))
		right := detailsBox.Render(sectionStyle().Render("Details") + "\n\n" + m.renderDetails(boxW-6, mode == "full"))

		return lipgloss.JoinVertical(lipgloss.Left, left, "", right)
	}

	rightW := colW - 3
	if rightW < 40 {
		rightW = 40
	}
	leftW := innerW - gap - rightW
	if leftW < 40 {
		boxW := max(innerW-1, 30)
		checksBox := boxStyle().Width(boxW)
		detailsBox := boxStyle().Width(boxW)

		left := checksBox.Render(sectionStyle().Render("Checks") + "\n\n" + m.renderChecks(boxW-6))
		right := detailsBox.Render(sectionStyle().Render("Details") + "\n\n" + m.renderDetails(boxW-6, mode == "full"))
		return lipgloss.JoinVertical(lipgloss.Left, left, "", right)
	}

	checksBox := boxStyle().Width(leftW)
	detailsBox := boxStyle().Width(rightW)

	left := checksBox.Render(sectionStyle().Render("Checks") + "\n\n" + m.renderChecks(leftW-6))
	right := detailsBox.Render(sectionStyle().Render("Details") + "\n\n" + m.renderDetails(rightW-6, mode == "full"))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
}

func (m model) renderChecks(wrapW int) string {
	ok := func(b bool) string {
		if b {
			return okStyle().Render("✓")
		}
		return noStyle().Render("✗")
	}

	readyLine := ""
	if m.res.CanRun {
		readyLine = okStyle().Render("READY") + "  Valorant should launch"
	} else {
		readyLine = noStyle().Render("NOT READY") + "  Missing requirement(s)"
	}

	core := []string{
		fmt.Sprintf("%s TPM 2.0", ok(m.res.Checks["TPM2"])),
		fmt.Sprintf("%s Secure Boot", ok(m.res.Checks["SecureBoot"])),
		fmt.Sprintf("%s UEFI", ok(m.res.Checks["UEFI"])),
		fmt.Sprintf("%s Disk GPT", ok(m.res.Checks["GPT"])),
		fmt.Sprintf("%s Vanguard installed", ok(m.res.Checks["Vanguard"])),
		fmt.Sprintf("%s vgc service exists", ok(m.res.Checks["VGCExists"])),
	}

	diag := []string{
		fmt.Sprintf("%s vgc running", ok(m.res.Checks["VGCRunning"])),
		fmt.Sprintf("%s vgk exists", ok(m.res.Checks["VGKExists"])),
		fmt.Sprintf("%s VBS disabled", ok(m.res.Checks["VBSDisabled"])),
		fmt.Sprintf("%s Hyper-V disabled", ok(m.res.Checks["HyperVOff"])),
		fmt.Sprintf("%s Secure Boot keys", ok(m.res.Checks["SBKeys"])),
	}

	block := strings.Join([]string{
		sectionStyle().Render("Valorant requirements"),
		readyLine,
		"",
		strings.Join(core, "\n"),
		"",
		sectionStyle().Render("Diagnostics"),
		"",
		strings.Join(diag, "\n"),
	}, "\n")

	return wrapText(block, wrapW)
}

func (m model) renderDetails(wrapW int, includeHardware bool) string {
	tpmVer := m.res.TPM.Version
	if tpmVer == "" && m.res.TPM.IsV2 {
		tpmVer = "2.0"
	}

	sbKeys := "present"
	if m.res.SecureBootKeys.Known {
		sbKeys = fmt.Sprintf("PK=%v KEK=%v db=%v dbx=%v",
			m.res.SecureBootKeys.PK, m.res.SecureBootKeys.KEK, m.res.SecureBootKeys.DB, m.res.SecureBootKeys.DBX)
	} else if m.res.SecureBoot.Enabled {
		sbKeys = "present (Secure Boot ON)"
	} else {
		sbKeys = "unknown"
	}

	keyW := 14
	if wrapW < 50 {
		keyW = 12
	}

	lineKV := func(k, v string) string {
		kf := keyStyle().Render(padRight(k, keyW))
		vw := max(wrapW-keyW-1, 10)
		val := wrapText(v, vw)
		val = indentWrapped(val, keyW+1)
		return kf + " " + strings.TrimPrefix(val, strings.Repeat(" ", keyW+1))
	}

	services := fmt.Sprintf("vgc=%v (%s)  vgk=%v (%s)",
		m.res.Vanguard.VGC.Running, m.res.Vanguard.VGC.Start,
		m.res.Vanguard.VGK.Running, m.res.Vanguard.VGK.Start,
	)

	main := []string{
		lineKV("TPM", fmt.Sprintf("%v/%v v%s (%s)", m.res.TPM.Present, m.res.TPM.Ready, tpmVer, m.res.TPM.Vendor)),
		lineKV("Secure Boot", fmt.Sprintf("%v (%s)", m.res.SecureBoot.Enabled, m.res.Boot.BIOSMode)),
		lineKV("SB Keys", sbKeys),
		lineKV("Disk", m.res.Disk.PartitionStyle),
		lineKV("Vanguard", fmt.Sprintf("%v  v%s", m.res.Vanguard.Installed, m.res.Vanguard.Version)),
		lineKV("Services", services),
	}

	if !includeHardware {
		return strings.Join(main, "\n")
	}

	hw := []string{
		"",
		sectionStyle().Render("Hardware"),
		"",
		lineKV("CPU", m.res.System.CPU),
		lineKV("GPU", m.res.System.GPU),
		lineKV("RAM", fmt.Sprintf("%d GiB", m.res.System.RAMGiB)),
		lineKV("Board", m.res.System.Motherboard),
		lineKV("OS", m.res.System.OS),
	}

	warns := []string{}
	if m.res.Virt.VBS_Enabled {
		warns = append(warns, "• VBS enabled: can cause Vanguard issues on some setups")
	}
	if m.res.Virt.HypervisorPresent && !m.res.Virt.HyperVEnabled {
		warns = append(warns, "• Hypervisor present: possible WSL / Device Guard / VM")
	}
	if len(warns) > 0 {
		hw = append(hw, "", warnStyle().Render("Warnings"), wrapText(strings.Join(warns, "\n"), wrapW))
	}

	return strings.Join(append(main, hw...), "\n")
}

func boxStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)
}

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
}
func subStyle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) }
func hintStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) }
func okStyle() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) }
func noStyle() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) }
func keyStyle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Bold(true) }
func sectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
}
func warnStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	return lipgloss.NewStyle().Width(width).Render(s)
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func indentWrapped(s string, indent int) string {
	prefix := strings.Repeat(" ", indent)
	lines := strings.Split(s, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func fitHeight(s string, h int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= h {
		return s
	}
	return strings.Join(lines[:h], "\n")
}
