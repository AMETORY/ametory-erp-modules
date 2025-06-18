package config

type ServerConfig struct {
	AppName         string `mapstructure:"app_name"`
	Port            string `mapstructure:"port"`
	SecretKey       string `mapstructure:"secret_key"`
	BaseURL         string `mapstructure:"base_url"`
	TokenExpiredDay int    `mapstructure:"token_expired_day"`
}
