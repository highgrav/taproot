package jsrun

type JSCallReturnValue struct {
	OK                bool                   `json:"ok"`
	Errors            []string               `json:"errors"`
	ResultCode        int32                  `json:"resultCode"`
	ResultDescription string                 `json:"resultDesc"`
	Results           map[string]interface{} `json:"results"`
}
