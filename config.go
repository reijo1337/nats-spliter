package main

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type config struct {
	SRC struct {
		URL         string
		ConnectWait time.Duration
		PubAckWait  time.Duration
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
