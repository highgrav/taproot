package acacia

type Right struct {
	Resource string            `json:"rscId"`
	Entity   string            `json:"entId"`
	Action   string            `json:"actId"`
	Metadata map[string]string `json:"meta"`
}
