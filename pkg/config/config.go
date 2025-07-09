package config

import (
	"fmt"
	"log"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var instance *Config

type EmailConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	Admin    string `mapstructure:"admin"`
}

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Stripe   StripeConfig   `mapstructure:"stripe"`
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	Email    EmailConfig    `mapstructure:"email"`
	BaseURL  string         `mapstructure:"base_url"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type StripeConfig struct {
	SecretKey     string `mapstructure:"secret_key"`
	PublicKey     string `mapstructure:"public_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
}

type OpenAIConfig struct {
	APIKey string `mapstructure:"api_key"`
}

func LoadConfig() (*Config, error) {
	// Try loading docker.env first, then .env
	if err := godotenv.Load("docker.env"); err != nil {
		log.Println("Warning: Could not find docker.env file, trying .env")
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: Could not find .env file, relying on system environment variables and defaults.")
		}
	}

	viper.AutomaticEnv()

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("server.port", "APP_PORT")
	viper.BindEnv("server.mode", "GIN_MODE")
	viper.BindEnv("stripe.secret_key", "STRIPE_SECRET_KEY")
	viper.BindEnv("stripe.public_key", "STRIPE_PUBLIC_KEY")
	viper.BindEnv("stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET")
	viper.BindEnv("openai.api_key", "OPENAI_API_KEY")

	viper.BindEnv("email.username", "EMAIL_USERNAME")
	viper.BindEnv("email.password", "EMAIL_PASSWORD")
	viper.BindEnv("email.from", "EMAIL_FROM")
	viper.BindEnv("email.admin", "EMAIL_ADMIN")

	viper.BindEnv("base_url", "BASE_URL")
	log.Printf("!!!Stripe secret key: %s", viper.GetString("stripe.secret_key"))
	log.Printf("!!!Stripe public key: %s", viper.GetString("stripe.public_key"))
	log.Printf("!!!Base URL: %s", viper.GetString("base_url"))
	log.Printf("!!!Email username: %s", viper.GetString("email.username"))
	log.Printf("!!!Email password: %s", viper.GetString("email.password"))
	log.Printf("!!!Email from: %s", viper.GetString("email.from"))
	log.Printf("!!!Admin email: %s", viper.GetString("email.admin"))

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "anastasia_gofman")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("server.port", "8010")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("stripe.secret_key", "sk_test_")
	viper.SetDefault("stripe.public_key", "pk_test_")
	viper.SetDefault("stripe.webhook_secret", "")
	viper.SetDefault("openai.api_key", "")
	viper.SetDefault("base_url", "http://91.105.196.19:8080/")
	viper.SetDefault("email.username", "")
	viper.SetDefault("email.password", "")
	viper.SetDefault("email.from", "")
	viper.SetDefault("email.admin", "")

	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	instance = &config
	return instance, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", cfg.Database.Port)
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid server port: %s", cfg.Server.Port)
	}

	validModes := map[string]bool{"debug": true, "release": true, "test": true}
	if !validModes[cfg.Server.Mode] {
		return fmt.Errorf("invalid gin mode: %s (must be debug, release, or test)", cfg.Server.Mode)
	}

	return nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func GetBaseURL() string {
	if viper.GetString("base_url") == "" {
		return "http://91.105.196.19:8080/"
	}
	return viper.GetString("base_url")
}

func GetConfig() *Config {
	return instance
}
