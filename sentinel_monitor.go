package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"strings"
	"sync/atomic"
)

type RedisSentinelMonitor struct {
	sentinelConn   redis.Conn
	sentinelMaster string
	currentMaster  atomic.Value // string
}

func NewRedisSentinelMonitor(sentinelAddress, sentinelMaster string) (*RedisSentinelMonitor, error) {
	var monitor RedisSentinelMonitor
	sentinelConn, err := redis.Dial("tcp", sentinelAddress)
	if err != nil {
		return nil, err
	}

	monitor.sentinelConn = sentinelConn
	monitor.sentinelMaster = sentinelMaster
	masterAddr, err := monitor.getInitialMaster()
	if err != nil {
		return nil, err
	}

	// Sentinel is good, let's initialize it
	monitor.setCurrentMaster(masterAddr)
	return &monitor, nil
}

func (monitor *RedisSentinelMonitor) getInitialMaster() (string, error) {
	result, err := redis.Strings(monitor.sentinelConn.Do("SENTINEL", "get-master-addr-by-name", monitor.sentinelMaster))
	if err != nil {
		return "", nil
	}

	address := fmt.Sprintf("%s:%s", result[0], result[1])
	return address, nil
}

func (monitor *RedisSentinelMonitor) getCurrentMaster() string {
	return monitor.currentMaster.Load().(string)
}

func (monitor *RedisSentinelMonitor) setCurrentMaster(addr string) {
	monitor.currentMaster.Store(addr)
}

func (monitor *RedisSentinelMonitor) monitorSentinelChanges() {
	psc := redis.PubSubConn{Conn: monitor.sentinelConn}
	if err := psc.Subscribe("switch-master"); err != nil {
		logger.Error("unable to subscribe to switch-master messages from sentinel", zap.Error(err))
		return
	}

	for {
		messageRaw := psc.Receive()
		message, wasMessage := messageRaw.(redis.Message)
		if !wasMessage {
			// not interested in subscriptions
			continue
		}

		// <master name> <oldip> <oldport> <newip> <newport>
		switchMessageData := strings.Split(string(message.Data), " ")
		if len(switchMessageData) != 5 {
			logger.Debug("ignoring invalid switch-master message", zap.Strings("message", switchMessageData))
			continue
		}

		relevantMaster := switchMessageData[0]
		if relevantMaster != monitor.sentinelMaster {
			logger.Debug("ignoring irrelevant switch-master message", zap.String("master", relevantMaster))
			continue
		}
		newMasterHost := switchMessageData[3]
		newMasterPort := switchMessageData[4]
		newMasterAddr := fmt.Sprintf("%s:%s", newMasterHost, newMasterPort)
		logger.Info("old master came down, new master elected - switching over",
			zap.String("old-master", monitor.sentinelMaster),
			zap.String("new-master", newMasterAddr))
		monitor.setCurrentMaster(newMasterAddr)
	}
}
