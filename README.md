# Taproot
## An opinionated embedded webserver for Go development

Taproot is an opinionated, embeddable web application server that streamlines development while giving developers the 
ability to intelligently trade performance for flexibility. 

Taproot provides extensible and sanely-defaulted batteries-included solutions for common requirements for modern 
small-to-midsized applications, with an emphasis on enabling rapid prototyping. Essentially, it intends to give you a quick 
start on web development and get out of your way.

Taproot includes:
- Logging, IP filtering, request throttling, bot/crawler detection, and metrics gathering;
- Feature flag integration;
- Automatic TLS management for local certificates, self-generated at runtime, or ACME cert provisioning;
- Integration points for user and session management, using encrypted or signed cookies or headers;
- Built-in metrics and administration servers;
- Acacia, a declarative security policy manager;
- Server-side Javascript and JSML, a JS-based templating language;
- Built-in SSE and Websocket hubs;
- Asynchronous and cron-style scheduled background job processing.

*HERE THERE BE DRAGONS: Taproot is pre-Alpha software! No guarantees are made regarding breaking changes, and at this point we are not avoiding 
breaking changes. Test scenarios are not currently implemented to the degree necessary. Things are likely to break in new, unexpected, and unfortunately exciting ways.*


Check out the `/docs` directory for more information on how to use Taproot.