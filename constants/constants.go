package constants

const (
	// Context key for the tracing correlation ID
	HTTP_CONTEXT_CORRELATION_KEY string = "taproot--app-corr-id"

	// Context key for content security policy nonce (useful in JSML, only needed if sending security headers)
	HTTP_CONTEXT_CSP_NONCE_KEY string = "taproot--csp-nonce" // identifies the CSP nonce that can be used to whitelist inline scripts and styles in templates (automatically injected in the strict-header middleware)

	// Context key for holding the injected user for the request
	HTTP_CONTEXT_USER_KEY string = "taproot--user"

	// Context key for holding the realm ID for the request
	HTTP_CONTEXT_REALM_KEY string = "taproot--realm"

	// Context key for holding the security domain ID for the request
	HTTP_CONTEXT_DOMAIN_KEY string = "taproot--domain"

	// Context key for holding any permissions returned from an Acacia policy application
	HTTP_CONTEXT_ACACIA_RIGHTS_KEY string = "taproot--rights"

	HTTP_CONTEXT_SESSION_KEY string = "taproot--skey"

	HTTP_CONTEXT_FFLAG_KEY string = "taproot--fflags"
)
