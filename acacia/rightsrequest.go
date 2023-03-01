package acacia

type RightsRequest struct {
	HttpRequest      HttpRequest      `json:"http"`
	ContextRequest   ContextRequest   `json:"ctx"`
	UserRightRequest UserRightRequest `json:"user""`
}

type HttpRequest struct {
	SourceIPAddress string `json:"srcIp"`
	TargetHost      string `json:"tgtHost"`
	TargetPort      string `json:"tgtPort""`
	TargetPath      string `json:"tgtPath"`
}

type ContextRequest struct {
	PathParams map[string]string `json:"pathParams"`
	Query      map[string]string `json:"query"`
	Body       map[string]string `json:"body"`
	Context    map[string]any    `json:"context"`
}

type UserRightRequest struct {
	UserID                 string            `json:"userId"`
	Username               string            `json:"username"`
	DisplayName            string            `json:"displayName"`
	Emails                 []string          `json:"emails"`
	Phones                 []string          `json:"phones"`
	IsVerified             bool              `json:"isVerified""`
	IsBlocked              bool              `json:"isBlocked"`
	IsActive               bool              `json:"isActive"`
	IsDeleted              bool              `json:"IsDeleted"`
	RequiresPasswordUpdate bool              `json:"requiresPasswordUpdate"`
	Workgroups             map[string]string `json:"wgs"`
	Labels                 []string          `json:"labels"`
}
