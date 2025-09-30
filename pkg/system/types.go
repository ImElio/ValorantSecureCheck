package system

// TPMInfo holds normalized TPM data coming from PowerShell Get-Tpm.
type TPMInfo struct {
	Present bool   `json:"present"`
	Ready   bool   `json:"ready"`
	IsV2    bool   `json:"isV2"`
	Version string `json:"version"`
	Vendor  string `json:"vendor"`
	RawJSON string `json:"rawJson"`
}

// SecureBoot indicates whether UEFI Secure Boot is enabled.
type SecureBoot struct {
	Enabled bool   `json:"enabled"`
	Source  string `json:"source"` // "registry" or "powershell"
}

// SystemInfo aggregates selected system hardware details.
type SystemInfo struct {
	CPU         string `json:"cpu"`
	GPU         string `json:"gpu"`
	RAMGiB      int    `json:"ramGiB"`
	Motherboard string `json:"motherboard"`
	OS          string `json:"os"`
}
