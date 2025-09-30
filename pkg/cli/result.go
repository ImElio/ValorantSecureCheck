package cli

import "valorantsecurecheck/pkg/system"

type Result struct {
	TPM 		system.TPMInfo
	SecureBoot	system.SecureBoot
	System		system.SystemInfo
	Checks		map[string]bool
	CanRun		bool
}