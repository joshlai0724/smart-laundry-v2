package configutil

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DeviceID       string        `mapstructure:"device_id"`
	Points         int32         `mapstructure:"points"`
	State          string        `mapstructure:"state"`
	BeaconInterval time.Duration `mapstructure:"beacon_interval"`
	Mosquitto      struct {
		Url      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"mosquitto"`
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

	err = viper.Unmarshal(&config)
	return
}

func GetConfigFile() string {
	return viper.ConfigFileUsed()
}
