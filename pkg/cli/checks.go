package cli

import "valorantsecurecheck/pkg/system"

func BuildChecks(tpm system.TPMInfo, sb system.SecureBoot, sys system.SystemInfo) map[string]bool {
	return map[string]bool{
		"TPM2":        tpm.Present && tpm.Ready && tpm.IsV2,
		"SecureBoot":  sb.Enabled,
		"CPU":         sys.CPU != "",
		"GPU":         sys.GPU != "",
		"RAM>=4GiB":   sys.RAMGiB >= 4,
		"Motherboard": sys.Motherboard != "",
	}
}

func CanRunValorant(checks map[string]bool) bool {
	return checks["TPM2"] &&
		checks["SecureBoot"] &&
		checks["CPU"] &&
		checks["GPU"] &&
		checks["RAM>=4GiB"] &&
		checks["Motherboard"]
}
