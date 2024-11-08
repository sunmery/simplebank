package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ENVIRONMENT          string        `mapstructure:"ENVIRONMENT"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	RedisPassword        string        `mapstructure:"REDIS_PASSWORD"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SEVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SEVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration string        `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

func LoadConfig(path string) (cfg *Config, err error) {
	viper.SetConfigName("app") // name of config file (without extension)
	viper.SetConfigType("env") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)  // optionally look for config in the working directory
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 自动从环境变量替换配置文件
	viper.AutomaticEnv()

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return
	}

	return
}
