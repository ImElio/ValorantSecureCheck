//go:build windows

// Robust TPM detection with three stages and resilient JSON parsing.
// 1) PowerShell Get-Tpm  -> JSON (select ONLY needed fields; handle object/array/string; UTF-8; force 64-bit pwsh)
// 2) PowerShell Get-CimInstance Win32_Tpm (MicrosoftTpm namespace)
// 3) PowerShell Get-WmiObject   Win32_Tpm (classic WMI)

package system

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type psTPM struct {
	TpmPresent                bool   `json:"TpmPresent"`
	TpmReady                  bool   `json:"TpmReady"`
	SpecVersion               string `json:"SpecVersion"`
	ManufacturerIdTxt         string `json:"ManufacturerIdTxt"`
	ManufacturerVersionFull20 string `json:"ManufacturerVersionFull20"`
}

func GetTPMInfo() (TPMInfo, error) {
	// 1) Primary: Get-Tpm (64-bit PowerShell, strict JSON)
	if info, err := getTPMViaPowerShell(); err == nil {
		return info, nil
	}

	// 2) Fallback: CIM
	if info, err := getTPMViaCIM(); err == nil {
		return info, nil
	}

	// 3) Fallback: classic WMI
	if info, err := getTPMViaWMIClassic(); err == nil {
		return info, nil
	}

	return TPMInfo{}, errors.New("TPM not detected via PowerShell/CIM/WMI")
}

func getTPMViaPowerShell() (TPMInfo, error) {
	// Select ONLY required fields to avoid PS 5.1 JSON issues like {"Length":N}
	psScript1 := strings.Join([]string{
		"$ErrorActionPreference='Stop';",
		"[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;",
		"if ($PSStyle) { $PSStyle.OutputRendering='PlainText' }",
		"$t = Get-Tpm | Select-Object TpmPresent,TpmReady,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20;",
		"$t | ConvertTo-Json -Depth 4 -Compress",
	}, " ")
	raw, msg, err := runPS64(psScript1)
	if err != nil {
		return TPMInfo{RawJSON: msg}, err
	}

	// If we received a quoted string, retry by forcing array pipeline.
	if len(raw) > 0 && raw[0] == '"' {
		psScript2 := strings.Join([]string{
			"$ErrorActionPreference='Stop';",
			"[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;",
			"if ($PSStyle) { $PSStyle.OutputRendering='PlainText' }",
			"$t = Get-Tpm | Select-Object TpmPresent,TpmReady,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20;",
			"@($t) | ConvertTo-Json -Depth 4 -Compress",
		}, " ")
		raw2, msg2, err2 := runPS64(psScript2)
		if err2 != nil {
			return TPMInfo{RawJSON: pickNonEmpty(msg, msg2)}, fmt.Errorf("Get-Tpm JSON string + array retry failed: %w", err2)
		}
		var arr []psTPM
		if jerr := json.Unmarshal(raw2, &arr); jerr == nil && len(arr) > 0 {
			inf := mapPSTPM(arr[0], string(raw2))
			if !hasUsefulTPMValues(inf) {
				return TPMInfo{RawJSON: string(raw2)}, errors.New("Get-Tpm JSON missing useful values (array)")
			}
			return inf, nil
		}
		return TPMInfo{RawJSON: string(raw2)}, errors.New("Get-Tpm returned unexpected JSON (string)")
	}

	// Try parse as object, then as array.
	var one psTPM
	if jerr := json.Unmarshal(raw, &one); jerr == nil {
		inf := mapPSTPM(one, string(raw))
		if !hasUsefulTPMValues(inf) {
			return TPMInfo{RawJSON: string(raw)}, errors.New("Get-Tpm JSON missing useful values (object)")
		}
		return inf, nil
	}
	var arr []psTPM
	if jerr := json.Unmarshal(raw, &arr); jerr == nil && len(arr) > 0 {
		inf := mapPSTPM(arr[0], string(raw))
		if !hasUsefulTPMValues(inf) {
			return TPMInfo{RawJSON: string(raw)}, errors.New("Get-Tpm JSON missing useful values (array)")
		}
		return inf, nil
	}

	return TPMInfo{RawJSON: string(raw)}, errors.New("Get-Tpm returned unexpected JSON")
}

type cimTPM struct {
	IsEnabled_InitialValue    bool   `json:"IsEnabled_InitialValue"`
	IsActivated_InitialValue  bool   `json:"IsActivated_InitialValue"`
	SpecVersion               string `json:"SpecVersion"`
	ManufacturerIdTxt         string `json:"ManufacturerIdTxt"`
	ManufacturerVersionFull20 string `json:"ManufacturerVersionFull20"`
}

