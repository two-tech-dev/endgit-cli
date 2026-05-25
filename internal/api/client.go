/*
Copyright © 2026 Two Tech Studio
*/
package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/two-tech-dev/endgit-cli/internal/config"
	"github.com/two-tech-dev/endgit-cli/internal/log"
)

const (
	maxRetries    = 3
	retryBaseWait = 500 * time.Millisecond
)

// authTransport adds User-Agent and (optionally) Authorization headers.
type authTransport struct {
	rt    http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "endgit-cli")
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.rt.RoundTrip(req)
}

// Client is an HTTP client for the EndGit API.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient creates a new EndGit API client.
// The client automatically attaches the stored API token (if any) to every request.
// If the stored access token is expired (or expiring within 5 minutes) and a
// refresh token is available, the tokens are silently refreshed before use.
func NewClient() *Client {
	cfg := config.GetConfig()

	base := cfg.APIURL
	if base == "" {
		base = "https://api.endgit.dev"
	}

	if cfg.APIToken != "" && cfg.RefreshToken != "" {
		if jwtExpiresWithin(cfg.APIToken, 5*time.Minute) {
			if refreshed, err := doRefreshTokens(base+"/api/v1", cfg.RefreshToken); err == nil {
				cfg.APIToken = refreshed.AccessToken
				cfg.RefreshToken = refreshed.RefreshToken
				_ = config.SaveConfig(cfg)
			} else {
				log.Debugf("Token refresh failed: %v", err)
			}
		}
	}

	return &Client{
		BaseURL: base + "/api/v1",
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &authTransport{
				rt:    http.DefaultTransport,
				token: cfg.APIToken,
			},
		},
	}
}

// jwtExpiresWithin returns true if the JWT's exp claim is within the given
// duration from now (also returns true if the token is already expired).
func jwtExpiresWithin(token string, d time.Duration) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(payload, &claims) != nil || claims.Exp == 0 {
		return false
	}
	return time.Now().Add(d).Unix() >= claims.Exp
}

// doRefreshTokens exchanges a refresh token for a new access + refresh token pair.
func doRefreshTokens(baseURL, refreshToken string) (*RefreshResponse, error) {
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	req, err := http.NewRequest("POST", baseURL+"/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "endgit-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed: HTTP %s", resp.Status)
	}

	var result struct {
		Success bool            `json:"success"`
		Data    RefreshResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("refresh failed: server returned success=false")
	}
	return &result.Data, nil
}

// RefreshTokens exchanges the given refresh token for a new token pair and
// persists the result to the config file.
func RefreshTokens(refreshToken string) (*RefreshResponse, error) {
	cfg := config.GetConfig()
	base := cfg.APIURL
	if base == "" {
		base = "https://api.endgit.dev"
	}
	result, err := doRefreshTokens(base+"/api/v1", refreshToken)
	if err != nil {
		return nil, err
	}
	cfg.APIToken = result.AccessToken
	cfg.RefreshToken = result.RefreshToken
	if saveErr := config.SaveConfig(cfg); saveErr != nil {
		log.Debugf("Failed to save refreshed tokens: %v", saveErr)
	}
	return result, nil
}

