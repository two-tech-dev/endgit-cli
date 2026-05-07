/*
Copyright © 2026 Two Tech Studio
*/
package common

import (
	"io"
	"net/http"
	"runtime"

	"github.com/two-tech-dev/endgit-cli/internal/api"
)

func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func IsHex(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func ShortHash(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}

func ResolveArtifactURL(b *api.Build) string {
	switch runtime.GOOS {
	case "windows":
		if b.ArtifactURLWin != "" {
			return b.ArtifactURLWin
		}
	case "linux":
		if b.ArtifactURLLinux != "" {
			return b.ArtifactURLLinux
		}
	}

	return b.ArtifactURL
}
