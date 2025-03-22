package solid

import "github.com/spf13/viper"

type Config struct {
	Prod bool `mapstructure:"PROD"`

	GrpcPort string `mapstructure:"GRPC_PORT"`

	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	S3Endpoint string `mapstructure:"S3_ENDPOINT"`
	S3Key      string `mapstructure:"S3_KEY"`
	S3Secret   string `mapstructure:"S3_SECRET"`
	S3Bucket   string `mapstructure:"S3_BUCKET"`
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.AddConfigPath("/etc/mlsolid/")
	viper.SetConfigName("mlsolid")
	viper.SetConfigName("yaml")

	viper.SetDefault("PROD", "true")

	viper.SetDefault("GRPC_PORT", "5000")

	viper.SetDefault("REDIS_ADDR", "redis:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("S3_ENDPOINT", "")
	viper.SetDefault("S3_KEY", "")
	viper.SetDefault("S3_SECRET", "")
	viper.SetDefault("S3_BUCKET", "")

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
