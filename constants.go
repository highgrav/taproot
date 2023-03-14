package taproot

const (
	HTTP_CONTEXT_CORRELATION_KEY string = "taproot--app-corr-id"

	HTTP_CONTEXT_CSP_NONCE_KEY     string = "taproot--csp-nonce" // identifies the CSP nonce that can be used to whitelist inline scripts and styles in templates (automatically injected in the strict-header middleware)
	HTTP_CONTEXT_USER_KEY          string = "taproot--user"
	HTTP_CONTEXT_REALM_KEY         string = "taproot--realm"
	HTTP_CONTEXT_DOMAIN_KEY        string = "taproot--domain"
	HTTP_CONTEXT_ACACIA_RIGHTS_KEY string = "taproot--rights"
)
