package taproot

import (
	"errors"
	"github.com/alexedwards/scs/v2"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"net"
	"net/http"
	"time"
)

// Overall configuration structure
type ServerConfig struct {
	ConfigFilePath       string
	FFlagsConfigFilePath string

	/* METRICS */
	UseMetricsServer bool	`mapstructure:"use_metrics_server"`
	UsePprof         bool	`mapstructure:"use_pprof"`

	/* ADMIN SERVER */
	UseAdminServer bool	`mapstructure:"use_admin_server"`
	AdminPort      int	`mapstructure:"admin_server_port"`

	/* REDIRECT */
	UseHttpsRedirectServer bool	`mapstructure:"use_https_redirect_server"`

	/* PANICS */
	PanicStackTraceDirectory string	`mapstructure:"stack_trace_directory"`

	/* HEADER/COOKIE SIGNING */
	RotateSessionSigningKeysEvery time.Duration	`mapstructure:"session_key_duration"`
	GracePeriodForSigningKeys     time.Duration	`mapstructure:"session_key_grace_duration"`
	UseEncryptedSessionTokens     bool			`mapstructure:"encrypt_session_tokens"`

	/* ACADIA SECURITY POLICIES */
	ListenForPolicyChanges   bool		`mapstructure:"acacia_listen_for_policy_changes"`
	SecurityPolicyDir        string		`mapstructure:"acacia_security_policy_dir"`
	DefaultRealm             string		`mapstructure:"default_realm"`
	DefaultDomain            string		`mapstructure:"default_domain"`
	DefaultUserSessionPrefix string		`mapstructure:"default_session_prefix"`

	/* STATIC FILE SERVING */
	StaticUrlPath  string				`mapstructure:"static_file_url_path"`
	StaticFilePath string				`mapstructure:"static_file_path"`

	/* SCRIPTS AND JSML */
	ScriptFilePath       string			`mapstructure:"script_file_path"`
	UseScripts           bool			`mapstructure:"use_scripts"`
	JSMLFilePath         string 		`mapstructure:"jsml_file_path"`					// Where are JSML files stored?
	UseJSML              bool   		`mapstructure:"use_jsml"`						// Use JSML file templates?
	JSMLCompiledFilePath string 		`mapstructure:"jsml_compiled_file_path"`		// A subdirectory under the ScriptFilePath where Taproot will put compiled JSML files

	/* QUEUE */
	WorkHub WorkHubConfig	`mapstructure:"workhub"`

	/* FEATURE FLAGS */
	Flags ffclient.Config 	// Configuration data for feature flag management

	/* SERVER INFO */
	HttpServer     HttpConfig	`mapstructure:"http_server_config"`
	RedirectServer HttpConfig	`mapstructure:"redirect_server_config"`
	MetricsServer  HttpConfig	`mapstructure:"metrics_server_config"`
	AdminServer    HttpConfig	`mapstructure:"admin_server_config"`

	Sessions SessionConfig		`mapstructure:"sessions"`
	FFlags   FFlagConfig		`mapstructure:"fflags"`
}

// Configuration for feature flag management
type FFlagConfig struct {
	Environment           string	`mapstructure:"env"`
	LogFlagUsage          bool		`mapstructure:"log_fflag_usage"`
	PollingIntervalInSecs int		`mapstructure:"polling_interval_secs"`
	OfflineOnly           bool		`mapstructure:"offline_only"`
}

// Configuration for session management
type SessionConfig struct {
	SessionStore        scs.Store
	ContextSessionStore scs.CtxStore
	SessionKeyPrefix    string			`mapstructure:"session_key_prefix"`
	LifetimeInMins      int				`mapstructure:"lifetime_in_mins"`
	IdleTimeoutInMins   int				`mapstructure:"idle_timeout_in_mins"`
	UseCookies          bool			`mapstructure:"use_cookies"`
	CookieName          string			`mapstructure:"cookie_name"`
	CookieDomain        string			`mapstructure:"cookie_domain"`
	CookieHttpOnly      bool			`mapstructure:"cookie_http_only"`
	CookiePath          string			`mapstructure:"cookie_path"`
	CookiePersist       bool			`mapstructure:"cookie_persist"`
	CookieSiteMode      http.SameSite	`mapstructure:"cookie_site_mode"`
	CookieSecure        bool			`mapstructure:"secure_cookie"`
}

// Configuration for the various HTTP servers (web server, HTTP redirect server, metrics server, and admin server)
type HttpConfig struct {
	FriendlyName           string				`mapstructure:"friendly_na,e"`
	ServerName             string				`mapstructure:"server_name"`
	Port                   int					`mapstructure:"port"`
	TLS                    TLSConfig			`mapstructure:"tls"`
	Timeouts               TimeoutConfig		`mapstructure:"timeout"`
	GlobalRateLimits       ApiRateLimitConfig	`mapstructure:"global_rate_limit"`
	IpRateLimits           ApiRateLimitConfig	`mapstructure:"ip_rate_limit"`
	CorsDomains            []string				`mapstructure:"cors_domains"`
	IPFilter               IPFilterConfig		`mapstructure:"ip_filter"`
	LogHandshakeErrorsWith func(error)
}

// Configuration for HTTP server rate limiting (global and per-ip)
type ApiRateLimitConfig struct {
	RequestsPerSecond         int				`mapstructure:"reguests_per_second"`
	BurstableRequests         int				`mapstructure:"burstable_requests"`
	ExemptNets                []net.IPNet
	ExemptNetworks            []string			`mapstructure:"exempt_networks"`
	SweepClientCacheInSeconds int				`mapstructure:"sweep_client_cache_secs"`
}

// Configuration for HTTP server timeouts
type TimeoutConfig struct {
	Server int			`mapstructure:"server"`
	Idle   int			`mapstructure:"idle"`
	Read   int			`mapstructure:"read"`
	Write  int			`mapstructure:"write"`
}

// Configuration for HTTP server TLS.  Note that Taproot will configure TLS using internally-generated self-signed certs, via ACME, or with local key/cert files.
type TLSConfig struct {
	UseTLS            bool			`mapstructure:"use_tls"`
	UseSelfSignedCert bool			`mapstructure:"use_self_signed_cert"`
	UseACME           bool			`mapstructure:"use_acme"`
	ACMEDirectory     string		`mapstructure:"acme_directory"`
	ACMEAllowedHosts  []string		`mapstructure:"acme_allowed_hosts"`
	ACMEHostName      string		`mapstructure:"acme_hostname"`
	LocalCertFilePath string		`mapstructure:"local_cert_filepath"`
	LocalKeyFilePath  string		`mapstructure:"local_key_filepath"`
}

type IPFilterConfig struct {
	BlockByDefault   bool			`mapstructure:"block_by_default"`
	BlockedCountries []string		`mapstructure:"blocked_countries"`
	AllowedCountries []string		`mapstructure:"allowed_countries"`
	BlockedCidrs     []string		`mapstructure:"blocked_cidrs"`
	AllowedCidrs     []string		`mapstructure:"allowed_cdrs"`
}

// Checks to see if a TLS config is valid (that is, does not conflict between using ACME and internally-generated self-signed certs)
func (c *TLSConfig) IsValid() (bool, error) {
	if c.UseSelfSignedCert && c.UseACME {
		return false, errors.New("Cannot use ACME and a self-signed cert!")
	}
	return true, nil
}

type WorkHubConfig struct {
	Name        string			`mapstructure:"name"`
	StorageDir  string			`mapstructure:"storage_path"`
	SegmentSize int				`mapstructure:"segment_size"`
}
