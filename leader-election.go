package main

import (
	"time"

	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

const LockKey = "consul-alerts/leader"

type LeaderElection struct {
	lock           *consulapi.Lock
	cleanupChannel chan struct{}
	stopChannel    chan struct{}
	leader         bool
}

func (l *LeaderElection) start() {
	clean := false
	for !clean {
		select {
		case <-l.cleanupChannel:
			clean = true
		default:
			log.Infoln("Running for leader election...")
			intChan, err := l.lock.Lock(l.stopChannel)
			if err != nil {
				log.Infoln("Failed to acquire lock")
				time.Sleep(10000 * time.Millisecond)
			} else {
				log.Infoln("Now acting as leader.")
				l.leader = true
			}
			if intChan != nil {
				<-intChan
				l.leader = false
				log.Infoln("Lost leadership.")
				l.lock.Unlock()
				l.lock.Destroy()
			}
		}
	}
}

func (l *LeaderElection) stop() {
	log.Infoln("cleaning up")
	l.cleanupChannel <- struct{}{}
	l.stopChannel <- struct{}{}
	l.lock.Unlock()
	l.lock.Destroy()
	l.leader = false
	log.Infoln("cleanup done")
}

func startLeaderElection(addr, dc, scheme, acl string) *LeaderElection {
	config := consulapi.DefaultConfig()
	config.Address = addr
	config.Datacenter = dc
	config.Token = acl
	config.Scheme = scheme
	client, _ := consulapi.NewClient(config)
	lock, _ := client.LockKey(LockKey)

	leader := &LeaderElection{
		lock:           lock,
		cleanupChannel: make(chan struct{}, 1),
		stopChannel:    make(chan struct{}),
	}

	go leader.start()

	return leader
}

func hasLeader() bool {
	return consulClient.CheckKeyExists(LockKey)
}
