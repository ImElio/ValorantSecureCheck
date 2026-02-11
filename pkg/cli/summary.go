package cli

import "fmt"

func PrintSummary(res Result) {
	if res.CanRun {
		fmt.Println("READY — all checks passed")
		return
	}
	for _, name := range []string{"TPM2", "SecureBoot", "CPU", "GPU", "RAM>=4GiB", "Motherboard", "Vanguard", "VGC"} {
		if !res.Checks[name] {
			fmt.Println("NOT READY — failing check:", humanName(name))
			return
		}
	}
	fmt.Println("NOT READY")
}

func humanName(k string) string {
	switch k {
	case "TPM2":
		return "TPM 2.0"
	case "SecureBoot":
		return "Secure Boot"
	case "RAM>=4GiB":
		return "RAM ≥ 4 GiB"
	case "Vanguard":
		return "Riot Vanguard installed"
	case "VGC":
		return "Vanguard service (vgc)"
	default:
		return k
	}
}
