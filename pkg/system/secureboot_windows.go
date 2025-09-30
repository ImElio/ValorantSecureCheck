//go:build windows

package system

import (
	"bytes"
	"os/exec"

	"golang.org/x/sys/windows/registry"
)

func CheckSecureBoot() (SecureBoot, error) {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`System\CurrentControlSet\Control\SecureBoot\State`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		defer k.Close()
		val, _, gerr := k.GetIntegerValue("UEFISecureBootEnabled")
		if gerr == nil {
			return SecureBoot{Enabled: val == 1, Source: "registry"}, nil
		}
	}

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", "$v = Confirm-SecureBootUEFI; if ($?) { if ($v) { 'True' } else { 'False' } }")
	var out bytes.Buffer
	cmd.Stdout = &out
	if perr := cmd.Run(); perr == nil {
		enabled := bytes.Contains(bytes.ToLower(bytes.TrimSpace(out.Bytes())), []byte("true"))
		return SecureBoot{Enabled: enabled, Source: "powershell"}, nil
	}

	return SecureBoot{Enabled: false, Source: "unknown"}, err
}
