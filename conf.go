package logs

import (
"github.com/micro/go-config"
"github.com/micro/go-config/source/envvar"
"github.com/micro/go-config/source/file"
"github.com/micro/go-config/source/flag"
)

type SphinxConfig struct {
		host string
		port int
		user string
}

type LogsConfig struct {
	network string
	sphinxql SphinxConfig
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
