package acaciaparser

import (
	"fmt"
	"testing"
)

func TestAcacia(t *testing.T) {
	input := `
	<policy>
    <manifest>
		<id>My-Policy</id>
		<ns>acacia</ns>
		<v>1.0.0</v>
        <name>My Policy</name>
        <desc>A sample policy</desc>
        <priority>10</priority>
    </manifest>
    <paths>
        <path>/api/v1/crm/:id</path>
		<path>/app/:tenantId/crm/:id</path>
    </paths>
    <effects>
		<allow>	
			"crm.search.self"
			"crm.read"
		</allow>
        <deny>
			"crm.search.all"
			"crm.write"
			"crm.create"
			"crm.delete"
			"crm.admin"
        </deny>
        <redirect>"/user/read/crm/:id"</redirect>
		<return>User not allowed</return>
		<returncode>300</returncode>
    </effects>
    <matches>
        <match type="json">
            {
                "ip":"",
                "domain":[],
                "userId":[],
                "wgId":[],
                "labels":[]
                "query":{
                },
                "body":{
                } 
            }
        </match>
    </matches>
</policy>
`
	p, _ := New(input)
	policy, err := p.Parse()
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Printf("%+v\n", policy)
}
