package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	PostgresHost        string
	PostgresPort        string
	PostgresUser        string
	PostgresPassword    string
	PostgresDB          string
	RabbitMQURL         string
	JWTSecretKey        string
	ServicesNetworkType string
	ProductPort         string
	OrderPort           string
	CartPort            string
	UserPort            string
	NotificationPort    string
	GatewayPort         string
	RedisURL            string
	RedisPassword       string
	RedisDB             int
	SMTPFrom            string
	SMTPUsername        string
	SMTPPass            string
	SMTPHost            string
	SMTPPort            int
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("/internet-shop/shared/config/config.env")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := &Config{
		PostgresHost:        viper.GetString("POSTGRES_HOST"),
		PostgresPort:        viper.GetString("POSTGRES_PORT"),
		PostgresUser:        viper.GetString("POSTGRES_USER"),
		PostgresPassword:    viper.GetString("POSTGRES_PASSWORD"),
		PostgresDB:          viper.GetString("POSTGRES_DB"),
		RabbitMQURL:         viper.GetString("RABBITMQ_URL"),
		JWTSecretKey:        viper.GetString("JWT_SECRET_KEY"),
		ServicesNetworkType: viper.GetString("SERVICES_NETWORK_TYPE"),
		ProductPort:         viper.GetString("PRODUCT_PORT"),
		OrderPort:           viper.GetString("ORDER_PORT"),
		CartPort:            viper.GetString("CART_PORT"),
		UserPort:            viper.GetString("USER_PORT"),
		NotificationPort:    viper.GetString("NOTIFICATION_PORT"),
		GatewayPort:         viper.GetString("GATEWAY_PORT"),
		RedisURL:            viper.GetString("REDIS_URL"),
		RedisPassword:       viper.GetString("REDIS_PASSWORD"),
		RedisDB:             viper.GetInt("REDIS_DB"),
		SMTPFrom:            viper.GetString("SMTP_FROM"),
		SMTPUsername:        viper.GetString("SMTP_FROM"),
		SMTPPass:            viper.GetString("SMTP_PASS"),
		SMTPHost:            viper.GetString("SMTP_HOST"),
		SMTPPort:            viper.GetInt("SMTP_PORT"),
	}
	return config, nil
}

func (cfg *Config) PostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
}

func (cfg *Config) LocalhostURL(cfgPort string) string {
	return fmt.Sprintf("localhost%s", cfgPort)
}
