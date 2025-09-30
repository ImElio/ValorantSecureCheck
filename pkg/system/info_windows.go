//go:build windows

package system

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/wmi"
)

type win32_Processor struct{ Name string }
type win32_VideoController struct{ Name string }
type win32_ComputerSystem struct{ TotalPhysicalMemory string }
type win32_BaseBoard struct{ Manufacturer, Product string }
type win32_OperatingSystem struct{ Caption string }

func GetSystemInfo() (SystemInfo, error) {
	var sys SystemInfo
	var err error

	var cpus []win32_Processor
	if e := wmi.Query("SELECT Name FROM Win32_Processor", &cpus); e == nil && len(cpus) > 0 {
		sys.CPU = cleanWS(cpus[0].Name)
	} else if e != nil {
		err = wrapErr(err, e)
	}

	var gpus []win32_VideoController
	if e := wmi.Query("SELECT Name FROM Win32_VideoController", &gpus); e == nil && len(gpus) > 0 {
		for _, g := range gpus {
			if !strings.Contains(strings.ToLower(g.Name), "microsoft basic display") {
				sys.GPU = cleanWS(g.Name); break
			}
		}
		if sys.GPU == "" { sys.GPU = cleanWS(gpus[0].Name) }
	} else if e != nil {
		err = wrapErr(err, e)
	}

	var cs []win32_ComputerSystem
	if e := wmi.Query("SELECT TotalPhysicalMemory FROM Win32_ComputerSystem", &cs); e == nil && len(cs) > 0 {
		if bytes, perr := strconv.ParseUint(cs[0].TotalPhysicalMemory, 10, 64); perr == nil {
			sys.RAMGiB = int(bytes / (1024 * 1024 * 1024))
		} else { err = wrapErr(err, perr) }
	} else if e != nil {
		err = wrapErr(err, e)
	}

	var bb []win32_BaseBoard
	if e := wmi.Query("SELECT Manufacturer, Product FROM Win32_BaseBoard", &bb); e == nil && len(bb) > 0 {
		sys.Motherboard = fmt.Sprintf("%s %s", cleanWS(bb[0].Manufacturer), cleanWS(bb[0].Product))
	} else if e != nil {
		err = wrapErr(err, e)
	}

	var osItems []win32_OperatingSystem
	if e := wmi.Query("SELECT Caption FROM Win32_OperatingSystem", &osItems); e == nil && len(osItems) > 0 {
		sys.OS = cleanWS(osItems[0].Caption)
	} else if e != nil {
		err = wrapErr(err, e)
	}

	return sys, err
}

func cleanWS(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\n", " ")) }
func wrapErr(base, newerr error) error {
	if base == nil { return newerr }
	return fmt.Errorf("%v; %w", base, newerr)
}
