//go:build windows

package system

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func GetVanguardInfo() (VanguardInfo, error) {
	var vi VanguardInfo

	vi.VGC = getServiceStatus("vgc")
	vi.VGK = getServiceStatus("vgk")

	vi.InstallPath = detectVanguardInstallPath()
	if vi.InstallPath != "" {
		vi.Installed = true
	}
	if !vi.Installed {
		vi.Installed = vi.VGC.Exists || vi.VGK.Exists
	}

	vi.Version = detectVanguardVersion(vi.InstallPath)

	if vi.InstallPath != "" {
		vgk := filepath.Join(vi.InstallPath, "vgk.sys")
		if _, err := os.Stat(vgk); err == nil {
			vi.DriverPresent = true
		}
	}

	return vi, nil
}

func detectVanguardInstallPath() string {
	common := `C:\Program Files\Riot Vanguard`
	if _, err := os.Stat(common); err == nil {
		return common
	}
	if p := serviceImageDir("vgc"); p != "" {
		return p
	}
	if p := serviceImageDir("vgk"); p != "" {
		return p
	}
	return ""
}

func serviceImageDir(service string) string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\`+service, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()

	ip, _, err := k.GetStringValue("ImagePath")
	if err != nil {
		return ""
	}

	ip = strings.TrimSpace(ip)
	ip = strings.Trim(ip, "\"")
	ip = strings.TrimPrefix(ip, "\\??\\")
	if i := strings.Index(ip, " "); i > 0 {
		ip = ip[:i]
	}

	dir := filepath.Dir(ip)
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return ""
}

func detectVanguardVersion(installPath string) string {
	if installPath == "" {
		return ""
	}
	vgc := filepath.Join(installPath, "vgc.exe")
	if _, err := os.Stat(vgc); err != nil {
		return ""
	}

	cmd := exec.Command(
		"powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", "(Get-Item '"+escapePSPath(vgc)+"').VersionInfo.FileVersion",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return strings.TrimSpace(out.String())
}

func escapePSPath(p string) string { return strings.ReplaceAll(p, "'", "''") }

func getServiceStatus(name string) ServiceStatus {
	var ss ServiceStatus

	// exists/running via sc query
	q := run("sc", "query", name)
	ss.Raw = q
	low := strings.ToLower(q)
	if strings.Contains(low, "does not exist") || strings.Contains(low, "non esiste") || strings.Contains(low, "failed") {
		ss.Exists = false
		ss.Start = "Unknown"
		return ss
	}
	ss.Exists = true
	ss.Running = strings.Contains(strings.ToUpper(q), "RUNNING") || strings.Contains(strings.ToUpper(q), "IN ESECUZIONE")

	if start, ok := getStartTypeFromPowerShell(name); ok && start != "Unknown" {
		ss.Start = start
		return ss
	}

	if st := getStartFromRegistry(name); st != "Unknown" {
		ss.Start = st
		return ss
	}

	if st := getStartFromSCQC(name); st != "Unknown" {
		ss.Start = st
		return ss
	}

	ss.Start = "Unknown"
	return ss
}

func getStartTypeFromPowerShell(name string) (string, bool) {
	cmd := exec.Command(
		"powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", "try { (Get-Service -Name '"+name+"' | Select-Object -ExpandProperty StartType) } catch { '' }",
	)
	out, _ := cmd.CombinedOutput()
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "Unknown", false
	}
	return normalizeStartType(s), true
}

func getStartFromRegistry(service string) string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\`+service, registry.QUERY_VALUE)
	if err != nil {
		return "Unknown"
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue("Start")
	if err != nil {
		return "Unknown"
	}

	// For drivers/services:
	// 0 Boot, 1 System, 2 Automatic, 3 Manual, 4 Disabled
	switch v {
	case 0:
		return "Boot"
	case 1:
		return "System"
	case 2:
		return "Automatic"
	case 3:
		return "Manual"
	case 4:
		return "Disabled"
	default:
		return "Unknown"
	}
}

func getStartFromSCQC(service string) string {
	qc := run("sc", "qc", service)

	re := regexp.MustCompile(`(?i)START_TYPE\s*:\s*(\d)`)
	m := re.FindStringSubmatch(qc)
	if len(m) == 2 {
		switch m[1] {
		case "0":
			return "Boot"
		case "1":
			return "System"
		case "2":
			return "Automatic"
		case "3":
			return "Manual"
		case "4":
			return "Disabled"
		}
	}

	up := strings.ToUpper(qc)
	switch {
	case strings.Contains(up, "BOOT_START"):
		return "Boot"
	case strings.Contains(up, "SYSTEM_START"):
		return "System"
	case strings.Contains(up, "AUTO_START") || strings.Contains(up, "AUTOMATIC"):
		return "Automatic"
	case strings.Contains(up, "DEMAND_START") || strings.Contains(up, "MANUAL"):
		return "Manual"
	case strings.Contains(up, "DISABLED"):
		return "Disabled"
	default:
		return "Unknown"
	}
}

func normalizeStartType(s string) string {
	up := strings.ToUpper(strings.TrimSpace(s))

	// PowerShell can return: Automatic, Manual, Disabled
	// Some drivers may map to: System / Boot (or weird strings). Handle them.
	switch {
	case strings.Contains(up, "BOOT"):
		return "Boot"
	case strings.Contains(up, "SYSTEM"):
		return "System"
	case strings.Contains(up, "AUTO") || strings.Contains(up, "AUTOMATIC"):
		return "Automatic"
	case strings.Contains(up, "MANUAL") || strings.Contains(up, "DEMAND"):
		return "Manual"
	case strings.Contains(up, "DISABLED"):
		return "Disabled"
	default:
		return "Unknown"
	}
}

func run(cmd string, args ...string) string {
	out, _ := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out))
}
