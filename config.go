package main

import (
	"crypto/md5"
	"fmt"
	"os"
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
			ConnectWait time.Duration `default:"10s" envconfig:"SRC_STAN_CONNECT_WAIT"`
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
	StanClient       string `envconfig:"STAN_CLIENT" required:"true"`
	SeparatorName    string `envconfig:"SEPARATOR_NAME" required:"true"`
	SendToAllOnError bool   `default:"true" envconfig:"SEND_TO_ALL"`
	DstFileLoc       string
	LogLevel         logrus.Level `default:"info" envconfig:"LOG_LEVEL"`
}

func parseConfig() (out config, err error) {
	if err := envconfig.Process("", &out); err != nil {
		return out, fmt.Errorf("process envconfig: %v", err)
	}
	// if out.StanClient == "" {
	// 	cid, err := getClient()
	// 	if err != nil {
	// 		return out, fmt.Errorf("generate client id: %v", err)
	// 	}
	// 	out.SRC.Stan.Client = cid
	// }
	return out, nil
}

func getClient() (string, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}
	hash := md5.Sum([]byte(hostName))
	return fmt.Sprintf("%x-%x-%x-%x-%x", hash[:4], hash[4:6], hash[6:8], hash[8:10], hash[10:]), nil

}
