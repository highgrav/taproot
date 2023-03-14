package acacia

type RightResponseType string

const (
	RESP_TYPE_RESPONSE RightResponseType = "response"
	RESP_TYPE_REDIRECT RightResponseType = "redirect"
	RESP_TYPE_RIGHTS   RightResponseType = "rights"
)

type RightCodeResponse struct {
	ReturnMsg  string `json:"returnMsg"`
	ReturnCode int    `json:"returnCode"`
}

type RightResponse struct {
	Type     RightResponseType `json:"responseType"`
	Response RightCodeResponse `json:"return"`
	Redirect string            `json:"redirectTo"`
	Rights   []string          `json:"rights"`
	Metadata map[string]string `json:"meta"`
}
