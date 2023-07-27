package authn

import (
	"github.com/highgrav/taproot/constants"
	"net/http"
	"time"
)

type User struct {
	RealmID                string              `json:"realmId"`
	DomainID               string              `json:"domainId"`
	UserID                 string              `json:"userId"`
	SessionID              string              `json:"sessionId"`
	SessionCreatedOn       time.Time           `json:"SessionCreatedOn"`
	Username               string              `json:"username"`
	DisplayName            string              `json:"displayName"`
	Emails                 []string            `json:"emails"`
	Phones                 []string            `json:"phones"`
	IsVerified             bool                `json:"isVerified""`
	IsBlocked              bool                `json:"isBlocked"`
	IsActive               bool                `json:"isActive"`
	IsDeleted              bool                `json:"IsDeleted"`
	RequiresPasswordUpdate bool                `json:"requiresPasswordUpdate"`
	Domains                []string            `json:"domains"`
	Workgroups             WorkgroupMembership `json:"wgs"`
	Labels                 DomainAssertions    `json:"-"` // maps Domains to labels
	Keys                   []string            `json:"keys"`
	SessionData            map[string]string   `json:"sessionData"`
}

func Anonymous() User {
	return User{
		SessionCreatedOn: time.Now(),
	}
}

func GetUserFromRequest(r *http.Request) (User, error) {
	user, ok := r.Context().Value(constants.HTTP_CONTEXT_USER_KEY).(User)
	if !ok {
		return Anonymous(), ErrUserNotAuthenticated
	}
	return user, nil
}
