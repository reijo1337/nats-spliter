package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

var (
	errNoSeparator  = errors.New("no separator field in input")
	errInvalidJSON  = errors.New("invalid json, must start with '{' or '['")
	errNoMsgHandler = errors.New("empty subscribition msg handler")
)

type (
	// STAN connects holder
	stanConnect struct {
		natsConn *nats.Conn
		stanConn stan.Conn
		sub      stan.Subscription
		subject  string // optional, required for dsts because they have no subscribitions
	}

	// NATS connection config
	natsConfig struct {
		url                 string
		token               string
		pingInterval        time.Duration
		maxPingsOutstanding int
		reconnectWait       time.Duration
		maxReconnects       int
	}

	// STAN connection config
	stanConfig struct {
		clusterID, clientID string
		natsConn            *nats.Conn
		connectWait         time.Duration
		pings               [2]int
		conLostHandler      stan.ConnectionLostHandler
	}

	// Subscribition config
	subConfig struct {
		subject               string
		group                 string
		handler               stan.MsgHandler
		setManualAckMode      bool
		maxInflight           int
		ackWait               time.Duration
		durableName           string
		startTime             string
		startAtTimeDelta      time.Duration
		startAtSequence       uint64
		startWithLastReceived bool
	}

	// destination configs
	dstConfig struct {
		SeparatorValue           string `json:"separator_value"`
		NatsURL                  string `json:"nats_url"`
		NatsToken                string `json:"nats_token"`
		NatsPingInterval         string `json:"nats_ping_interval"`
		NatsMaxPingsOutstangding int    `json:"nats_max_pings_outstanding"`
		NatsReconnectWait        string `json:"nats_reconnect_wait"`
		NatsMaxReconnects        int    `json:"nats_max_reconnects"`
		StanCluster              string `json:"stan_cluster"`
		StanConnectWait          string `json:"stan_connect_wait"`
		StanPings                [2]int `json:"stan_pings"`
		StanSubject              string `json:"stan_subject"`
	}
)

func (nc natsConfig) connect() (*nats.Conn, error) {
	opts := make([]nats.Option, 0, 5)
	if nc.token != "" {
		opts = append(opts, nats.Token(nc.token))
	}
	if nc.pingInterval > 0 {
		opts = append(opts, nats.PingInterval(nc.pingInterval))
	}
	if nc.maxPingsOutstanding != 0 {
		opts = append(opts, nats.MaxPingsOutstanding(nc.maxPingsOutstanding))
	}
	if nc.reconnectWait > 0 {
		opts = append(opts, nats.ReconnectWait(nc.reconnectWait))
	}
	if nc.maxReconnects != 0 {
		opts = append(opts, nats.MaxReconnects(nc.maxReconnects))
	}
	return nats.Connect(nc.url, opts...)
}

func (sc stanConfig) connect(natsConn *nats.Conn) (stan.Conn, error) {
	opts := make([]stan.Option, 0, 5)
	opts = append(opts, stan.NatsConn(natsConn))
	if sc.connectWait > 0 {
		opts = append(opts, stan.ConnectWait(sc.connectWait))
	}
	if sc.pings != [2]int{} {
		opts = append(opts, stan.Pings(sc.pings[0], sc.pings[1]))
	}

	opts = append(opts, stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
		log.Fatalf("stan connection lost, reason: %v\n", reason)
	}))
	return stan.Connect(sc.clusterID, sc.clientID)
}

func (sc *subConfig) connect(stanConn stan.Conn) (stan.Subscription, error) {
	if sc == nil {
		return nil, nil
	}
	if sc.handler == nil {
		return nil, errNoMsgHandler
	}
	opts := make([]stan.SubscriptionOption, 0, 8)
	if sc.setManualAckMode {
		opts = append(opts, stan.SetManualAckMode())
	}
	if sc.maxInflight > 0 {
		opts = append(opts, stan.MaxInflight(sc.maxInflight))
	}
	if sc.ackWait > 0 {
		opts = append(opts, stan.AckWait(sc.ackWait))
	}
	if sc.durableName != "" {
		opts = append(opts, stan.DurableName(sc.durableName))
	}
	if sc.startTime != "" {
		startTime, err := time.Parse("2006-01-02T15:04:05", sc.startTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start time for subscribition [%s]: %v", sc.startTime, err)
		}
		opts = append(opts, stan.StartAtTime(startTime))
	}
	if sc.startAtTimeDelta > 0 {
		opts = append(opts, stan.StartAtTimeDelta(sc.startAtTimeDelta))
	}
	if sc.startAtSequence > 0 {
		opts = append(opts, stan.StartAtSequence(sc.startAtSequence))
	}
	if sc.startWithLastReceived {
		opts = append(opts, stan.StartWithLastReceived())
	}
	return stanConn.QueueSubscribe(sc.subject, sc.group, sc.handler, opts...)
}

func getStanConnect(natsCfg natsConfig, stanCfg stanConfig, subCfg *subConfig, subject string) (*stanConnect, error) {
	natsConn, err := natsCfg.connect()
	if err != nil {
		return nil, fmt.Errorf("nats %s connect: %v", natsCfg.url, err)
	}

	stanConn, err := stanCfg.connect(natsConn)
	if err != nil {
		natsConn.Close()
		return nil, fmt.Errorf("stan connect: %v", err)
	}

	sub, err := subCfg.connect(stanConn)
	if err != nil {
		stanConn.Close()
		natsConn.Close()
		return nil, fmt.Errorf("subscribe on subject %s: %v", subCfg.subject, err)
	}
	return &stanConnect{natsConn, stanConn, sub, subject}, nil
}

func (sc *stanConnect) Close() error {
	if sc.sub != nil {
		if err := sc.sub.Unsubscribe(); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	}
	if err := sc.stanConn.Close(); err != nil {
		return fmt.Errorf("close stan: %v", err)
	}
	sc.natsConn.Close()
	return nil
}
