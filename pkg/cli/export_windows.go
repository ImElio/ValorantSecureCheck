//go:build windows

package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ExportJSONToFileAndOpen(res any) (string, error) {
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}

	dir := filepath.Join(os.TempDir(), "ValorantSecureCheck")
	_ = os.MkdirAll(dir, 0755)

	name := "report_" + time.Now().Format("20060102_150405") + ".json"
	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, b, 0644); err != nil {
		return "", err
	}

	_ = exec.Command("notepad.exe", path).Start()
	return path, nil
}
