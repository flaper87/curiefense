package pkg

import "github.com/spf13/viper"

func NewConfig() *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()
	return v
}
