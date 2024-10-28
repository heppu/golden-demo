package main

import (
	"flag"
	"log"
	"log/slog"
	"syscall"

	"github.com/elisasre/go-common/v2/service"
	"github.com/elisasre/go-common/v2/service/module/httpserver"
	"github.com/elisasre/go-common/v2/service/module/siglistener"
	"github.com/heppu/golden-demo/api"
	"github.com/heppu/golden-demo/api/handler"
	"github.com/heppu/golden-demo/store"
)

func main() {
	var dsn, addr string
	flag.StringVar(&dsn, "dsn", "postgres://demo:demo@127.0.0.1:5432/demo?sslmode=disable", "DB DSN")
	flag.StringVar(&addr, "addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	slog.Info("Starting golden-demo with config", slog.String("addr", addr))
	slog.Info("openapi ui available at", slog.String("addr", "http://"+addr+"/openapi"))
	s := store.New(dsn)
	if err := service.Run(service.Modules{
		s,
		httpserver.New(
			httpserver.WithAddr(addr),
			api.WithHandler(handler.New(s)),
		),
		siglistener.New(syscall.SIGINT, syscall.SIGTERM),
	}); err != nil {
		log.Fatal(err)
	}
}
