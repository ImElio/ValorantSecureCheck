package cli

import (
	"strings"

	"valorantsecurecheck/pkg/system"
)

func BuildChecks(
	tpm system.TPMInfo,
	sb system.SecureBoot,
	sys system.SystemInfo,
	boot system.BootInfo,
	disk system.DiskInfo,
	virt system.VirtualizationInfo,
	vg system.VanguardInfo,
	keys system.SecureBootKeys,
) map[string]bool {

	sbKeysOK := false
	if sb.Enabled {
		// Secure Boot ON
		sbKeysOK = true
	} else if keys.Known {
		sbKeysOK = keys.PK && keys.KEK && keys.DB
	}

	return map[string]bool{
		"TPM2":       tpm.Present && tpm.Ready && tpm.IsV2,
		"SecureBoot": sb.Enabled,
		"UEFI":       strings.EqualFold(boot.BIOSMode, "UEFI"),
		"GPT":        strings.EqualFold(disk.PartitionStyle, "GPT"),
		"Vanguard":   vg.Installed,
		"VGCExists":  vg.VGC.Exists,

		"SBKeys":      sbKeysOK,
		"VGCRunning":  vg.VGC.Running,
		"VGKExists":   vg.VGK.Exists,
		"HyperVOff":   !virt.HyperVEnabled,
		"VBSDisabled": !virt.VBS_Enabled,

		"CPU":       sys.CPU != "",
		"RAM>=4GiB": sys.RAMGiB >= 4,
	}
}

func CanRunValorant(checks map[string]bool) bool {
	return checks["TPM2"] &&
		checks["SecureBoot"] &&
		checks["UEFI"] &&
		checks["GPT"] &&
		checks["Vanguard"] &&
		checks["VGCExists"]
}
