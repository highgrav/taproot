# GETTING STARTED

To get started with Taproot, you'll need to create an `authn.IUserStore`-conformant class that lets Taproot access 
user data, and a `session.IStore`-conformant class that lets Taproot persist session data. (For convenience, `IStore` is 
compliant with a subset of the `Store` interface in `github.com/alexedwards/scs`, so you can use any of the pre-existing 
stores from that package). Also set up a `deck`-compliant logging provider to route logging (of which there will be a lot).

From there, you need to create a `Config` object, add any `sql`-compliant databases, additional middleware, websocket or SSE 
hubs, asynchronous work handlers, scheduled cronjobs, and your `http` handlers.

We recommend creating a `/site` directory to hold any server-side web code. Our practice is to have a `/site/views` 
directory for JSML files, `/site/~scripts` for compiled JS scripts, `/site/static` for static files, and `/site/policies` 
for Acacia route-level security policy files.

See `QUICKSTART.md` for more.