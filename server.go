package main

import (
	"go.uber.org/zap"
	"io"
	"net"
	"time"
)

type SerpentinisedServer struct {
	monitor  *RedisSentinelMonitor
	listener net.Listener
}

func NewSerpentinisedServer(addr string, monitor *RedisSentinelMonitor) (*SerpentinisedServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	var server SerpentinisedServer
	server.listener = listener
	server.monitor = monitor
	return &server, nil
}

func (server *SerpentinisedServer) Listen() (err error) {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return err
		}

		logger.Debug("received redis connection", zap.String("client", conn.RemoteAddr().String()))

		go func(conn net.Conn) {
			masterAddr := server.monitor.getCurrentMaster()
			masterConn, err := net.DialTimeout("tcp", masterAddr, time.Duration(*timeout)*time.Second)
			if err != nil {
				logger.Error("unable to connect to redis sentinel master",
					zap.String("client", conn.RemoteAddr().String()),
					zap.String("master", masterAddr))
				_ = conn.Close()
				return
			}

			logger.Debug("wiring up connection to master",
				zap.String("client", conn.RemoteAddr().String()),
				zap.String("master", masterAddr))
			go copyAndClose(masterConn, conn)
			go copyAndClose(conn, masterConn)
		}(conn)
	}
}

func copyAndClose(from, to net.Conn) {
	defer from.Close()
	if _, err := io.Copy(from, to); err != nil {
		if err == io.EOF {
			return
		}
		logger.Debug("i/o connection error when copying ",
			zap.String("from", from.RemoteAddr().String()),
			zap.String("to", to.RemoteAddr().String()))
	}
}
