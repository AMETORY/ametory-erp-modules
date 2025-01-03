package config

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
}

var App *Config

// LoadConfig memuat konfigurasi dari file atau environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // Nama file config (tanpa ekstensi)
	viper.SetConfigType("yaml")   // Format file config (yaml, json, toml, dll.)
	viper.AddConfigPath(".")      // Path ke file config (direktori saat ini)
	viper.AutomaticEnv()          // Baca environment variables

	// Baca file config
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("gagal membaca file config: %v", err)
	}

	// Unmarshal config ke struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("gagal unmarshal config: %v", err)
	}

	App = &cfg
	return &cfg, nil
}

// InitDB menginisialisasi koneksi database berdasarkan konfigurasi
func InitDB(cfg *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Database.Name) // Name adalah path ke file SQLite
	case "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
		dialector = sqlserver.Open(dsn)
	default:
		return nil, fmt.Errorf("database type tidak didukung: %s", cfg.Database.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungkan ke database: %v", err)
	}

	return db, nil
}
