## Acacia Authentication Language

The Acacia authentication language lets you apply JSON-based policies to application routes.

Acacia policies:
- Apply at the route level;
- Use Quamina JSON-matching patterns;
- Can return a list of strings, deny access to a route, or redirect to another route;
- Can trigger specific log events.

Acacia is designed to be used after a user is bound to an HTTP request using the `authn.User` model. However, Acacia 
itself makes no assumptions about the shape of the data passed to it, so you can inject custom middleware into the stack 
and then call Acacia manually to test if a match exists.

Acacia 

### Example
~~~
<policy>
    <manifest ns="acacia" v="1.0.0">
        <name>My Policy</name>
        <desc>A sample policy</desc>
    </manifest>
    <paths>
        <path name="/api/v1/user/:id"/>
    </paths>
    <rights>
        <permit>
            <right></right>
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
</policy>
~~~