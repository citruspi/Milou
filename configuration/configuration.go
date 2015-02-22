package configuration

import (
	"flag"
	"log"

	"github.com/FogCreek/mini"
)

type hostConfiguration struct {
	TTL int64
}

type webConfiguration struct {
	Address string
	Status  bool
}

type travisConfiguration struct {
	Authenticate bool
	Token        string
}

var (
	Host   hostConfiguration
	Web    webConfiguration
	Travis travisConfiguration
)

func Process() {
	path := flag.String("config", "/etc/iago.ini", "Configuration file path")
	flag.Parse()

	config, err := mini.LoadConfiguration(*path)

	if err != nil {
		log.Fatal(err)
	}

	Host.TTL = config.IntegerFromSection("Host", "TTL", 30)

	Web.Address = config.StringFromSection("Web", "Address", "127.0.01:8000")
	Web.Status = config.BooleanFromSection("Web", "Status", true)

	Travis.Authenticate = config.BooleanFromSection("Travis", "Authenticate", false)
	Travis.Token = config.StringFromSection("Travis", "Token", "")
}
