# Acacia Policy Language

Acacia is a convenient way to write policy scripts that can deny users, redirect them, or inject allowed rights at the 
endpoint level.

Acacia policy scripts are XML files that contain information about the routes to match, the effects of a policy match, 
and JSON that will be matched against the HTTP request's context.

Taproot uses `github.com/timbray/quamina` to handle high-speed JSON matching for policies.

### Rights Request Format
~~~
rightsRequest: {
    "securityRealmId":"...",
    "securityDomainId":"...",
    "time":"2023-04-05 12:00:00",
    "context":{
        "somekey", "somevalue"
    },
    "request":{
        "ip":"...",
        "headers":{
            
        },
        "method":"GET",
        "route":"/some/:tenantId/route/:id",
        "url":"/some/12345/route/456456",
        "params":{
            "tenantId":"12345",
            "id":"456456"
        },
        "queryString":"",
        "body":""
    },
    "user:":{
        "realmId":"...",
        "domainId":"...",
        "userId":"...",
        "username":"...",
        "displayName":"...",
        "emails":["..."],
        "phones":["..."],
        "isVerified":true,
        "isBlocked": false,
        "isActive": true,
        "IsDeleted": false,
        "requiresPasswordUpdate": false,
        "wgs":["..."],
        "labels": ["..."],
        "fflags" :{
            "someFlag":true,
            "someOtherFlag":"some-value"
        }
    }
}
~~~

The `user` entity is a domain-specific version of the `authn.User` struct.
