package main

import (
	"context"
	"net"
	"os"
	"os/signal"

	"github.com/codeine-dev/go-gateway/pkg/configuration"
	"github.com/codeine-dev/go-gateway/pkg/gateway"
	"github.com/codeine-dev/go-gateway/pkg/ingress"
	quicserver "github.com/codeine-dev/go-gateway/pkg/quic_server"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.Println("Hello world!")

	ingress := ingress.NewTCPIngressServer(&net.TCPAddr{Port: 1234})

	configService := configuration.NewMockConfigurationService()

	handler := gateway.MakeAgentHandler(configService)

	server := quicserver.MakeQuickServer(&net.UDPAddr{Port: 1234})

	ctx := context.Background()

	err := server.Start(ctx, handler)

	if err != nil {
		log.Errorln("Failed to start server", err)
	}

	log.Infoln("QUIC Server status: ", server.Status(ctx))

	ingress.Start(ctx, handler)

	if err != nil {
		log.Errorln("Failed to start ingress server", err)
	}

	log.Infoln("Ingress Server status: ", ingress.Status(ctx))

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	<-c

	log.Infoln("SIGINT detected, shutting down")

	log.Infoln("Shutting down handler")
	handler.CloseAll()

	log.Infoln("Shutting down ingress")
	ingress.Stop(ctx)

	log.Infoln("Shutting down server")
	server.Stop(ctx)
}
