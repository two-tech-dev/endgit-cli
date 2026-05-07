/*
Copyright © 2026 Two Tech Studio
*/
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/two-tech-dev/endgit-cli/internal/config"
)

type userAgentTransport struct {
	rt http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "endgit-cli")
	return t.rt.RoundTrip(req)
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient() *Client {
	cfg := config.GetConfig()

	base := cfg.APIURL
	if base == "" {
		base = "https://api.endgit.dev"
	}

	return &Client{
		BaseURL: base + "/api/v1",
		HTTP: &http.Client{
			Timeout: 10 * time.Minute,
			Transport: &userAgentTransport{
				rt: http.DefaultTransport,
			},
		},
	}
}

func (c *Client) GetPlugins(query string) (*Response, error) {
	u := fmt.Sprintf("%s/plugins", c.BaseURL)

	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	q := parsedURL.Query()
	q.Set("q", query)
	parsedURL.RawQuery = q.Encode()

	resp, err := c.HTTP.Get(parsedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}

	var data Response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) GetPlugin(name string) (*Plugin, error) {
	u := fmt.Sprintf("%s/plugins/%s", c.BaseURL, name)

	resp, err := c.HTTP.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}

	var res struct {
		Success bool   `json:"success"`
		Data    Plugin `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res.Data, nil
}

func (c *Client) GetBuilds(plugin string) (*BuildResponse, error) {
	u := fmt.Sprintf("%s/builds/plugin/%s", c.BaseURL, plugin)

	resp, err := c.HTTP.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: %s", resp.Status)
	}

	var res BuildResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

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

func (c *Client) DownloadFile(url string, destPath string, onProgress func(downloaded int64, total int64)) error {
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &progressWriter{
		writer:   out,
		total:    resp.ContentLength,
		progress: onProgress,
	}

	buf := make([]byte, 1024*1024) // 1MB buffer for faster disk writes
	_, err = io.CopyBuffer(pw, resp.Body, buf)
	return err
}
