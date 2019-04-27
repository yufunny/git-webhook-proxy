package main

import (
	"encoding/json"
	"git-proxy/client/config"
	"git-proxy/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
)

var (
	configPath   string
	systemConfig *config.SystemConfig
	g            errgroup.Group
)

func main() {
	{
		app := cli.NewApp()
		app.Name = "git proxy client"
		app.Usage = "client for git-proxy"
		app.Version = "0.1.0"
		app.Authors = []cli.Author{
			{
				Name:  "yufu",
				Email: "mxy@yufu.fun",
			},
		}
		app.Flags = []cli.Flag{
			cli.StringFlag{
				Name:        "config, c",
				Usage:       "config path",
				Value:       "client/config.yaml",
				Destination: &configPath,
			},
		}
		app.Action = run

		log.Infof("[MAIN]Run start")
		err := app.Run(os.Args)
		if nil != err {
			log.Errorf("[MAIN]Run error; " + err.Error())
		}
	}
}

func run(_ *cli.Context) {
	var err error
	systemConfig, err = config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("[MAIN]parse config error:%s", err.Error())
	}

	if systemConfig.Mode == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", systemConfig.Server)
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
		os.Exit(1)
	}

	log.Info("connect success")
	go subscribe(conn, systemConfig.Subscribes)
	readCon(conn)

}

func readCon(conn net.Conn) {
	defer conn.Close()
	tmpBuffer := make([]byte, 0)

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Errorf("connection error: %s, %s", conn.RemoteAddr().String(), err.Error())
			conn.Close()
			return
		}
		protocolId, msg := protocol.DePack(append(tmpBuffer, buffer[:n]...))
		if protocolId < 0 {
			tmpBuffer = make([]byte, 0)
		} else if protocolId == 0 {
			tmpBuffer = append(tmpBuffer, buffer[:n]...)
		} else {
			log.Infof("got message from server: %d,%s", protocolId, msg)
		}
	}
}

func subscribe(conn net.Conn, subscribes []config.SubscribeConfig) {
	urls := make([]string, 0)
	for _, subscribe := range subscribes {
		urls = append(urls, subscribe.Url)
	}
	msg, _ := json.Marshal(urls)
	conn.Write(protocol.EnPack(101, msg))
	log.Infof("subscribes:%s", msg)
}
