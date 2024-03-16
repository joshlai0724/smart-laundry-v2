package configutil

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	StoreID  string `mapstructure:"store_id"`
	Password string `mapstructure:"password"`
	DB       struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"database"`
	Iot struct {
		Url           string        `mapstructure:"url"`
		RetryInterval time.Duration `mapstructure:"retry_interval"`
	} `mapstructure:"iot"`
	Mosquitto struct {
		Url      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"mosquitto"`
	RecordsResend struct {
		Interval  time.Duration `mapstructure:"interval"`
		BatchSize int32         `mapstructure:"batch_size"`
	} `mapstructure:"records_resend"`
	Info struct {
		EdgeVersionFile string `mapstructure:"edge_version_file"`
	} `mapstructure:"info"`
}

func Load(env string) (config Config, err error) {
	viper.AddConfigPath("./configs")
	if len(env) > 0 {
		viper.SetConfigName(env)
	} else {
		viper.SetConfigName("dev") // default
	}
	viper.SetConfigType("toml")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if len(env) > 0 {
		viper.SetConfigName(env + "-vars")
	} else {
		viper.SetConfigName("dev-vars") // default
	}
	viper.SetConfigType("toml")

	if err = viper.MergeInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func GetConfigFile() string {
	return viper.ConfigFileUsed()
}
