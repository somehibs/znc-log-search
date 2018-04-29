package logs

import (
"github.com/micro/go-config"
"github.com/micro/go-config/source/envvar"
"github.com/micro/go-config/source/file"
"github.com/micro/go-config/source/flag"
)

type SphinxConfig struct {
		Host string
		Port int
		User string
}

type LogsConfig struct {
	Network string
	Sphinx SphinxConfig
}

func GetConf() LogsConfig {
	return GetConfByName("./config.json")
}

func GetConfByName(filename string) LogsConfig {
	var conf = config.NewConfig()
	conf.Load(envvar.NewSource(), flag.NewSource(), file.NewSource(file.WithPath(filename)))
	var confObj = LogsConfig{}
	conf.Get().Scan(&confObj)
	return confObj
}
