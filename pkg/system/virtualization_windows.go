//go:build windows

package system

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

type psVirt struct {
	HypervisorPresent bool `json:"HypervisorPresent"`
}

type psDeviceGuard struct {
	VirtualizationBasedSecurityStatus int `json:"VirtualizationBasedSecurityStatus"`
}

func GetVirtualizationInfo() (VirtualizationInfo, error) {
	var vi VirtualizationInfo
	var err error

	// Hypervisor present
	{
		script := strings.Join([]string{
			"$ErrorActionPreference='SilentlyContinue';",
			"$c = Get-CimInstance Win32_ComputerSystem | Select-Object HypervisorPresent;",
			"$c | ConvertTo-Json -Compress",
		}, " ")
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
		var out bytes.Buffer
		cmd.Stdout = &out
		if e := cmd.Run(); e == nil {
			var v psVirt
			if j := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &v); j == nil {
				vi.HypervisorPresent = v.HypervisorPresent
			}
		} else {
			err = e
		}
	}

	// Hyper-V feature state (best-effort)
	{
		state := runPSOneLine("(Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All).State")
		vi.HyperVEnabled = strings.Contains(strings.ToLower(state), "enabled")
	}

	// VBS (DeviceGuard)
	{
		script := strings.Join([]string{
			"$ErrorActionPreference='SilentlyContinue';",
			"$d = Get-CimInstance -Namespace root\\Microsoft\\Windows\\DeviceGuard -ClassName Win32_DeviceGuard |",
			"     Select-Object -First 1 VirtualizationBasedSecurityStatus;",
			"$d | ConvertTo-Json -Compress",
		}, " ")
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
		var out bytes.Buffer
		cmd.Stdout = &out
		if e := cmd.Run(); e == nil {
			var dg psDeviceGuard
			if j := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &dg); j == nil {
				vi.VBS_Enabled = dg.VirtualizationBasedSecurityStatus != 0
			}
		}
	}

	return vi, err
}

func runPSOneLine(command string) string {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", command)
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}
