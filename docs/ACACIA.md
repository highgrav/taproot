# Acacia Policy Language

Acacia is a declarative security policy language that reduces the burden on developers to write access-control logic 
within business code. With Acacia, you define 

Acacia policies are run before the endpoint logic is called, and can be used to control access to the 
endpoint as well as 

Acacia policy scripts are XML files that contain information about the routes to match, the effects of a policy match, 
and JSON that will be matched against the HTTP request's context.

Taproot uses `github.com/timbray/quamina` to handle high-speed JSON matching for policies.

### Sample Acacia Policies

The following policy simply checks to see if a user is logged in, and redirects them to a login page if not:
~~~
<policy>
    <manifest>
		<id>crm-reject</id>
		<ns>acacia</ns>
		<v>1.0.0</v>
        <name>Lead Reject</name>
        <desc>Policy for redirecting not-logged-in-users from the CRM site</desc>
        <priority>1000</priority>
    </manifest>
    <paths>
        <path>/crm/*</path>
    </paths>
    <effects>
		<redirect>"/app/login"</redirect>
    </effects>
    <matches>
        <match type="json">
            {
                "user":{
                    "userId":[""]
                }
            }
        </match>
    </matches>
</policy>
~~~

This policy checks to see if the user is part of a workgroup with the name "crm::admin" and, if so, injects a number of 
rights into the session context (`r.Context().Val(constantsHTTP_CONTEXT_ACACIA_RIGHTS_KEY)`), so they can be queried by 
the business or presentation logic:
~~~
<policy>
    <manifest>
		<id>crm-admin</id>
		<ns>acacia</ns>
		<v>1.0.0</v>
        <name>Lead Admin</name>
        <desc>Policy for giving admin rights for CRM to admins</desc>
        <priority>10</priority>
    </manifest>
    <paths>
        <path>/crm/view-crm</path>
        <path>/api/v1/crm/*</path>
    </paths>
    <effects>
		<allow>
		    "crm.search"
		    "crm.list"
			"crm.view"
			"crm.edit"
			"crm.delete"
			"crm.add"
		</allow>
    </effects>
    <matches>
        <match type="json">
            {
                "user":{
                    "wgNames":["crm::admin"]
                }
            }
        </match>
    </matches>
</policy>
~~~


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
