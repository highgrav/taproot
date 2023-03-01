package authn

type DomainAssertions map[string]string

type WorkgroupMembership map[string]map[string]string

// (Defining a pointer receiver on this seems non-idiomatic)
func (mem WorkgroupMembership) ByDomain(domainId string) map[string]string {
	return mem[domainId]
}

func (mem WorkgroupMembership) AddDomain(domainId string) {
}

func (mem WorkgroupMembership) AddWorkgroup(domainId, workgroupId, workgroupName string) {

}

func (mem WorkgroupMembership) RemoveDomain(domainId string) {

}

func (mem WorkgroupMembership) RemoveWorkgroupById(domainId, workgroupId string) {

}

func (mem WorkgroupMembership) RemoveWorkgroupByName(domainId, workgroupName string) {

}

type User struct {
	RealmID                string              `json:"realmId"`
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
}

func Anonymous() User {
	return User{}
}
