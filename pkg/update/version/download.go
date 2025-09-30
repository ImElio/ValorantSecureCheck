package version

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DownloadAsset(ctx context.Context, url, dir, filename string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, filename)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "ValorantSecureCheck-Updater")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download status %d", resp.StatusCode)
	}
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return dst, nil
}

func Unzip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	base := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
	dest := filepath.Join(os.TempDir(), base+"-unzipped-"+fmt.Sprint(time.Now().UnixNano()))
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}

	for _, f := range r.File {
		fp := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fp, dest) {
			return "", fmt.Errorf("zip entry outside dest: %s", f.Name)
		}
		if f.FileHeader.FileInfo().IsDir() {
			if err := os.MkdirAll(fp, 0o755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out, err := os.Create(fp)
		if err != nil {
			rc.Close()
			return "", err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return "", err
		}
		out.Close()
		rc.Close()
	}
	return dest, nil
}
