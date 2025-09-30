package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

func LatestRelease(owner, repo string) (Release, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ValorantSecureCheck-Updater")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Release{}, fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Release{}, err
	}
	return r, nil
}

func FindCLIAsset(r GitHubRelease) (Asset, bool) {
	for _, a := range r.Assets {
		name := strings.ToLower(a.Name)
		if strings.HasSuffix(name, ".zip") &&
			strings.Contains(name, "cli") &&
			strings.Contains(name, "windows") {
			return a, true
		}
	}
	return Asset{}, false
}

func LatestReleaseContext(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	rel, err := latestViaAPI(ctx, owner, repo)
	if err == nil {
		return rel, nil
	}
	return latestViaHTML(ctx, owner, repo)
}

func latestViaAPI(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ValorantSecureCheck-Updater")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api status %d", resp.StatusCode)
	}

	var r GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func latestViaHTML(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	u := fmt.Sprintf("https://github.com/%s/%s/releases/latest", owner, repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "ValorantSecureCheck-Updater")

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		return nil, fmt.Errorf("no redirect location from %s", u)
	}

	var tag string
	if idx := strings.LastIndex(loc, "/tag/"); idx != -1 {
		tag = loc[idx+len("/tag/"):]
		if i := strings.IndexAny(tag, "?#"); i != -1 {
			tag = tag[:i]
		}
	} else {
		trimmed := strings.TrimRight(loc, "/")
		parts := strings.Split(trimmed, "/")
		if len(parts) > 0 {
			tag = parts[len(parts)-1]
			if i := strings.IndexAny(tag, "?#"); i != -1 {
				tag = tag[:i]
			}
		}
	}

	if tag == "" {
		return nil, fmt.Errorf("cannot parse tag from Location: %s", loc)
	}

	return fetchReleaseByTag(ctx, owner, repo, tag)
}

func fetchReleaseByTag(ctx context.Context, owner, repo, tag string) (*GitHubRelease, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ValorantSecureCheck-Updater")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api status %d for tag %s", resp.StatusCode, tag)
	}

	var r GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}