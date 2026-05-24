/*
Copyright © 2026 Two Tech Studio
*/
package api

import "runtime"

// Response is the API response for plugin searches.
type Response struct {
	Success    bool       `json:"success"`
	Data       Data       `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Data contains plugin search results.
type Data struct {
	Plugins []Plugin `json:"plugins"`
}

// Plugin represents a single plugin in the registry.
type Plugin struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	DisplayName     string   `json:"displayName"`
	Description     string   `json:"description"`
	LongDescription string   `json:"longDescription"`
	IconURL         string   `json:"iconUrl"`
	RepoURL         string   `json:"repoUrl"`
	License         string   `json:"license"`
	Tags            []string `json:"tags"`
	Keywords        []string `json:"keywords"`
	PluginType      string   `json:"pluginType"`
	Downloads       int      `json:"downloads"`
	Stars           int      `json:"stars"`
	CommentCount    int      `json:"commentCount"`
	HeatScore       int      `json:"heatScore"`
	Status          string   `json:"status"`
	TrustScore      int      `json:"trustScore"`
	LatestVersion   string   `json:"latestVersion"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`

	Author Author `json:"author"`
}

// Author represents a plugin author.
type Author struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
}

// User represents the authenticated user.
type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// Pagination contains pagination information.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// DeviceCodeResponse is returned when requesting a device authorization code.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceTokenResponse is returned when the device authorization is complete.
type DeviceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Username     string `json:"username"`
}

// RefreshResponse is returned by the /auth/refresh endpoint.
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}

// DeviceAuthError represents a protocol-level error during device authorization polling.
type DeviceAuthError struct {
	Code string
}

func (e *DeviceAuthError) Error() string {
	return e.Code
}

// BuildResponse is the API response for build queries.
type BuildResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Builds []Build `json:"builds"`
	} `json:"data"`
}

// Build represents a single plugin build.
type Build struct {
	BuildNumber      int    `json:"buildNumber"`
	CommitHash       string `json:"commitHash"`
	Status           string `json:"status"`
	ArtifactURL      string `json:"artifactUrl"`
	ArtifactURLWin   string `json:"artifactUrlWin"`
	ArtifactURLLinux string `json:"artifactUrlLinux"`
}

// ResolveArtifactURL returns the appropriate artifact URL for the current OS.
func (b *Build) ResolveArtifactURL() string {
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
