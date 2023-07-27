package acacia

import (
	"github.com/highgrav/taproot/authn"
	"github.com/tomasen/realip"
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

// A UserRightRequest is a slimmed-down version of the user struct
type UserRightRequest struct {
	UserID                 string                           `json:"userId"`
	Username               string                           `json:"username"`
	DisplayName            string                           `json:"displayName"`
	Emails                 []string                         `json:"emails"`
	Phones                 []string                         `json:"phones"`
	IsVerified             bool                             `json:"isVerified"`
	IsBlocked              bool                             `json:"isBlocked"`
	IsActive               bool                             `json:"isActive"`
	IsDeleted              bool                             `json:"IsDeleted"`
	RequiresPasswordUpdate bool                             `json:"requiresPasswordUpdate"`
	Workgroups             map[string][]authn.UserWorkgroup `json:"wgs"`
	Labels                 []string                         `json:"labels"`
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
			Workgroups:             make(map[string][]authn.UserWorkgroup),
			Labels:                 make([]string, 0),
		},
		Context: nil,
	}

	wgs := user.Workgroups[domain]
	for _, v := range wgs {
		if rr.User.Workgroups[domain] == nil {
			rr.User.Workgroups[domain] = make([]authn.UserWorkgroup, 0)
		}
		rr.User.Workgroups[domain] = append(rr.User.Workgroups[domain], authn.UserWorkgroup{
			ID:   v.ID,
			Name: v.Name,
		})
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
