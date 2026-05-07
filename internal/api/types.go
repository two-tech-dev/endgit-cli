package api

type Response struct {
	Success    bool       `json:"success"`
	Data       Data       `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Data struct {
	Plugins []Plugin `json:"plugins"`
}

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
	Status          string   `json:"status"`
	TrustScore      int      `json:"trustScore"`
	LatestVersion   string   `json:"latestVersion"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`

	Author Author `json:"author"`
}

type Author struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
