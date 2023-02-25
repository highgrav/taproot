# Taproot
## An opinionated webserver for Go development

Taproot is an opinionated webserver framework that provides a rapid development experience for creating web applications
and APIs. Taproot creates a simple `http.Server`-compatible interface that provides sane default behavior for 
configuration, certificate handling, middleware, logging, session management, and observability.

In addition, Taproot provides various capabilities to increase developer velocity:
- Integration with custom user providers
- JSON-based security policies (applied at the route level)
- Server-side Javascript runtime
- Custom view templates

