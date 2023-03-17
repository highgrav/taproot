# Taproot
## An opinionated embedded webserver for Go development

Taproot is an opinionated, embeddable web application server that streamlines development while giving developers the 
ability to intelligently trade performance for flexibility. Taproot provides extensible and sanely-defaulted batteries-included 
solutions for common requirements for modern small-to-midsized applications:

- Logging, IP filtering, request throttling, feature flags, and metrics gathering;
- Automatic TLS management for local certificates, self-generated at runtime, or ACME cert provisioning;
- Integration points for user and session management;
- Built-in metrics and administration servers;
- Acacia, a declarative security policy manager;
- Server-side Javascript and JSML, a JS-based templating language;
- Built-in SSE and Websocket hubs; and
- Asynchronous and cron-style scheduled job processing.



*HERE THERE BE DRAGONS: Taproot is pre-Alpha software! No guarantees are made regarding breaking changes, and at this point we are not avoiding 
breaking changes.*

Check out the `/docs` directory for more information on how to use Taproot.