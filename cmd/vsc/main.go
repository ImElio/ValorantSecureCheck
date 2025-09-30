package main

// Thin orchestrator: parse flags, gather data, delegate rendering to pkg/cli.

import (
	"flag"
	"fmt"
	"os"

	"valorantsecurecheck/pkg/cli"
	"valorantsecurecheck/pkg/system"
)

var (
	flagJSON     = flag.Bool("json", false, "Print pretty JSON and exit")
	flagTable    = flag.Bool("table", false, "Print a simple table instead of TUI")
	flagSummary  = flag.Bool("summary", false, "Print a compact READY/NOT READY summary")
	flagExitCode = flag.Bool("exit-code", false, "Exit 0 if Valorant-ready, 1 otherwise")
	flagVerbose  = flag.Bool("v", false, "Print warnings to stderr (TUI hides them)")
)

func main() {
	flag.Parse()

	tpm, errTPM := system.GetTPMInfo()
	sb, errSB := system.CheckSecureBoot()
	sys, errSYS := system.GetSystemInfo()

	checks := cli.BuildChecks(tpm, sb, sys)
	canRun := cli.CanRunValorant(checks)
	res := cli.Result{
		TPM:        tpm,
		SecureBoot: sb,
		System:     sys,
		Checks:     checks,
		CanRun:     canRun,
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

	// 4) Output mode
	switch {
	case *flagJSON:
		cli.PrintJSON(res)
	case *flagTable:
		cli.PrintTable(res)
	case *flagSummary:
		cli.PrintSummary(res)
	default:
		cli.RunTUI(res) // default: interactive view. J to toggle JSON, Q to quit
	}

	if *flagExitCode {
		if res.CanRun {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
