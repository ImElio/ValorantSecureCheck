package version

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Local struct {
	CLIBinPath string
	DistDir    string
}

const DefaultCLIBinary = "vsc.exe"

func ResolveLocal(cliFlag, distFlag string) Local {
	exe, _ := os.Executable()
	base := "."
	if exe != "" {
		base = filepath.Dir(exe)
	}
	cli := cliFlag
	if cli == "" {
		cli = filepath.Join(base, DefaultCLIBinary)
	} else if !filepath.IsAbs(cli) {
		cli = filepath.Join(base, cli)
	}
	dist := distFlag
	if dist == "" {
		dist = filepath.Dir(cli)
	} else if !filepath.IsAbs(dist) {
		dist = filepath.Join(base, dist)
	}
	return Local{CLIBinPath: cli, DistDir: dist}
}

func CLIPresent(l Local) bool {
	_, err := os.Stat(l.CLIBinPath)
	return err == nil
}

func DetectCLIVersion(l Local) string {
	if l.CLIBinPath != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, l.CLIBinPath, "--version").Output()
		if err == nil {
			if v := Canonical(string(out)); v != "" {
				return v
			}
		}
	}
	if v := readText(filepath.Join(l.DistDir, "VERSION")); v != "" {
		return Canonical(v)
	}
	if b, err := os.ReadFile(filepath.Join(l.DistDir, "version.json")); err == nil {
		var m map[string]any
		if json.Unmarshal(b, &m) == nil {
			if v, ok := m["version"].(string); ok {
				return Canonical(v)
			}
		}
	}
	return ""
}

func Canonical(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimPrefix(s, "v")
}

func CompareSemver(a, b string) int {
	ap := parse(a)
	bp := parse(b)
	for i := 0; i < 3; i++ {
		if ap[i] > bp[i] {
			return 1
		}
		if ap[i] < bp[i] {
			return -1
		}
	}
	return 0
}

func parse(s string) [3]int {
	var out [3]int
	if s == "" {
		return out
	}
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	parts := strings.SplitN(s, ".", 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, ch := range parts[i] {
			if ch < '0' || ch > '9' {
				break
			}
			n = n*10 + int(ch-'0')
		}
		out[i] = n
	}
	return out
}

func readText(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}
