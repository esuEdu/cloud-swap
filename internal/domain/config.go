package domain

type Config struct {
	DefaultRegion string `json:"default_region"`
	DefaultOutput string `json:"default_output"`
	FzfEnabled    bool   `json:"fzf_enabled"`
}