func getTPMViaCIM() (TPMInfo, error) {
	psScript := strings.Join([]string{
		"$ErrorActionPreference='Stop';",
		"[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;",
		"if ($PSStyle) { $PSStyle.OutputRendering='PlainText' }",
		"$t = Get-CimInstance -Namespace 'root/cimv2/Security/MicrosoftTpm' -ClassName Win32_Tpm |",
		"     Select-Object IsEnabled_InitialValue, IsActivated_InitialValue, SpecVersion, ManufacturerIdTxt, ManufacturerVersionFull20;",
		"$t | ConvertTo-Json -Depth 4 -Compress",
	}, " ")
	raw, msg, err := runPS64(psScript)
	if err != nil {
		return TPMInfo{RawJSON: msg}, err
	}

	var one cimTPM
	if jerr := json.Unmarshal(raw, &one); jerr != nil {
		var arr []cimTPM
		if jerr2 := json.Unmarshal(raw, &arr); jerr2 != nil || len(arr) == 0 {
			return TPMInfo{RawJSON: string(raw)}, jerr
		}
		one = arr[0]
	}

	isV2 := strings.Contains(one.SpecVersion, "2.0") || strings.TrimSpace(one.ManufacturerVersionFull20) != ""
	present := one.IsEnabled_InitialValue || one.IsActivated_InitialValue
	ready := one.IsEnabled_InitialValue && one.IsActivated_InitialValue

	return TPMInfo{
		Present: present,
		Ready:   ready,
		IsV2:    isV2,
		Version: one.SpecVersion,
		Vendor:  one.ManufacturerIdTxt,
		RawJSON: string(raw),
	}, nil
}

func getTPMViaWMIClassic() (TPMInfo, error) {
	psScript := strings.Join([]string{
		"$ErrorActionPreference='Stop';",
		"[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;",
		"if ($PSStyle) { $PSStyle.OutputRendering='PlainText' }",
		"$t = Get-WmiObject -Namespace 'root\\CIMV2\\Security\\MicrosoftTpm' -Class Win32_Tpm |",
		"     Select-Object IsEnabled_InitialValue, IsActivated_InitialValue, SpecVersion, ManufacturerIdTxt, ManufacturerVersionFull20;",
		"$t | ConvertTo-Json -Depth 4 -Compress",
	}, " ")
	raw, msg, err := runPS64(psScript)
	if err != nil {
		return TPMInfo{RawJSON: msg}, err
	}

	var one cimTPM
	if jerr := json.Unmarshal(raw, &one); jerr != nil {
		var arr []cimTPM
		if jerr2 := json.Unmarshal(raw, &arr); jerr2 != nil || len(arr) == 0 {
			return TPMInfo{RawJSON: string(raw)}, jerr
		}
		one = arr[0]
	}

	isV2 := strings.Contains(one.SpecVersion, "2.0") || strings.TrimSpace(one.ManufacturerVersionFull20) != ""
	present := one.IsEnabled_InitialValue || one.IsActivated_InitialValue
	ready := one.IsEnabled_InitialValue && one.IsActivated_InitialValue

	return TPMInfo{
		Present: present,
		Ready:   ready,
		IsV2:    isV2,
		Version: one.SpecVersion,
		Vendor:  one.ManufacturerIdTxt,
		RawJSON: string(raw),
	}, nil
}

// --- helpers ---

func mapPSTPM(p psTPM, raw string) TPMInfo {
	isV2 := strings.Contains(p.SpecVersion, "2.0") || strings.TrimSpace(p.ManufacturerVersionFull20) != ""
	return TPMInfo{
		Present: p.TpmPresent,
		Ready:   p.TpmReady,
		IsV2:    isV2,
		Version: p.SpecVersion,
		Vendor:  p.ManufacturerIdTxt,
		RawJSON: raw,
	}
}

// hasUsefulTPMValues = true only if any meaningful field has non-zero value.
// (Field names existing with null values are NOT considered valid.)
func hasUsefulTPMValues(inf TPMInfo) bool {
	if inf.Present || inf.Ready {
		return true
	}
	if strings.TrimSpace(inf.Version) != "" {
		return true
	}
	if strings.TrimSpace(inf.Vendor) != "" {
		return true
	}
	return false
}

// runPS64 runs Windows PowerShell (64-bit) explicitly to avoid WOW64 redirection issues.
func runPS64(script string) ([]byte, string, error) {
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		winDir = os.Getenv("SystemRoot")
	}
	pwsh := filepath.Join(winDir, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")

	cmd := exec.Command(pwsh,
		"-NoProfile", "-NoLogo", "-NonInteractive", "-ExecutionPolicy", "Bypass",
		"-Command", script)

	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return bytes.TrimSpace(out.Bytes()), strings.TrimSpace(errBuf.String()), err
}

func pickNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
