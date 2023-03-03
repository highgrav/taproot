package taproot

import "github.com/spf13/viper"

// LoadConfig() handles all the Viper setup and management
func LoadConfig(cfgDirs []string) (ServerConfig, error) {
	cfg := ServerConfig{}
	viper.SetConfigName("taproot")
	for _, v := range cfgDirs {
		viper.AddConfigPath(v)
	}
	return cfg, nil
}
