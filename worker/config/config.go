package config

import (
	"github.com/spf13/viper"
)

var G_settings settings

type settings struct {
	Server Server `toml:"server"`
	Etcd   Etcd   `toml:"etcd"`
}

type Server struct {
	Port int `toml:"port"`
}

type Etcd struct {
	EndPoints   []string `toml:"endPoints"`
	DialTimeout int64    `toml:"dialTimeout"`
}

func init() {
	vp := viper.New()
	// 从项目根目录启动 go run cmd/worker/main.go （相对项目根目录，否则路径不对）
	vp.AddConfigPath("./worker/config/")
	vp.SetConfigName("config")
	vp.SetConfigType("toml")
	if err := vp.ReadInConfig(); err != nil {
		panic(err.Error())
	}

	if err := vp.Unmarshal(&G_settings); err != nil {
		panic(err.Error())
	}
}
