package acaciaparser

import "testing"

func TestAcacia(t *testing.T) {
	input := `
	<policy>
    <manifest ns="acacia" v="1.0.0">
        <name>My Policy</name>
        <desc>A sample policy</desc>
        <priority>10</priority>
    </manifest>
    <paths>
        <path name="/api/v1/user/:id"/>
    </paths>
    <effects>
        <rights>
			"ml.denied"
        </rights>
        <redirect>"/to/some/:id"</redirect>
		<deny 300>User not allowed</deny>
    </effects>
    <log>
        <permit>
        </permit>
        <deny>
			<info>User was denied access</info>
        </deny>
        <all>
        </all>
    </log>
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
	p.Parse()
}
