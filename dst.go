package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func parseDsts(dstFileLoc, clientID string) (map[string]*stanConnect, error) {
	dstConfigFile, err := os.Open(dstFileLoc)
	if err != nil {
		return nil, fmt.Errorf("open dst config file: %v", err)
	}
	byteConfig, err := ioutil.ReadAll(dstConfigFile)
	if err != nil {
		return nil, fmt.Errorf("read dst config file: %v", err)
	}
	dsts := make([]dstConfig, 0)
	if err := json.Unmarshal(byteConfig, &dsts); err != nil {
		return nil, fmt.Errorf("parse dst config file: %v", err)
	}
	out := make(map[string]*stanConnect, len(dsts))
	for i := range dsts {
		natsCfg, stanCfg, err := dsts[i].getConnectConfigs(clientID)
		if err != nil {
			return nil, fmt.Errorf("get nats/stan config from %d dst config: %v", i, err)
		}
		con, err := getStanConnect(natsCfg, stanCfg, nil, dsts[i].StanSubject)
		if err != nil {
			return nil, fmt.Errorf("connect to %d dst: %v", i, err)
		}
		out[dsts[i].SeparatorValue] = con
	}
	return out, nil
}

func (dc *dstConfig) getConnectConfigs(clientID string) (nc natsConfig, sc stanConfig, err error) {
	pingInterval, err := time.ParseDuration(dc.NatsPingInterval)
	if err != nil {
		return nc, sc, fmt.Errorf("parse dst nats ping interval: %v", err)
	}
	reconnectWait, err := time.ParseDuration(dc.NatsReconnectWait)
	if err != nil {
		return nc, sc, fmt.Errorf("parse dst nats reconect wait: %v", err)
	}
	connectWait, err := time.ParseDuration(dc.StanConnectWait)
	if err != nil {
		return nc, sc, fmt.Errorf("parse dst stan connect wait: %v", err)
	}

	return natsConfig{
			url:                 dc.NatsURL,
			token:               dc.NatsToken,
			pingInterval:        pingInterval,
			maxPingsOutstanding: dc.NatsMaxPingsOutstangding,
			reconnectWait:       reconnectWait,
			maxReconnects:       dc.NatsMaxReconnects,
		},
		stanConfig{
			clusterID:   dc.StanCluster,
			clientID:    clientID,
			connectWait: connectWait,
			pings:       dc.StanPings,
		}, nil
}

type dctConfigJSON dstConfig

func (dc *dstConfig) UnmarshalJSON(data []byte) error {
	// default values
	tmp := &dctConfigJSON{
		NatsPingInterval:         "3s",
		NatsMaxPingsOutstangding: 20,
		NatsReconnectWait:        "5s",
		NatsMaxReconnects:        20,
		StanConnectWait:          "10s",
		StanPings:                [2]int{3, 20},
	}
	if err := json.Unmarshal(data, tmp); err != nil {
		return err
	}
	*dc = dstConfig(*tmp)
	return nil
}

// Proccess one message, try to get separator
func getSeparator(msg []byte, separatorName string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return "", fmt.Errorf("unmarshal message: %v", err)
	}
	sep, ok := data[separatorName].(string)
	if !ok {
		return "", errNoSeparator
	}
	return sep, nil
}

// Proccess one array message, try to get separator
func getSeparatorArray(msg []byte, separatorName string) (string, error) {
	var data []map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return "", fmt.Errorf("unmarshal message: %v", err)
	}
	sep, ok := data[0][separatorName].(string)
	if !ok {
		return "", errNoSeparator
	}
	return sep, nil
}
