package taproot

import (
	"errors"
	"github.com/spf13/viper"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"net"
)

type ServerConfig struct {

	/* METRICS */
	UseMetricsServer bool
	MetricsPort      int

	/* ACADIA SECURITY POLICIES */
	ListenForPolicyChanges bool
	SecurityPolicyDir      string

	/* STATIC FILE SERVING */
	StaticFilePath      string
	StaticFileDirectory string

	/* SCRIPTS AND JSML */
	ScriptFilePath       string
	UseScripts           bool
	JSMLFilePath         string // Where are JSML files stored?
	UseJSMLFiles         bool   // Use JSML file templates?
	JSMLCompiledFilePath string // A subdirectory under the ScriptFilePath where Taproot will put compiled JSML files

	/* FEATURE FLAGS */
	Flags ffclient.Config // Configuration data for feature flag management

	/* SERVER INFO */
	HttpServer     HttpConfig
	RedirectServer HttpConfig
	MetricsServer  HttpConfig
	AdminServer    HttpConfig
}

type HttpConfig struct {
	ServerName       string
	Port             int
	TLS              TLSConfig
	Timeouts         TimeoutConfig
	Session          SessionConfig
	GlobalRateLimits ApiRateLimitConfig
	IpRateLimits     ApiRateLimitConfig
	CorsDomains      []string
}

type ApiRateLimitConfig struct {
	RequestsPerSecond         int
	BurstableRequests         int
	ExemptNets                []net.IPNet
	SweepClientCacheInSeconds int
}

type TimeoutConfig struct {
	Server int
	Idle   int
	Read   int
	Write  int
}

type TLSConfig struct {
	UseTLS              bool
	UseHTTPSRedirection bool // Start a port 80 Server to force redirects to the main port (this will automatically take place if using ACME)
	UseSelfSignedCert   bool
	UseACME             bool
	ACMEDirectory       string
	ACMEAllowedHost     string
	ACMEHostName        string
	LocalCertFilePath   string
	LocalKeyFilePath    string
}

type SessionConfig struct {
}

func (c *TLSConfig) IsValid() (bool, error) {
	if c.UseSelfSignedCert && c.UseACME {
		return false, errors.New("Cannot use ACME and a self-signed cert!")
	}
	return true, nil
}

// LoadConfig() handles all the Viper setup and management
func LoadConfig(cfgDirs []string) (ServerConfig, error) {
	cfg := ServerConfig{}
	viper.SetConfigName("taproot")
	for _, v := range cfgDirs {
		viper.AddConfigPath(v)
	}
	return cfg, nil
}
