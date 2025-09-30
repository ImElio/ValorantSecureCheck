package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	ui "valorantsecurecheck/pkg/update/ui"
	ver "valorantsecurecheck/pkg/update/version"
)

var (
	flagCLI          = flag.String("cli", "", "Path to vsc.exe (optional; default: alongside updater)")
	flagDist         = flag.String("dist", "", "Working dir (optional; default: alongside CLI)")
	flagOwner        = flag.String("owner", "ImElio", "GitHub owner/org")
	flagRepo         = flag.String("repo", "ValorantSecureCheck", "GitHub repo")
	flagAssetContains= flag.String("asset", "cli", "Asset name must contain this substring")

	flagBGCheck      = flag.Bool("bg-check", false, "Background check: exit 0 if up-to-date; if update exists, open UI and start download")
	flagQuiet        = flag.Bool("quiet", false, "Suppress console output when used with --bg-check and nothing to do")
	flagAutoUpdateUI = flag.Bool("autoupdate-ui", false, "When starting the UI, immediately begin the update if available")
)

func main() {
	flag.Parse()

	if *flagBGCheck {
		runBackgroundCheck(*flagOwner, *flagRepo, *flagAssetContains, *flagQuiet)
		return
	}

	opts := ui.Options{
		Owner:      *flagOwner,
		Repo:       *flagRepo,
		AutoUpdate: *flagAutoUpdateUI,
	}
	if err := ui.Run(opts); err != nil {
		log.Fatal(err)
	}
}

func runBackgroundCheck(owner, repo, assetFilter string, quiet bool) {
	paths := ver.ResolveLocal(*flagCLI, *flagDist)
	installed := ver.CLIPresent(paths)
	local := ver.DetectCLIVersion(paths)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	rel, err := ver.LatestReleaseContext(ctx, owner, repo)
	if err != nil {
		if !quiet {
			log.Printf("update: latest release check failed: %v\n", err)
		}
		return
	}
	remote := ver.Canonical(rel.TagName)

	needsUpdate := !installed || local == "" || ver.CompareSemver(local, remote) < 0
	if !needsUpdate {
		if !quiet {
			log.Println("update: up to date")
		}
		return
	}

	self, _ := os.Executable()
	cmd := exec.Command(self,
		"-owner", owner,
		"-repo", repo,
		"-autoupdate-ui",
	)
	_ = cmd.Start()
	time.Sleep(100 * time.Millisecond)
}
