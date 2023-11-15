package main

import (
	"context"
	"net"
	"os"
	"os/signal"

	"github.com/codeine-dev/go-gateway/pkg/configuration"
	quicclient "github.com/codeine-dev/go-gateway/pkg/quic_client"
	"github.com/sirupsen/logrus"
)

func main() {
	configService := configuration.NewMockAgentConfigurationService()

	client := quicclient.MakeQuicClient(&net.UDPAddr{Port: 1234}, configService)

	ctx := context.Background()

	err := client.Connect(ctx)

	if err != nil {
		logrus.Errorln("Failed to connect to server:", err)
	} else {
		logrus.Infoln("Connected to server just fine !!")
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	logrus.Infoln("Waiting for Ctrl+c")
	<-c
	logrus.Infoln(" Ctrl+c detected")

	err = client.Stop(ctx)
	if err != nil {
		logrus.Errorln("Failed to shutdown client:", err)
	}
}
