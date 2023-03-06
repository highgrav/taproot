package acacia

type RightResponse struct {
	Deny     bool              `json:"deny"`
	Redirect string            `json:"redirectTo"`
	Rights   []string          `json:"rights"`
	Metadata map[string]string `json:"meta"`
}
