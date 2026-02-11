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

	keys, errKeys := system.GetSecureBootKeys(sb)
	boot, errBoot := system.GetBootInfo()
	disk, errDisk := system.GetBootDiskInfo()
	virt, errVirt := system.GetVirtualizationInfo()
	vg, errVG := system.GetVanguardInfo()
	sys, errSYS := system.GetSystemInfo()

	checks := cli.BuildChecks(tpm, sb, sys, boot, disk, virt, vg, keys)
	res := cli.Result{
		TPM:            tpm,
		SecureBoot:     sb,
		SecureBootKeys: keys,
		Boot:           boot,
		Disk:           disk,
		Virt:           virt,
		Vanguard:       vg,
		System:         sys,
		Checks:         checks,
		CanRun:         cli.CanRunValorant(checks),
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
		if errKeys != nil {
			fmt.Fprintln(os.Stderr, "[warn] SecureBoot keys check error:", errKeys)
		}
		if errBoot != nil {
			fmt.Fprintln(os.Stderr, "[warn] Boot info error:", errBoot)
		}
		if errDisk != nil {
			fmt.Fprintln(os.Stderr, "[warn] Disk info error:", errDisk)
		}
		if errVirt != nil {
			fmt.Fprintln(os.Stderr, "[warn] Virtualization info error:", errVirt)
		}
		if errVG != nil {
			fmt.Fprintln(os.Stderr, "[warn] Vanguard info error:", errVG)
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
		system.TrySetConsoleSize(150, 42)
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
