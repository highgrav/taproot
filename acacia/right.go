package acacia

type RightResponse struct {
	ReturnMsg  string            `json:"returnMsg"`
	ReturnCode int               `json:"returnCode"`
	Redirect   string            `json:"redirectTo"`
	Rights     []string          `json:"rights"`
	Metadata   map[string]string `json:"meta"`
}