// doGetWithRetry performs a GET request with exponential backoff retry on transient errors.
func (c *Client) doGetWithRetry(url string) (*http.Response, error) {
	var lastErr error

	for attempt := range maxRetries {
		resp, err := c.HTTP.Get(url)
		if err != nil {
			lastErr = err
			log.Debugf("GET %s attempt %d failed: %v", url, attempt+1, err)
			time.Sleep(retryBaseWait * time.Duration(math.Pow(2, float64(attempt))))
			continue
		}

		// Retry on 5xx server errors
		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %s: %s", resp.Status, truncate(string(body), 200))
			log.Debugf("GET %s attempt %d: %v", url, attempt+1, lastErr)
			time.Sleep(retryBaseWait * time.Duration(math.Pow(2, float64(attempt))))
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// apiError attempts to extract a human-readable error message from an API response.
func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	// Try to parse structured API error
	var apiResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(body, &apiResp) == nil {
		if apiResp.Message != "" {
			return fmt.Errorf("API error (HTTP %s): %s", resp.Status, apiResp.Message)
		}
		if apiResp.Error != "" {
			return fmt.Errorf("API error (HTTP %s): %s", resp.Status, apiResp.Error)
		}
	}

	// Fallback to raw body
	if len(body) > 0 {
		return fmt.Errorf("API error (HTTP %s): %s", resp.Status, truncate(string(body), 200))
	}
	return fmt.Errorf("API error: HTTP %s", resp.Status)
}

// GetPlugins queries the API for plugins matching the given query string.
func (c *Client) GetPlugins(query string) (*Response, error) {
	u := fmt.Sprintf("%s/plugins", c.BaseURL)

	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := parsedURL.Query()
	q.Set("q", query)
	parsedURL.RawQuery = q.Encode()

	resp, err := c.doGetWithRetry(parsedURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var data Response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

// GetPlugin retrieves a specific plugin by name.
func (c *Client) GetPlugin(name string) (*Plugin, error) {
	u := fmt.Sprintf("%s/plugins/%s", c.BaseURL, url.PathEscape(name))

	resp, err := c.doGetWithRetry(u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("plugin %q not found in registry", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var res struct {
		Success bool   `json:"success"`
		Data    Plugin `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode plugin response: %w", err)
	}

	return &res.Data, nil
}

// GetBuilds retrieves available builds for a plugin.
func (c *Client) GetBuilds(plugin string) (*BuildResponse, error) {
	u := fmt.Sprintf("%s/builds/plugin/%s", c.BaseURL, url.PathEscape(plugin))

	resp, err := c.doGetWithRetry(u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch builds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var res BuildResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode builds response: %w", err)
	}

	return &res, nil
}

// progressWriter wraps an io.Writer and tracks download progress.
type progressWriter struct {
	writer   io.Writer
	total    int64
	written  int64
	progress func(downloaded int64, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)
	if pw.progress != nil {
		pw.progress(pw.written, pw.total)
	}
	return n, err
}

// DownloadFile downloads a file from the given URL and saves it into destDir.
// The filename is resolved from the Content-Disposition header (if present),
// falling back to the URL path basename. Returns the filename that was saved.
// The download is written to a temporary file first and renamed on success,
// so a failed download never leaves a partial file behind.
func (c *Client) DownloadFile(downloadURL string, destDir string, onProgress func(downloaded int64, total int64)) (string, error) {
	resp, err := c.HTTP.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download error: HTTP %s", resp.Status)
	}

	filename := resolveFilename(resp, downloadURL)
	destPath := path.Join(destDir, filename)

	// Write to a temp file to avoid partial/corrupt files on failure
	tmpPath := destPath + ".tmp"

	out, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	pw := &progressWriter{
		writer:   out,
		total:    resp.ContentLength,
		progress: onProgress,
	}

	buf := make([]byte, 1024*1024) // 1MB buffer for faster disk writes
	_, copyErr := io.CopyBuffer(pw, resp.Body, buf)

	// Always close before rename/cleanup
	out.Close()

	if copyErr != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write file: %w", copyErr)
	}

	// Atomic rename (move temp → final path)
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to finalize download: %w", err)
	}

	return filename, nil
}

// resolveFilename extracts the filename from a Content-Disposition header,
// falling back to the URL path basename.
func resolveFilename(resp *http.Response, downloadURL string) string {
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if name, ok := params["filename"]; ok && name != "" {
				return name
			}
		}
	}

	if u, err := url.Parse(downloadURL); err == nil {
		if base := path.Base(u.Path); base != "/" && base != "." {
			return base
		}
	}

	return "download"
}

// GetLatestReleaseAssetURL retrieves the download URL for a specific asset from the latest release.
func (c *Client) GetLatestReleaseAssetURL(repo string, assetName string) (string, error) {
	ghURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := c.HTTP.Get(ghURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitHub release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error: HTTP %s", resp.Status)
	}

	var data struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	for _, asset := range data.Assets {
		if asset.Name == assetName {
			return asset.URL, nil
		}
	}

	return "", fmt.Errorf("asset %q not found in release %s", assetName, data.TagName)
}

// GetLatestReleaseTag returns the tag_name of the latest GitHub release.
func (c *Client) GetLatestReleaseTag(repo string) (string, error) {
	ghURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := c.HTTP.Get(ghURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitHub release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error: HTTP %s", resp.Status)
	}

	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode GitHub response: %w", err)
	}
	return data.TagName, nil
}

// truncate returns the first n characters of a string, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (c *Client) doPost(rawURL string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		reqBody = &buf
	}

	req, err := http.NewRequest(http.MethodPost, rawURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.HTTP.Do(req)
}

func (c *Client) RequestDeviceCode() (*DeviceCodeResponse, error) {
	u := fmt.Sprintf("%s/auth/device", c.BaseURL)

	resp, err := c.doPost(u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var result struct {
		Success bool               `json:"success"`
		Data    DeviceCodeResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode device code response: %w", err)
	}

	return &result.Data, nil
}

func (c *Client) PollDeviceToken(deviceCode string) (*DeviceTokenResponse, error) {
	u := fmt.Sprintf("%s/auth/device/token", c.BaseURL)

	resp, err := c.doPost(u, map[string]string{"device_code": deviceCode})
	if err != nil {
		return nil, fmt.Errorf("failed to poll device token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Success bool                `json:"success"`
			Data    DeviceTokenResponse `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to decode token response: %w", err)
		}
		return &result.Data, nil
	}

	// Extract the RFC 8628 error code (authorization_pending, slow_down, etc.)
	var errResp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, &DeviceAuthError{Code: errResp.Error}
	}

	return nil, fmt.Errorf("unexpected response: HTTP %s", resp.Status)
}

// GetMe retrieves the currently authenticated user's info.
func (c *Client) GetMe() (*User, error) {
	u := fmt.Sprintf("%s/auth/me", c.BaseURL)

	resp, err := c.doGetWithRetry(u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var res struct {
		Success bool `json:"success"`
		Data    User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &res.Data, nil
}
