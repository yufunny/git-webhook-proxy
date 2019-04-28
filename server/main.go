package main

import (
	"git-webhook-proxy/server/config"
	"git-webhook-proxy/server/handler"
	"git-webhook-proxy/server/server"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"time"
)

var (
	configPath   string
	systemConfig *config.SystemConfig
	g            errgroup.Group
)

func main() {
	{
		app := cli.NewApp()
		app.Name = "git proxy"
		app.Usage = "proxy git web hook"
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
				Value:       "server/config.yaml",
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

	mux := http.NewServeMux()
	mux.HandleFunc("/notify", handler.NotifyHandler)
	//api接口
	httpServer := &http.Server{
		Addr:         systemConfig.Listen.Http,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		return httpServer.ListenAndServe()
	})

	log.Debug("Starting the socket server ...")
	g.Go(func() error {
		server.Start(systemConfig.Listen.Socket)
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
