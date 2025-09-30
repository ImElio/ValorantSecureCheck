package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"valorantsecurecheck/internal/buildinfo"
	"valorantsecurecheck/pkg/cli"
	"valorantsecurecheck/pkg/system"
)

var (
	flagJSON     = flag.Bool("json", false, "Print pretty JSON and exit")
	flagTable    = flag.Bool("table", false, "Print a simple table instead of TUI")
	flagSummary  = flag.Bool("summary", false, "Print a compact READY/NOT READY summary")
	flagExitCode = flag.Bool("exit-code", false, "Exit 0 if Valorant-ready, 1 otherwise")
	flagVerbose  = flag.Bool("v", false, "Print warnings to stderr (TUI hides them)")
	flagShowVer  = flag.Bool("version", false, "Print version and exit")
)

func main() {
	flag.Parse()

	if *flagShowVer {
		fmt.Println(buildinfo.Version)
		return
	}

	spawnBackgroundUpdater()

	tpm, errTPM := system.GetTPMInfo()
	sb, errSB := system.CheckSecureBoot()
	sys, errSYS := system.GetSystemInfo()

	checks := cli.BuildChecks(tpm, sb, sys)
	res := cli.Result{
		TPM:        tpm,
		SecureBoot: sb,
		System:     sys,
		Checks:     checks,
		CanRun:     cli.CanRunValorant(checks),
	}

	if *flagVerbose {
		if errTPM != nil {
			fmt.Fprintln(os.Stderr, "[warn] TPM check error:", errTPM)
		}
		if errSB != nil {
			fmt.Fprintln(os.Stderr, "[warn] SecureBoot check error:", errSB)
		}
		if errSYS != nil {
			fmt.Fprintln(os.Stderr, "[warn] System info error:", errSYS)
		}
	}

	switch {
	case *flagJSON:
		cli.PrintJSON(res)
	case *flagTable:
		cli.PrintTable(res)
	case *flagSummary:
		cli.PrintSummary(res)
	default:
		cli.RunTUI(res)
	}

	if *flagExitCode {
		if res.CanRun {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func spawnBackgroundUpdater() {
	self, _ := os.Executable()
	base := filepath.Dir(self)
	updater := filepath.Join(base, "vsc-update.exe")
	if _, err := os.Stat(updater); err != nil {
		return 
	}

	cmd := exec.Command(updater,
		"-owner", "ImElio",
		"-repo", "ValorantSecureCheck",
		"-bg-check",
		"-quiet",
		"-asset", "cli",
	)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	_ = cmd.Start()
}
