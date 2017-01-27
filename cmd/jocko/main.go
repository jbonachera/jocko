package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tj/go-gracefully"
	"github.com/travisjeffery/jocko/broker"
	"github.com/travisjeffery/jocko/server"
	"github.com/travisjeffery/simplelog"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logDir      = kingpin.Flag("logdir", "A comma separated list of directories under which to store log files").Default("/tmp/jocko").String()
	tcpAddr     = kingpin.Flag("tcpaddr", "HTTP Address to listen on").String()
	raftDir     = kingpin.Flag("raftdir", "Directory for Raft to store data").String()
	raftAddr    = kingpin.Flag("raftaddr", "Address for Raft to bind on").String()
	raftPort    = kingpin.Flag("raftport", "Port for Raft to bind on").Int()
	serfPort    = kingpin.Flag("serfport", "Port for Serf to bind on").Default("7946").Int()
	serfMembers = kingpin.Flag("serfmembers", "List of existing Serf members").Strings()
	brokerID    = kingpin.Flag("id", "Broker ID").Int32()
	debugLogs   = kingpin.Flag("debug", "Enable debug logs").Default("false").Bool()
)

func main() {
	kingpin.Parse()

	logLevel := simplelog.INFO
	if *debugLogs {
		logLevel = simplelog.DEBUG
	}
	logger := simplelog.New(os.Stdout, logLevel, "jocko")

	addr, err := net.ResolveTCPAddr("tcp", *tcpAddr)
	if err != nil {
		panic(err)
	}

	store, err := broker.New(*brokerID,
		broker.DataDir(*logDir),
		broker.LogDir(*logDir),
		broker.Logger(logger),
		broker.BindAddr(*raftAddr),
		broker.RaftPort(*raftPort),
		broker.SerfPort(*serfPort),
		broker.SerfMembers(*serfMembers),
		broker.Port(addr.Port),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with new broker: %s\n", err)
		os.Exit(1)
	}
	srv := server.New(*tcpAddr, store, logger)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %s\n", err)
		os.Exit(1)
	}

	defer srv.Close()

	gracefully.Timeout = 10 * time.Second
	gracefully.Shutdown()

	if err := store.Shutdown(); err != nil {
		panic(err)
	}
}
