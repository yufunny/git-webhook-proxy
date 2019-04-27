package handler

import (
	"git-proxy/server/server"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

func NotifyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	logrus.Debugf("notify body:%s", body)
	w.Write([]byte("ok"))
	res := gjson.Parse(string(body))
	url := res.Get("repository.ssh_url").String()
	logrus.Debugf("notify url:%s", url)
	server.GetServer().Dispatch(url, string(body))
}
