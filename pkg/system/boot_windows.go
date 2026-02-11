//go:build windows

package system

import (
	"bytes"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func GetBootInfo() (BootInfo, error) {
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`, registry.QUERY_VALUE); err == nil {
		defer k.Close()
		if v, _, e := k.GetIntegerValue("PEFirmwareType"); e == nil {
			switch v {
			case 2:
				return BootInfo{BIOSMode: "UEFI"}, nil
			case 1:
				return BootInfo{BIOSMode: "Legacy"}, nil
			}
		}
	}

	out := runPS("(Get-ComputerInfo).BiosFirmwareType")
	up := strings.ToUpper(strings.TrimSpace(out))
	if strings.Contains(up, "UEFI") {
		return BootInfo{BIOSMode: "UEFI"}, nil
	}
	if strings.Contains(up, "LEGACY") || strings.Contains(up, "BIOS") {
		return BootInfo{BIOSMode: "Legacy"}, nil
	}
	return BootInfo{BIOSMode: "Unknown"}, nil
}

func GetSecureBootKeys(sb SecureBoot) (SecureBootKeys, error) {
	var keys SecureBootKeys

	base, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\SecureBoot\Keys`, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		// Se SecureBoot è ON e non leggiamo keys: non è problema, sono presenti "per forza".
		keys.Known = false
		keys.KeysPresentForSure = sb.Enabled

		// errori tipici: access denied / privilege not held
		if err == windows.ERROR_ACCESS_DENIED || err == windows.ERROR_PRIVILEGE_NOT_HELD {
			return keys, nil
		}
		return keys, nil
	}
	defer base.Close()

	keys.Known = true

	if k, e := registry.OpenKey(base, "PK", registry.QUERY_VALUE); e == nil {
		keys.PK = true
		_ = k.Close()
	}
	if k, e := registry.OpenKey(base, "KEK", registry.QUERY_VALUE); e == nil {
		keys.KEK = true
		_ = k.Close()
	}
	if k, e := registry.OpenKey(base, "db", registry.QUERY_VALUE); e == nil {
		keys.DB = true
		_ = k.Close()
	}
	if k, e := registry.OpenKey(base, "dbx", registry.QUERY_VALUE); e == nil {
		keys.DBX = true
		_ = k.Close()
	}

	keys.KeysPresentForSure = sb.Enabled
	return keys, nil
}

func runPS(command string) string {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return strings.TrimSpace(out.String())
}
