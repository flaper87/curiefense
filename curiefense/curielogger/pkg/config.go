package pkg

import (
	"github.com/curiefense/curiefense/curielogger/pkg/outputs"
	"github.com/spf13/viper"
	"gopkg.in/dealancer/validate.v2"
	"strings"
)

func NewConfig() (*viper.Viper, outputs.Config) {
	v := viper.New()
	v.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.AddConfigPath("./../")
	viper.AddConfigPath("/etc/curielogger/")
	viper.SetConfigName("curielogger")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("CURIELOGGER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg outputs.Config
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	err = validate.Validate(&cfg)
	return v, cfg
}
