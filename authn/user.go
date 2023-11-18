package authn

import (
	"github.com/highgrav/taproot/constants"
	"net/http"
	"time"
)

type User struct {
	EnvironmentID          string              `json:"envId,omitempty"`
	RealmID                string              `json:"realmId,omitempty"`
	DomainID               string              `json:"domainId,omitempty"`
	SessionID              string              `json:"sessionId,omitempty"`
	UserID                 string              `json:"userId"`
	SessionCreatedOn       time.Time           `json:"SessionCreatedOn"`
	Username               string              `json:"username"`
	DisplayName            string              `json:"displayName,omitempty"`
	Emails                 []string            `json:"emails,omitempty"`
	Phones                 []string            `json:"phones,omitempty"`
	IsVerified             bool                `json:"isVerified,omitempty""`
	IsBlocked              bool                `json:"isBlocked,omitempty"`
	IsActive               bool                `json:"isActive,omitempty"`
	IsDeleted              bool                `json:"IsDeleted,omitempty"`
	RequiresPasswordUpdate bool                `json:"requiresPasswordUpdate,omitempty"`
	Domains                []string            `json:"domains,omitempty"`
	Workgroups             WorkgroupMembership `json:"wgs,omitempty"`
	Labels                 DomainAssertions    `json:"-"` // maps Domains to labels
	Keys                   []string            `json:"keys,omitempty"`
	SessionData            map[string]string   `json:"sessionData,omitempty"`
	AvatarID               string              `json:"avatarId,omitempty"`
	PreferredLocale        string              `json:"preferredLocale,omitempty"`
	AdditionalData         map[string]any      `json:"additionalData,omitempty"`
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
