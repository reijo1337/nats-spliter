package main

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

type stanConnect struct {
	natsConn *nats.Conn
	stanConn stan.Conn
	sub      stan.Subscription
}

func getStanConnect(natsURL, clusterID, clientID, subject string, handler stan.MsgHandler,
	natsOpts []nats.Option, stanOpts []stan.Option, subOpts []stan.SubscriptionOption) (*stanConnect, error) {
	natsConn, err := nats.Connect(natsURL, natsOpts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect to url %s: %v", natsURL, err)
	}

	stanOpts = append(stanOpts, stan.NatsConn(natsConn))
	stanConn, err := stan.Connect(clusterID, clientID, stanOpts...)
	if err != nil {
		natsConn.Close()
		return nil, fmt.Errorf("stan connect to cluster %s by client %s: %v", clusterID, clientID, err)
	}
	sub, err := stanConn.Subscribe(subject, handler, subOpts...)
	if err != nil {
		stanConn.Close()
		natsConn.Close()
		return nil, fmt.Errorf("subscribe on subject %s: %v", subject, err)
	}
	return &stanConnect{natsConn, stanConn, sub}, nil
}

func (sc *stanConnect) Close() error {
	if err := sc.sub.Unsubscribe(); err != nil {
		return fmt.Errorf("unsubscribe: %v", err)
	}
	if err := sc.stanConn.Close(); err != nil {
		return fmt.Errorf("close stan: %v", err)
	}
	sc.natsConn.Close()
	return nil
}
