# Metrics Server
Taproot has the ability to start a metrics server that provides basic information about system performance. You can configure 
the metrics server at startup.

Like the admin server, the metrics server should never be exposed to the outside world, and it is recommended to both 
set IP filtering on the server through configuration as well as filtering at the firewall level.

### Endpoints
The following endpoints are supported:
- `/`: Returns a list of all endpoints that have metrics gathered for them. Note that this is not all registered endpoints;
metrics are not gathered for an endpoint until it has been hit by a user.
- `/global`: Returns global server metrics
- `/stats?path=/some/path`: Returns metrics for `/some/path`.

If you set `UsePprof` on the `ServerConfig` struct, then a selected set of pprof endpoints will also be available 
starting at `/debug/pprof`.