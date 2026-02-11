package cli

import "valorantsecurecheck/pkg/system"

type Result struct {
	TPM            system.TPMInfo
	SecureBoot     system.SecureBoot
	SecureBootKeys system.SecureBootKeys
	Boot           system.BootInfo
	Disk           system.DiskInfo
	Virt           system.VirtualizationInfo
	Vanguard       system.VanguardInfo
	System         system.SystemInfo
	Checks         map[string]bool
	CanRun         bool
}
