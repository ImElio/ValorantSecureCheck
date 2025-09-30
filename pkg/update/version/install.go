package version

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func InstallCLI(extractedRoot, binName, destPath string) error {
	var cliPath string
	err := filepath.WalkDir(extractedRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(p) == binName {
			cliPath = p
			return errors.New("found") // break
		}
		return nil
	})
	if err != nil && err.Error() != "found" {
		return err
	}
	if cliPath == "" {
		return errors.New("vsc.exe not found inside the archive")
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	return copyFile(cliPath, destPath)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
