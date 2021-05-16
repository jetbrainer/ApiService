package main

import (
	"context"
	"log"

	"github.com/JetBrainer/ApiService/src/config"
	"github.com/JetBrainer/ApiService/src/manager"
	"github.com/JetBrainer/ApiService/src/server/http"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Printf("[ERROR] Init config %v", err)
		return
	}

	requestMan := manager.NewRequester()

	opts := []http.Option{
		http.WithContext(ctx),
		http.WithConfig(cfg.Server),
		http.WithRequestManager(requestMan),
	}

	srv := http.NewHTTPServer(opts...)

	srv.Run(cancel)

}



