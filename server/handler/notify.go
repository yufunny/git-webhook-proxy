package handler

import (
	"git-webhook-proxy/server/server"
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
	gitlabHeader := r.Header.Get("X-Gitlab-Event")
	res := gjson.Parse(string(body))
	var url string
	if gitlabHeader == "" {
		url = res.Get("repository.ssh_url").String()
	} else {
		url = res.Get("project.ssh_url").String()
	}
	if url == "" {
		return
	}
	logrus.Debugf("notify url:%s", url)
	go server.GetServer().Dispatch(url, string(body))
}
