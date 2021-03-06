package main

import (
	"flag"
	"math/rand"
	"net"
	"time"

	"github.com/nm-morais/go-babel/pkg"
	"github.com/nm-morais/go-babel/pkg/peer"
)

func main() {
	rand.Seed(time.Now().Unix() + rand.Int63())
	minProtosPort := 7000
	maxProtosPort := 8000
	minAnalyticsPort := 8000
	maxAnalyticsPort := 9000

	var protosPortVar int
	var analyticsPortVar int
	var randProtosPort bool
	var randAnalyticsPort bool

	flag.IntVar(&protosPortVar, "analytics", 1201, "analytics")
	flag.IntVar(&analyticsPortVar, "protos", 1200, "protos")
	flag.BoolVar(&randProtosPort, "rprotos", false, "port")
	flag.BoolVar(&randAnalyticsPort, "ranalytics", false, "port")
	flag.Parse()

	if randProtosPort {
		protosPortVar = rand.Intn(maxProtosPort-minProtosPort) + minProtosPort
	}

	if randAnalyticsPort {
		analyticsPortVar = rand.Intn(maxAnalyticsPort-minAnalyticsPort) + minAnalyticsPort
	}

	config := pkg.Config{
		SmConf: pkg.StreamManagerConf{
			DialTimeout: 3 * time.Second,
		},
		LogFolder:        "/Users/nunomorais/go/src/github.com/nm-morais/go-babel/logs/",
		HandshakeTimeout: 1 * time.Second,
		Peer:             peer.NewPeer(net.IPv4(0, 0, 0, 0), uint16(protosPortVar), uint16(analyticsPortVar)),
	}

	p := pkg.NewProtoManager(config)
	contactNode := peer.NewPeer(net.IPv4(0, 0, 0, 0), uint16(1200), uint16(1200))
	p.RegisterListenAddr(&net.TCPAddr{IP: config.Peer.IP(), Port: int(config.Peer.ProtosPort())})
	p.RegisterProtocol(NewPingPongProtocol(contactNode, p))
	p.Start()
}
