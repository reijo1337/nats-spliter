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
			Cluster     string        `requeried:"true"`
			ConnectWait time.Duration `default:"60s" envconfig:"SRC_STAN_CONNECT_WAIT"`
			Pings       [2]int        `default:"3,20"`
		}
		Sub struct {
			Subject               string `requeried:"true"`
			Group                 string
			SetManualAckMode      bool `default:"true" envconfig:"SRC_SUB_SET_MANUAL_ACK_MODE"`
			MaxInflight           int  `default:"1" envconfig:"SRC_SUB_MAX_INFLIGHT"`
			ackWait               time.Duration
			DurableName           string
			StartTime             string
			StartAtTimeDelta      time.Duration
			StartAtSequence       uint64
			StartWithLastReceived bool
		}
	}
	Delay struct {
		Loop    time.Duration `default:"3s"`
		IfError time.Duration `default:"3s"`
	}
	StanClient       string       `envconfig:"STAN_CLIENT" required:"true"`
	SeparatorName    string       `envconfig:"SEPARATOR_NAME" required:"true"`
	SendToAllOnError bool         `default:"true" envconfig:"SEND_TO_ALL"`
	DstFileLoc       string       `envconfig:"DST_FILE_LOC"`
	LogLevel         logrus.Level `default:"info" envconfig:"LOG_LEVEL"`
}

func parseConfig() (out config, err error) {
	if err := envconfig.Process("", &out); err != nil {
		return out, fmt.Errorf("process envconfig: %v", err)
	}
	return out, nil
}
