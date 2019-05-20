package main

import (
	"encoding/json"
	"fmt"
	"git-webhook-proxy/client/config"
	"git-webhook-proxy/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var (
	configPath   string
	systemConfig *config.SystemConfig
	g            errgroup.Group
	beforePull = "git reset --hard && git checkout "
	subscribeMap = make(map[string]*config.SubscribeConfig, 0)
)

func main() {
	{
		app := cli.NewApp()
		app.Name = "git proxy client"
		app.Usage = "client for git-webhook-proxy"
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
	subSocket(conn, systemConfig)
	if systemConfig.HttpListen != "" {
		go subHttp()
	}
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
			log.Debugf("got message from server: %d,%s", protocolId, msg)
			go deal(string(msg))
		}
	}
}

func subSocket(conn net.Conn, config *config.SystemConfig) {
	urls := make([]string, 0)
	for _, subscribe := range config.Subscribes {
		subscribeMap[subscribe.Url] = &subscribe
		urls = append(urls, subscribe.Url)
	}
	msg, _ := json.Marshal(urls)
	conn.Write(protocol.EnPack(101, msg))
	log.Infof("subscribes:%s", msg)
}



func subHttp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/notify", NotifyHandler)
	//api接口
	httpServer := &http.Server{
		Addr:         systemConfig.HttpListen,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	httpServer.ListenAndServe()
}


func NotifyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	log.Debugf("notify :%s", body)
	go deal(string(body))
	w.Write([]byte("ok"))
}

func deal(body string) {
	res := gjson.Parse(body)
	url := res.Get("repository.ssh_url").String()

	var subscribed *config.SubscribeConfig
	var ok bool
	if subscribed, ok = subscribeMap[url]; !ok {
		log.Infof("un subscribed url:%s", url)
		return
	}
	branch := res.Get("ref").String()
	if branch != "refs/heads/" + subscribed.Branch {
		log.Infof("branch not match:%s|%s|%s", url, branch, subscribed.Branch)
		return
	}
	log.Infof("notify url:%s", url)

	command := fmt.Sprintf("%s && git pull", beforePull)
	log.Infof("cmd:%s", command)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = subscribed.Directory
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	log.Infof("cmd output:%s", out)
}
