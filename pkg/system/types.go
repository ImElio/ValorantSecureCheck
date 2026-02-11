package system

type TPMInfo struct {
	Present bool   `json:"present"`
	Ready   bool   `json:"ready"`
	IsV2    bool   `json:"isV2"`
	Version string `json:"version"`
	Vendor  string `json:"vendor"`
	RawJSON string `json:"rawJson"`
}

type SecureBoot struct {
	Enabled bool   `json:"enabled"`
	Source  string `json:"source"`
}

type SecureBootKeys struct {
	Known              bool `json:"known"`
	KeysPresentForSure bool `json:"keysPresentForSure"`
	PK                 bool `json:"pk"`
	KEK                bool `json:"kek"`
	DB                 bool `json:"db"`
	DBX                bool `json:"dbx"`
}

type BootInfo struct {
	BIOSMode string `json:"biosMode"` // "UEFI" / "Legacy" / "Unknown"
}

type DiskInfo struct {
	PartitionStyle string `json:"partitionStyle"` // "GPT" / "MBR" / "RAW" / "Unknown"
}

type VirtualizationInfo struct {
	HypervisorPresent bool `json:"hypervisorPresent"`
	HyperVEnabled     bool `json:"hyperVEnabled"`
	VBS_Enabled       bool `json:"vbsEnabled"`
}

type ServiceStatus struct {
	Exists  bool   `json:"exists"`
	Running bool   `json:"running"`
	Start   string `json:"start"` // Automatic/Manual/Disabled/Unknown
	Raw     string `json:"raw"`
}

type VanguardInfo struct {
	Installed     bool          `json:"installed"`
	InstallPath   string        `json:"installPath"`
	Version       string        `json:"version"`
	VGC           ServiceStatus `json:"vgc"`
	VGK           ServiceStatus `json:"vgk"`
	DriverPresent bool          `json:"driverPresent"`
}

type SystemInfo struct {
	CPU         string `json:"cpu"`
	GPU         string `json:"gpu"`
	RAMGiB      int    `json:"ramGiB"`
	Motherboard string `json:"motherboard"`
	OS          string `json:"os"`
}
