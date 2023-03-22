package authn

type DomainAssertions map[string][]string

type WorkgroupMembership map[string]map[string]string

// (Defining a pointer receiver on this seems non-idiomatic)
func (mem WorkgroupMembership) ByDomain(domainId string) map[string]string {
	return mem[domainId]
}

func (mem WorkgroupMembership) AddDomain(domainId string) {
	_, ok := mem[domainId]
	if !ok {
		mem[domainId] = make(map[string]string)
	}
}

func (mem WorkgroupMembership) AddWorkgroup(domainId, workgroupId, workgroupName string) {
	_, ok := mem[domainId]
	if !ok {
		mem[domainId] = make(map[string]string)
	}
	mem[domainId][workgroupId] = workgroupName
}

func (mem WorkgroupMembership) RemoveDomain(domainId string) {
	delete(mem, domainId)
}

func (mem WorkgroupMembership) RemoveWorkgroupById(domainId, workgroupId string) {
	// TODO
}

func (mem WorkgroupMembership) RemoveWorkgroupByName(domainId, workgroupName string) {
	// TODO
}

type User struct {
	RealmID                string              `json:"realmId"`
	DomainID               string              `json:"domainId"`
	UserID                 string              `json:"userId"`
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
	Workgroups             WorkgroupMembership `json:"wgs"`    // maps Domains to WgIDs to unique names
	Labels                 DomainAssertions    `json:"labels"` // maps Domains to labels
	FFlags                 map[string]any      `json:"fflags"` // Populated by app server
	Keys                   []string            `json:"keys"`
	SessionData            map[string]string   `json:"sessionData"`
}

func Anonymous() User {
	return User{}
}
