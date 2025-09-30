package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ver "valorantsecurecheck/pkg/update/version"
)

type Options struct {
	Owner      string
	Repo       string
	AutoUpdate bool
}

const (
	cliBinaryName = ver.DefaultCLIBinary // "vsc.exe"

	cooldown       = 60 * time.Second
	requestTimeout = 15 * time.Second

	statusBoxMinWidth = 52
)

type status int

const (
	statUnknown status = iota
	statNotInstalled
	statUpToDate
	statUpdateAvailable
)

type model struct {
	owner, repo string
	autoUpdate  bool

	paths        ver.Local
	localVersion string
	latest       string
	stat         status

	notice      string
	nextAllowed time.Time
	busy        bool
}

func Run(opts Options) error {
	if opts.Owner == "" {
		opts.Owner = "ImElio"
	}
	if opts.Repo == "" {
		opts.Repo = "ValorantSecureCheck"
	}
	m := newModel(opts)
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func newModel(opts Options) model {
	paths := ver.ResolveLocal("", "")
	return model{
		owner:        opts.Owner,
		repo:         opts.Repo,
		autoUpdate:   opts.AutoUpdate,
		paths:        paths,
		localVersion: "-",
		latest:       "-",
		stat:         statUnknown,
		notice:       "",
		nextAllowed:  time.Now(),
	}
}


type (
	refreshMsg struct {
		local       string
		latest      string
		st          status
		notice      string
		nextAllowed time.Time
		err         error
	}
	doneMsg struct {
		notice string
		err    error
	}
)

func (m model) Init() tea.Cmd {
	// first refresh always ignores cooldown
	return m.refreshCmd(true)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshMsg:
		if !msg.nextAllowed.IsZero() {
			m.nextAllowed = msg.nextAllowed
		}
		if msg.err != nil {
			m.notice = shortErr(msg.err)
			return m, nil
		}
		m.localVersion = msg.local
		m.latest = msg.latest
		m.stat = msg.st
		if m.autoUpdate && m.stat == statUpdateAvailable && !m.busy {
			m.busy = true
			m.notice = "Updating…"
			return m, m.updateCmd()
		}
		return m, nil

	case doneMsg:
		m.busy = false
		if msg.err != nil {
			m.notice = shortErr(msg.err)
			return m, nil
		}
		m.notice = msg.notice
		// after update, force a fresh status (ignores cooldown)
		return m, m.refreshCmd(true)

	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "r":
			return m, m.refreshCmd(false)

		case "o":
			_ = exec.Command("rundll32", "url.dll,FileProtocolHandler",
				"https://github.com/"+m.owner+"/"+m.repo+"/releases/latest").Start()
			m.notice = "Opened releases in browser"
			return m, nil

		case "u":
			if m.stat == statUpdateAvailable && !m.busy {
				m.busy = true
				m.notice = "Updating…"
				return m, m.updateCmd()
			}
			return m, m.refreshCmd(true)
		}
	}
	return m, nil
}

func (m model) View() string {
	title := styleTitle().Render("ValorantSecureCheck Updater")
	menu := styleMenu().Render("[U] Update   [R] Refresh   [O] Open releases   [Q] Quit")

	statText := "-"
	switch m.stat {
	case statNotInstalled:
		statText = styleWarn().Render("NOT INSTALLED")
	case statUpToDate:
		statText = styleOK().Render("UP TO DATE")
	case statUpdateAvailable:
		statText = styleInfo().Render("UPDATE AVAILABLE")
	}

	body := fmt.Sprintf("%s\n%s\n%s\n%s\n",
		row("Status", statText),
		row("Local", m.localVersion),
		row("Latest", m.latest),
		row("Binary", cliBinaryName),
	)
	box := styleBox().Width(maxInt(statusBoxMinWidth, lenW(body)+6)).Render(body)

	notice := ""
	if m.notice != "" {
		notice = "\n" + styleNotice().Render(m.notice)
	}

	footer := styleFaint().Render("Desktop App: coming soon")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		menu,
		"",
		box,
		notice,
		"\n"+footer,
	)
}


func (m model) refreshCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		if !force && now.Before(m.nextAllowed) {
			remain := time.Until(m.nextAllowed).Round(time.Second)
			return refreshMsg{
				local:       m.localVersion,
				latest:      m.latest,
				st:          m.stat,
				notice:      fmt.Sprintf("Next refresh allowed in %s", remain),
				nextAllowed: m.nextAllowed,
			}
		}

		paths := ver.ResolveLocal("", "")
		local := "-"
		installed := ver.CLIPresent(paths)
		if installed {
			if v := ver.DetectCLIVersion(paths); v != "" {
				local = v
			} else {
				local = "unknown"
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		rel, err := ver.LatestReleaseContext(ctx, m.owner, m.repo)
		if err != nil {
			return refreshMsg{
				err:         err,
				nextAllowed: time.Now().Add(cooldown),
			}
		}
		remote := ver.Canonical(rel.TagName)

		st := statNotInstalled
		if installed {
			if ver.CompareSemver(local, remote) < 0 {
				st = statUpdateAvailable
			} else {
				st = statUpToDate
			}
		}

		return refreshMsg{
			local:       local,
			latest:      remote,
			st:          st,
			notice:      fmt.Sprintf("Next refresh allowed in %s", cooldown.Truncate(time.Second)),
			nextAllowed: time.Now().Add(cooldown),
		}
	}
}

func (m model) updateCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		rel, err := ver.LatestReleaseContext(ctx, m.owner, m.repo)
		if err != nil {
			return doneMsg{err: err}
		}
		asset, ok := ver.FindCLIAsset(*rel)
		if !ok {
			return doneMsg{err: errors.New(`asset not found (expected name containing "cli" and "windows")`)}
		}

		tmpZip, err := ver.DownloadAsset(ctx, asset.BrowserDownloadURL, os.TempDir(), asset.Name)
		if err != nil {
			return doneMsg{err: err}
		}
		root, err := ver.Unzip(tmpZip)
		if err != nil {
			return doneMsg{err: err}
		}
		if err := ver.InstallCLI(root, cliBinaryName, ver.ResolveLocal("", "").CLIBinPath); err != nil {
			return doneMsg{err: err}
		}
		return doneMsg{notice: "Update completed"}
	}
}


func styleTitle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) }
func styleMenu() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) }
func styleBox() lipgloss.Style    { return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2) }
func styleOK() lipgloss.Style     { return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) }
func styleWarn() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) }
func styleInfo() lipgloss.Style   { return lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true) }
func styleFaint() lipgloss.Style  { return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) }
func styleNotice() lipgloss.Style { return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1) }

func row(k, v string) string { return fmt.Sprintf("%-8s %s", k, v) }

func shortErr(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "403") {
		return "GitHub rate limit: try again later"
	}
	return msg
}
func maxInt(a, b int) int { if a > b { return a } ; return b }

// approx width counting runes except newlines
func lenW(s string) int {
	n := 0
	for _, r := range s {
		if r != '\n' {
			n++
		}
	}
	return n
}
