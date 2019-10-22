package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	src, err := getStanConnect(
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
		"",
	)
	if err != nil {
		logrus.Fatalf("create src stan: %v", err)
	}
	logrus.Info("connected to src")
	// Destination NATS connects mapped on separate value
	dstMap, err := parseDsts(cfg.DstFileLoc, cfg.StanClient)
	if err != nil {
		logrus.Fatalf("get dst connects: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		delay := cfg.Delay.Loop
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			msg := <-msgs
			// choose the send way
			var separator string
			switch msg.Data[0] {
			case 123: // one json
				separator, err = getSeparator(msg.Data, cfg.SeparatorName)
			case 91: // array of json
				separator, err = getSeparatorArray(msg.Data, cfg.SeparatorName)
			default:
				separator, err = "", errInvalidJSON
			}
			if err != nil {
				logrus.Errorf("get separator from incoming message: %v", err)
			}

			// send to dst
			dstConn, ok := dstMap[separator]
			if ok {
				if err := dstConn.stanConn.Publish(dstConn.subject, msg.Data); err != nil {
					logrus.Errorf("puplish message to one dst: %v", err)
					delay = cfg.Delay.IfError
					continue
				}
			} else if cfg.SendToAllOnError {
				hasError := false
				for _, dst := range dstMap {
					if err := dst.stanConn.Publish(dst.subject, msg.Data); err != nil {
						logrus.Errorf("puplish message to array dsts: %v", err)
						hasError = true
						break
					}
				}
				if hasError {
					delay = cfg.Delay.IfError
					continue
				}
			}
			msg.Ack()
			delay = cfg.Delay.Loop
		}
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	logrus.Info("Started")

	<-osSignals
	logrus.Info("starting shutdown")
	cancel()
	for sep, dst := range dstMap {
		if err := dst.Close(); err != nil {
			logrus.Errorf("close %s dst: %v", sep, err)
		}
	}
	if err := src.Close(); err != nil {
		logrus.Errorf("close src: %v", err)
	}

	logrus.Info("shutdown")
}
