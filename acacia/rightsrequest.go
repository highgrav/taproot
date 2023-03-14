package acacia

import (
	"github.com/tomasen/realip"
	"highgrav/taproot/v1/authn"
	"net/http"
	"strings"
	"time"
)

type RightsRequest struct {
	RequestDateTime time.Time        `json:"time"`
	RealmID         string           `json:"realmId"`
	DomainID        string           `json:"domainId"`
	Http            HttpRequest      `json:"http"`
	Query           QueryRequest     `json:"query"`
	User            UserRightRequest `json:"user"`
	Context         map[string]any   `json:"ctx"`
}

type HttpRequest struct {
	SourceIPAddress string              `json:"srcIp"`
	TargetHost      string              `json:"tgtHost"`
	TargetPort      string              `json:"tgtPort"`
	TargetPath      string              `json:"tgtPath"`
	Headers         map[string][]string `json:"headers"`
}

type QueryRequest struct {
	PathParams map[string]string   `json:"path"`
	Query      map[string][]string `json:"query"`
	Body       map[string]string   `json:"body"`
	Context    map[string]any      `json:"context"`
}

type UserRightRequest struct {
	UserID                 string            `json:"userId"`
	Username               string            `json:"username"`
	DisplayName            string            `json:"displayName"`
	Emails                 []string          `json:"emails"`
	Phones                 []string          `json:"phones"`
	IsVerified             bool              `json:"isVerified"`
	IsBlocked              bool              `json:"isBlocked"`
	IsActive               bool              `json:"isActive"`
	IsDeleted              bool              `json:"IsDeleted"`
	RequiresPasswordUpdate bool              `json:"requiresPasswordUpdate"`
	Workgroups             map[string]string `json:"wgs"`
	WorkgroupIds           []string          `json:"wgIds"`
	WorkgroupNames         []string          `json:"wgNames"`
	Labels                 []string          `json:"labels"`
}

func NewRightsRequest(realm string, domain string, user authn.User, r *http.Request) *RightsRequest {
	rr := &RightsRequest{
		RequestDateTime: time.Now(),
		RealmID:         realm,
		DomainID:        domain,
		Http: HttpRequest{
			SourceIPAddress: realip.FromRequest(r),
			TargetHost:      r.Host,
			TargetPort:      r.URL.Port(),
			TargetPath:      r.URL.Path,
			Headers:         make(map[string][]string),
		},
		Query: QueryRequest{
			PathParams: make(map[string]string),
			Query:      make(map[string][]string),
			Body:       nil,
			Context:    nil,
		},
		User: UserRightRequest{
			UserID:                 user.UserID,
			Username:               user.Username,
			DisplayName:            user.DisplayName,
			Emails:                 user.Emails,
			Phones:                 user.Phones,
			IsVerified:             user.IsVerified,
			IsBlocked:              user.IsBlocked,
			IsActive:               user.IsActive,
			IsDeleted:              user.IsDeleted,
			RequiresPasswordUpdate: user.RequiresPasswordUpdate,
			Workgroups:             make(map[string]string),
			WorkgroupIds:           make([]string, 0),
			WorkgroupNames:         make([]string, 0),
			Labels:                 make([]string, 0),
		},
		Context: nil,
	}

	wgs := user.Workgroups[domain]
	for k, v := range wgs {
		rr.User.Workgroups[k] = v
		rr.User.WorkgroupNames = append(rr.User.WorkgroupNames, v)
		rr.User.WorkgroupIds = append(rr.User.WorkgroupIds, k)
	}

	lbls := user.Labels[domain]
	for _, v := range lbls {
		rr.User.Labels = append(rr.User.Labels, v)
	}

	// add headers
	for k, v := range r.Header {
		rr.Http.Headers[k] = v
	}

	// add query
	q := r.URL.Query()
	for k, v := range q {
		if strings.HasPrefix(k, ":") {
			rr.Query.PathParams[k[1:]] = v[0]
		} else {
			rr.Query.Query[k] = v
		}
	}
	return rr
}
