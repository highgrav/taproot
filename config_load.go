package taproot

import "github.com/spf13/viper"

const (
	DefaultConfigFile                 string = "taproot.yaml"
	DefaultAppServerIPFiltersFile     string = "taproot_ipf.yaml"
	DefaultMetricsServerIPFiltersFile string = "metrics_ipf.yaml"
	DefaultAdminServerIPFiltersFile   string = "admin_ipf.yaml"
)

// LoadConfig() handles all the Viper setup and management
func loadConfig(cfgDirs []string) (ServerConfig, error) {
	cfg := ServerConfig{}
	viper.SetConfigName("taproot")
	for _, v := range cfgDirs {
		viper.AddConfigPath(v)
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("ipfilters")
	for _, v := range cfgDirs {
		viper.AddConfigPath(v)
	}
	err = viper.MergeInConfig()
	if err != nil {
		panic(err)
	}

	return cfg, nil
}
