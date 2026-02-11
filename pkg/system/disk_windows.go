//go:build windows

package system

import (
	"bytes"
	"os/exec"
	"strings"
)

func GetBootDiskInfo() (DiskInfo, error) {
	script := strings.Join([]string{
		"$ErrorActionPreference='Stop';",
		"$d = Get-Disk | Where-Object IsBoot -eq $true | Select-Object -First 1 -ExpandProperty PartitionStyle;",
		"if ($null -eq $d) { 'Unknown' } else { $d }",
	}, " ")

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	val := strings.TrimSpace(out.String())
	if val == "" {
		val = "Unknown"
	}

	up := strings.ToUpper(val)
	switch up {
	case "GPT", "MBR", "RAW":
		return DiskInfo{PartitionStyle: up}, err
	default:
		return DiskInfo{PartitionStyle: val}, err
	}
}
