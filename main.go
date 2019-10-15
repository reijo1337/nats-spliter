package main

import (
	stan "github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		logrus.Fatalf("getting config: %v", err)
	}
	logrus.SetLevel(cfg.LogLevel)
	// chan for messages from source
	msgs := make(chan *stan.Msg, 1)

	// Source NATS connection
	srv, err := getStanConnect(
		natsConfig{
			url:                 cfg.SRC.Nats.URL,
			token:               cfg.SRC.Nats.Token,
			pingInterval:        cfg.SRC.Nats.PingInterval,
			maxPingsOutstanding: cfg.SRC.Nats.MaxPingsOutstanding,
			reconnectWait:       cfg.SRC.Nats.ReconnectWait,
			maxReconnects:       cfg.SRC.Nats.MaxReconnects,
		},
		stanConfig{
			clusterID:   cfg.SRC.Stan.Cluster,
			clientID:    cfg.StanClient,
			connectWait: cfg.SRC.Stan.ConnectWait,
			pings:       cfg.SRC.Stan.Pings,
		},
		&subConfig{
			subject:               cfg.SRC.Sub.Subject,
			group:                 cfg.SRC.Sub.Group,
			setManualAckMode:      cfg.SRC.Sub.SetManualAckMode,
			maxInflight:           cfg.SRC.Sub.MaxInflight,
			ackWait:               cfg.SRC.Sub.ackWait,
			durableName:           cfg.SRC.Sub.DurableName,
			startTime:             cfg.SRC.Sub.StartTime,
			startAtTimeDelta:      cfg.SRC.Sub.StartAtTimeDelta,
			startAtSequence:       cfg.SRC.Sub.StartAtSequence,
			startWithLastReceived: cfg.SRC.Sub.StartWithLastReceived,
			handler:               func(msg *stan.Msg) { msgs <- msg },
		},
	)
	if err != nil {
		logrus.Fatalf("create src stan: %v", err)
	}
	dstMap, err := parseDsts(cfg.DstFileLoc, cfg.StanClient)
	if err != nil {
		logrus.Fatalf("get dst connects: %v", err)
	}
}
