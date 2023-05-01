package authn

import (
	"github.com/highgrav/taproot/v1/constants"
	"net/http"
	"time"
)

type DomainAssertions map[string][]string

type UserWorkgroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WorkgroupMembership map[string][]UserWorkgroup

// NOTE: We don't define methods on WorkgroupMembership{} because goja doesn't like to coerce
// structs with receiver methods into JS-land.

func ByDomain(mem WorkgroupMembership, domainId string) []UserWorkgroup {
	return mem[domainId]
}

func AddDomain(mem WorkgroupMembership, domainId string) {
	_, ok := mem[domainId]
	if !ok {
		mem[domainId] = make([]UserWorkgroup, 0)
	}
}

func AddWorkgroup(mem WorkgroupMembership, domainId, workgroupId, workgroupName string) {
	_, ok := mem[domainId]
	if !ok {
		mem[domainId] = make([]UserWorkgroup, 0)
	}
	mem[domainId] = append(mem[domainId], UserWorkgroup{
		ID:   workgroupId,
		Name: workgroupName,
	})
}

func RemoveDomain(mem WorkgroupMembership, domainId string) {
	delete(mem, domainId)
}

func RemoveWorkgroupById(mem WorkgroupMembership, domainId, workgroupId string) {
	// TODO
}

func RemoveWorkgroupByName(mem WorkgroupMembership, domainId, workgroupName string) {
	// TODO
}

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
