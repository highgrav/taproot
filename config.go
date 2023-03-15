package taproot

import (
	"errors"
	"github.com/alexedwards/scs/v2"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"net"
	"net/http"
)

// Overall configuration structure
type ServerConfig struct {
	ConfigFilePath         string
	IPFilterConfigFilePath string
	FFlagsConfigFilePath   string

	/* METRICS */
	UseMetricsServer bool
	MetricsPort      int

	UseAdminServer bool
	AdminPort      int

	UseHttpsRedirectServer bool

	/* ACADIA SECURITY POLICIES */
	ListenForPolicyChanges bool
	SecurityPolicyDir      string

	/* STATIC FILE SERVING */
	StaticUrlPath       string
	StaticFileDirectory string

	/* SCRIPTS AND JSML */
	ScriptFilePath       string
	UseScripts           bool
	JSMLFilePath         string // Where are JSML files stored?
	UseJSML              bool   // Use JSML file templates?
	JSMLCompiledFilePath string // A subdirectory under the ScriptFilePath where Taproot will put compiled JSML files

	/* QUEUE */
	AsyncQueue MQueueConfig

	/* FEATURE FLAGS */
	Flags ffclient.Config // Configuration data for feature flag management

	/* SERVER INFO */
	HttpServer     HttpConfig
	RedirectServer HttpConfig
	MetricsServer  HttpConfig
	AdminServer    HttpConfig

	Sessions SessionConfig
	FFlags   FFlagConfig
}

// Configuration for feature flag management
type FFlagConfig struct {
	Environment           string
	LogFlagUsage          bool
	PollingIntervalInSecs int
	OfflineOnly           bool
}

// Configuration for session management
type SessionConfig struct {
	SessionStore        scs.Store
	ContextSessionStore scs.CtxStore
	LifetimeInSecs      int
	IdleTimeoutInSecs   int
	UseCookies          bool
	CookieName          string
	CookieDomain        string
	CookieHttpOnly      bool
	CookiePath          string
	CookiePersist       bool
	CookieSiteMode      http.SameSite
	CookieSecure        bool
}

// Configuration for the various HTTP servers (web server, HTTP redirect server, metrics server, and admin server)
type HttpConfig struct {
	FriendlyName     string
	ServerName       string
	Port             int
	TLS              TLSConfig
	Timeouts         TimeoutConfig
	GlobalRateLimits ApiRateLimitConfig
	IpRateLimits     ApiRateLimitConfig
	CorsDomains      []string
	IPFilter         IPFilterConfig
}

// Configuration for HTTP server rate limiting (global and per-ip)
type ApiRateLimitConfig struct {
	RequestsPerSecond         int
	BurstableRequests         int
	ExemptNets                []net.IPNet
	ExemptNetworks            []string
	SweepClientCacheInSeconds int
}

// Configuration for HTTP server timeouts
type TimeoutConfig struct {
	Server int
	Idle   int
	Read   int
	Write  int
}

// Configuration for HTTP server TLS.  Note that Taproot will configure TLS using internally-generated self-signed certs, via ACME, or with local key/cert files.
type TLSConfig struct {
	UseTLS            bool
	UseSelfSignedCert bool
	UseACME           bool
	ACMEDirectory     string
	ACMEAllowedHost   string
	ACMEHostName      string
	LocalCertFilePath string
	LocalKeyFilePath  string
}

type IPFilterConfig struct {
	BlockByDefault   bool
	BlockedCountries []string
	AllowedCountries []string
	BlockedCidrs     []string
	AllowedCidrs     []string
}

// Checks to see if a TLS config is valid (that is, does not conflict between using ACME and internally-generated self-signed certs)
func (c *TLSConfig) IsValid() (bool, error) {
	if c.UseSelfSignedCert && c.UseACME {
		return false, errors.New("Cannot use ACME and a self-signed cert!")
	}
	return true, nil
}

type MQueueConfig struct {
	Name        string
	StorageDir  string
	SegmentSize int
}
