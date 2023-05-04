package authn

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
	} else {
		for _, v := range mem[domainId] {
			if v.ID == workgroupId {
				return
			}
		}
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
