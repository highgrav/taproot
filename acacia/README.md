## Acacia Authentication Language

The Acacia authentication language lets you apply JSON-based policies to application routes.

### Example
`<policy>
    <manifest ns="acacia" v="1.0.0">
        <name>My Policy</name>
        <desc>A sample policy</desc>
    </manifest>
    <paths>
        <path name="/api/v1/user/:id"/>
    </paths>
    <rights>
        <permit>
        </permit>
        <deny>
        </deny>
    </rights>
    <log>
        <permit>
        </permit>
        <deny>
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
</policy>`
