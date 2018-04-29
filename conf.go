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

type ArangoConfig struct {
	Endpoints []string
	User string
	Password string
}

type LogsConfig struct {
	Network string
	Whitelist []string // for whitelisting specific channels
	Sphinx SphinxConfig
	Queues map[string]int
	Arango ArangoConfig
}

var cachedFile = ""
var cache = LogsConfig{Network: "_not_a_real_network_probably"}

func GetConf() LogsConfig {
	return GetConfByName("./config.json")
}

func GetConfByName(filename string) LogsConfig {
	if cachedFile == filename {
		return cache
	}
	var conf = config.NewConfig()
	conf.Load(file.NewSource(file.WithPath("./default_config.json")))
	conf.Load(file.NewSource(file.WithPath(filename)))
	conf.Load(envvar.NewSource(), flag.NewSource())
	var confObj = LogsConfig{}
	conf.Get().Scan(&confObj)
	cachedFile = filename
	cache = confObj
	return confObj
}
