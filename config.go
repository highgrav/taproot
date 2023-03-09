package taproot

import (
	"errors"
	"github.com/alexedwards/scs/v2"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"net"
	"net/http"
)

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

type FFlagConfig struct {
	Environment           string
	LogFlagUsage          bool
	PollingIntervalInSecs int
	OfflineOnly           bool
}

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

type HttpConfig struct {
	ServerName       string
	Port             int
	TLS              TLSConfig
	Timeouts         TimeoutConfig
	GlobalRateLimits ApiRateLimitConfig
	IpRateLimits     ApiRateLimitConfig
	CorsDomains      []string
}

type ApiRateLimitConfig struct {
	RequestsPerSecond         int
	BurstableRequests         int
	ExemptNets                []net.IPNet
	ExemptNetworks            []string
	SweepClientCacheInSeconds int
}

type TimeoutConfig struct {
	Server int
	Idle   int
	Read   int
	Write  int
}

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

func (c *TLSConfig) IsValid() (bool, error) {
	if c.UseSelfSignedCert && c.UseACME {
		return false, errors.New("Cannot use ACME and a self-signed cert!")
	}
	return true, nil
}
