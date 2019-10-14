package main

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type config struct {
	SRC struct {
		Nats struct {
			URL                 string `requeried:"true"`
			Token               string
			PingInterval        time.Duration `default:"3s" envconfig:"SRC_NATS_PING_INTERVAL"`
			MaxPingsOutstanding int           `default:"20" envconfig:"SRC_NATS_MAX_PINGS_OUTSTANDING"`
			ReconnectWait       time.Duration `default:"5s" envconfig:"SRC_NATS_RECONNECT_WAIT"`
			MaxReconnects       int           `default:"20"`
		}
		Stan struct {
			Cluster     string `requeried:"true"`
			Client      string
			ConnectWait time.Duration `default:"10s" envconfig:"SRC_STAN_CONNECT_WAIT"`
			Pings       [2]int        `default:"3,20"`
		}
		Sub struct {
			Subject               string `requeried:"true"`
			Group                 string
			SetManualAckMode      bool `default:"true" envconfig:"SRC_SUB_SET_MANUAL_ACK_MODE"`
			MaxInflight           int  `default:"1" envconfig:"SRC_SUB_MAX_INFLIGHT"`
			ackWait               time.Duration
			durableName           string
			startTime             string
			startAtTimeDelta      time.Duration
			startAtSequence       uint64
			startWithLastReceived bool
		}
	}
	DstFileLoc string
	LogLevel   logrus.Level `default:"info" envconfig:"LOG_LEVEL"`
}

func parseConfig() (out config, err error) {
	if err := envconfig.Process("", &out); err != nil {
		return out, fmt.Errorf("process envconfig: %v", err)
	}
	return out, nil
}
