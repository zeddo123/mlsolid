package solid

import (
	"github.com/spf13/viper"
)

type Config struct {
	Prod bool `mapstructure:"prod"`

	GrpcPort string `mapstructure:"grpc_port"`

	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	S3Endpoint string `mapstructure:"s3_endpoint"`
	S3Key      string `mapstructure:"s3_key"`
	S3Secret   string `mapstructure:"s3_secret"`
	S3Bucket   string `mapstructure:"s3_bucket"`
}

func LoadConfig(path string) (Config, error) {
	viper.SetConfigName("mlsolid")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath("/etc/mlsolid/")

	viper.SetDefault("prod", "true")

	viper.SetDefault("grpc_port", "5000")

	viper.SetDefault("redis_addr", "redis:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)

	viper.SetDefault("s3_endpoint", "")
	viper.SetDefault("s3_key", "")
	viper.SetDefault("s3_secret", "")
	viper.SetDefault("s3_bucket", "")

	viper.AutomaticEnv()

	config := Config{}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
