package server

import (
	"git-webhook-proxy/server/protocol"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn      net.Conn
	Events    chan *event
	key       string
	read      bool
	Deadline  time.Time
	SessionId int
	lock      *sync.Mutex //  锁
}

type PMessage struct {
	Id  int
	Msg string
}

type event struct {
	trigger string
	key     string
	msg     string
}

func (c *Connection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.conn.Close()
	c.Events <- &event{
		"close",
		c.key,
		"",
	}
}

func (c *Connection) Listen() {
	defer c.Close()
	//缓冲区，存储被截断的数据
	tmpBuffer := make([]byte, 0)
	//接收解包
	readerChannel := make(chan *PMessage, 10000)
	go c.reader(readerChannel)

	buffer := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			logrus.Errorf("connection error: %s, %s", c.conn.RemoteAddr().String(), err.Error())
			return
		}
		logrus.Debugf("receive msg :%s", buffer)
		protocolId, msg := protocol.DePack(append(tmpBuffer, buffer[:n]...))
		if protocolId < 0 {
			tmpBuffer = make([]byte, 0)
		} else if protocolId == 0 {
			tmpBuffer = append(tmpBuffer, buffer[:n]...)
			continue
		} else {
			pMsg := &PMessage{
				protocolId,
				string(msg),
			}
			readerChannel <- pMsg //接收的信息写入通道
			tmpBuffer = make([]byte, 0)
		}

	}
}

func (c *Connection) heartbeat() {
	c.response(protocol.Heartbeat+1, "{\"code\":200}")
}

//获取通道数据
func (c *Connection) reader(msg chan *PMessage) {
	for {
		select {
		case data := <-msg:
			logrus.Debugf("receive msg: %d,%s", data.Id, data.Msg) //打印通道内的信息
			jsonRes := gjson.Parse(data.Msg)
			c.Deadline = time.Now().Add(5 * time.Minute)
			switch data.Id {
			case protocol.Heartbeat:
				c.heartbeat()
			case protocol.Subscribe:
				c.subscribe(&jsonRes)
			}

		}
	}
}

func (c *Connection) response(protocolId int, msg string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	logrus.Debugf("response to client:%s:%d:%s", c.conn.RemoteAddr().String(), protocolId, msg)
	c.conn.Write(protocol.EnPack(protocolId, []byte(msg)))
}

func (c *Connection) subscribe(result *gjson.Result) {
	for _, url := range result.Array() {
		c.Events <- &event{
			"subscribe",
			c.key,
			url.Str,
		}
	}
}

func (c *Connection) Send(body string) {
	c.response(200, body)
}

func (c *Connection) TimeOutEvent() {
	c.response(310, "{\"code\":200, \"message\":\"bye bye~\"}")
}
