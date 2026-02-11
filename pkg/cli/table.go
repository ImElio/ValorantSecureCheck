package cli

import "fmt"

func PrintTable(res Result) {
	fmt.Println("+----------------------+---------+")
	fmt.Printf("| %-20s | %-7s |\n", "CHECK", "STATUS")
	fmt.Println("+----------------------+---------+")

	printRow("TPM 2.0", res.Checks["TPM2"])
	printRow("Secure Boot", res.Checks["SecureBoot"])
	printRow("Secure Boot Keys", res.Checks["SBKeys"])
	printRow("BIOS Mode UEFI", res.Checks["BIOSUEFI"])
	printRow("Boot Disk GPT", res.Checks["DiskGPT"])
	printRow("Hyper-V Disabled", res.Checks["HyperVOff"])

	printRow("Vanguard Installed", res.Checks["Vanguard"])
	printRow("Service vgc exists", res.Checks["VGC"])
	printRow("Service vgc running", res.Checks["VGCRunning"])
	printRow("Service vgk exists", res.Checks["VGK"])

	printRow("CPU", res.Checks["CPU"])
	printRow("GPU", res.Checks["GPU"])
	printRow("RAM ≥ 4 GiB", res.Checks["RAM>=4GiB"])
	printRow("Motherboard", res.Checks["Motherboard"])

	fmt.Println("+----------------------+---------+")
}

func printRow(label string, ok bool) {
	status := "✗ Not OK"
	if ok {
		status = "✓ OK"
	}
	fmt.Printf("| %-20s | %-7s |\n", label, status)
}
