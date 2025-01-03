package config

type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // postgres, mysql, sqlite, mssql
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}
