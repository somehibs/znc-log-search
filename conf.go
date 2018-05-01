package logs

import (
	"fmt"

	"github.com/micro/go-config"
	"github.com/micro/go-config/source/envvar"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/flag"
)

type SphinxConfig struct {
		Dsn string
}

type ArangoConfig struct {
	Endpoints []string
	User string
	Password string
	Db string
}

type LogsConfig struct {
	Network string
	Whitelist []string // for whitelisting specific channels
	Permissions map[int][]string
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

	e := conf.Load(file.NewSource(file.WithPath("./default_config.json")),
						envvar.NewSource(),
						flag.NewSource(),
						file.NewSource(file.WithPath(filename)))

	if e != nil {
		panic(fmt.Sprintf("Error loading config %s", e))
	}

	var confObj = LogsConfig{}
	g := conf.Get()
	fmt.Println("Fetching now")
	g.Scan(&confObj)
	fmt.Println("scanned now")

	cachedFile = filename
	cache = confObj
	return confObj
}
