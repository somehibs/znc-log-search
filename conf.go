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
	Whitelist []string // for whitelisting specific channels
	Sphinx SphinxConfig
	Queues map[string]int
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
	conf.Load(envvar.NewSource(), flag.NewSource(), file.NewSource(file.WithPath("./default_config.json")), file.NewSource(file.WithPath(filename)))
	var confObj = LogsConfig{}
	conf.Get().Scan(&confObj)
	cachedFile = filename
	cache = confObj
	return confObj
}
