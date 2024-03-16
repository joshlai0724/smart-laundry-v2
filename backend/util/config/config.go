package configutil

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	MaxPasswordAttempts   int16 `mapstructure:"max_password_attempts"`
	MaxUserNameLength     int16 `mapstructure:"max_user_name_length"`
	MaxStoreNameLength    int16 `mapstructure:"max_store_name_length"`
	MaxStoreAddressLength int16 `mapstructure:"max_store_address_length"`
	StorePasswordLength   int16 `mapstructure:"store_password_length"`
	DB                    struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"database"`
	Redis struct {
		Url string `mapstructure:"url"`
	} `mapstructure:"redis"`
	Rabbitmq struct {
		Url string `mapstructure:"url"`
	} `mapstructure:"rabbitmq"`
	Web struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"web"`
	Iot struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"iot"`
	VerCode struct {
		CheckPhoneNumberOwner struct {
			MaxMsgPerTimePeriod int           `mapstructure:"max_msg_per_time_period"`
			TimePeriod          time.Duration `mapstructure:"time_period"`
			LiveTime            time.Duration `mapstructure:"live_time"`
			Length              int           `mapstructure:"length"`
		} `mapstructure:"check_phone_number_owner"`
		ResetPassword struct {
			MaxMsgPerTimePeriod int           `mapstructure:"max_msg_per_time_period"`
			TimePeriod          time.Duration `mapstructure:"time_period"`
			LiveTime            time.Duration `mapstructure:"live_time"`
			Length              int           `mapstructure:"length"`
		} `mapstructure:"reset_password"`
	} `mapstructure:"ver_code"`
	Token struct {
		SymmetricKey         string        `mapstructure:"symmetric_key"`
		AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
		RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
	} `mapstructure:"token"`
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
