# Acacia Authentication Language

The Acacia authentication language lets you apply JSON-based policies to application routes.

### Example
{
    "manifest":{
        "ns":"acacia",
        "v":"1.0.0",
        "name":"my policy",
        "desc":"A sample policy"
    },
    "paths": [
            "/api/v1/user/:id"
    ],
    "rights":{
        "permit":[
        ],
        "deny":[
        ]
    },
    "log":{
        "permit":[],
        "deny":[],
        "any":[],
    },
    "match":{
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
}