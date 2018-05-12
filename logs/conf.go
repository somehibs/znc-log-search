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
	User      string
	Password  string
	Db        string
}

type IndexerConfig struct {
	Daily             bool
	DefaultPermission int
	Permissions       map[int][]string
	Whitelist         []string // for whitelisting specific channels
}

type LogsConfig struct {
	ApiUrl     string
	Prometheus bool
	Network    string
	Queues     map[string]int
	Sphinx     SphinxConfig
	Arango     ArangoConfig
	Indexer    IndexerConfig
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
	fmt.Println("Loading config...")
	g := conf.Get()
	g.Scan(&confObj)
	//panic(fmt.Sprintf("%+v\n",confObj))
	fmt.Println("Config loaded!")

	cachedFile = filename
	cache = confObj
	return confObj
}
