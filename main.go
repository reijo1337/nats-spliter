package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		logrus.Fatalf("getting config: %v", err)
	}
	logrus.SetLevel(cfg.LogLevel)

}
