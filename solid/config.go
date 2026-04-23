package solid

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// Config configuration struct.
type Config struct {
	RootURL     string `mapstructure:"root_url"`
	FrontendURL string `mapstructure:"frontend_url"`
	Prod        bool   `mapstructure:"prod"`

	APIKeyAccess         bool     `mapstructure:"api_key_access"`
	GoogleClientID       string   `mapstructure:"google_client_id"`
	GoogleSecretID       string   `mapstructure:"google_secret_id"`
	GoogleAllowedDomains []string `mapstructure:"google_allowed_domains"`

	APIPort  string `mapstructure:"api_port"`
	GrpcPort string `mapstructure:"grpc_port"`

	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	S3Endpoint string `mapstructure:"s3_endpoint"`
	S3Key      string `mapstructure:"s3_key"`
	S3Secret   string `mapstructure:"s3_secret"`
	S3Bucket   string `mapstructure:"s3_bucket"`
	S3Region   string `mapstructure:"s3_region"`
	S3Prefix   string `mapstructure:"s3_prefix"`

	EnableBEngine          bool   `mapstructure:"enable_bengine"`
	DockerRegistryUsername string `mapstructure:"docker_registry_username"`
	DockerRegistryPassword string `mapstructure:"docker_registry_password"`
	HostSourceVolume       string `mapstructure:"host_source_volume"`
}

// LoadConfig loads mlsolid's configuration file from the path specified.
// It also checks for any configuration file under /etc/mlsolid.
func LoadConfig(path string) (Config, error) {
	viper.SetConfigName("mlsolid")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath("/etc/mlsolid/")

	viper.SetDefault("root_url", "http://localhost:8050")
	viper.SetDefault("frontend_url", "http://localhost:5173")
	viper.SetDefault("prod", "true")

	viper.SetDefault("api_key_access", false)
	viper.SetDefault("google_client_id", "***")
	viper.SetDefault("google_secret_id", "***")
	viper.SetDefault("google_allowed_domains", []string{})

	viper.SetDefault("api_port", "8050")
	viper.SetDefault("grpc_port", "5000")

	viper.SetDefault("redis_addr", "redis:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)

	viper.SetDefault("s3_endpoint", "")
	viper.SetDefault("s3_key", "")
	viper.SetDefault("s3_secret", "")
	viper.SetDefault("s3_bucket", "")
	viper.SetDefault("s3_region", "")
	viper.SetDefault("s3_prefix", "artifacts")

	viper.SetDefault("enable_bengine", false)
	viper.SetDefault("docker_registry_username", "")
	viper.SetDefault("docker_registry_password", "")
	viper.SetDefault("host_source_volume", "")

	viper.AutomaticEnv()

	config := Config{} //nolint: exhaustruct

	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return config, fmt.Errorf("configuration file not found: %w", err)
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("could not parse config: %w", err)
	}

	return config, nil
}
