package server

import (
	"git-proxy/utils"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type SocketServer struct {
	listener    net.Listener
	connections map[string]*Connection
	urls        map[string][]string
	tick        *time.Ticker //  定时器
}

func GetServer() *SocketServer {
	return server
}

var server *SocketServer

func Start(listen string) {
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		panic("[server]Error listening:" + err.Error())
	}
	server = &SocketServer{
		listener:    listener,
		connections: make(map[string]*Connection),
		urls:        make(map[string][]string, 0),
		tick:        time.NewTicker(time.Minute),
	}
	server.Start()
}

func (s *SocketServer) Start() {
	eventChan := make(chan *event, 1000)
	go s.conEvents(eventChan)
	go s.clear()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Errorf("[server] Error accepting:%s", err.Error())
			continue
		}
		key, err := utils.GetGuid()
		if nil != err {
			conn.Close()
			log.Errorf("[server] connection get guid failed")
			continue
		}
		con := &Connection{
			conn,
			eventChan,
			key,
			true,
			time.Now().Add(5 * time.Minute),
			0,
			&sync.Mutex{},
		}
		s.connections[key] = con
		go con.Listen()
	}
}

func (s *SocketServer) Dispatch(url, body string) {
	conns := s.getCons(url)

	for _, conn := range conns {
		conn.Send(body)
	}
}

func (s *SocketServer) getCons(url string) []*Connection {
	conns := make([]*Connection, 0)
	if keys, ok := s.urls[url]; ok {
		for _, key := range keys {
			if conn, ok := s.connections[key]; ok {
				conns = append(conns, conn)
			}
		}
	}
	return conns
}

func (s *SocketServer) conEvents(ev chan *event) {
	for {
		select {
		case e := <-ev:
			switch e.trigger {
			case "subscribe":
				exist := false
				for _, key := range s.urls[e.msg] {
					if key == e.key {
						exist = true
						break
					}
				}
				if !exist {
					s.urls[e.msg] = append(s.urls[e.msg], e.key)
				}
			case "close":
				log.Debugf("[server]close event %v", e)
				delete(s.connections, e.key)
			}
		}
	}
}

func (s *SocketServer) clear() {
	for {
		select {
		case <-s.tick.C:
			now := time.Now()
			for k, conn := range s.connections {
				if conn.Deadline.Before(now) {
					conn.TimeOutEvent()
					conn.Close()
					delete(s.connections, k)
				}
			}
		}
	}
}
