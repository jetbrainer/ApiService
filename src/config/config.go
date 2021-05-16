package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	*Server
}

type Server struct {
	HTTPAddr string `json:"http_addr"`
	*TLS
}

type TLS struct {
	KeyFile  string `json:"key_file"`
	CertFile string `json:"cert_file"`
}

func InitConfig() (*Configuration, error) {
	var cfg Configuration
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
