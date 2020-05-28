package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func GetRequestHref(r *http.Request) string {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	return strings.Join([]string{scheme, r.Host, r.RequestURI}, "")
}

func GetRequestHostname(r *http.Request) (hostname string) {
	if _url, err := url.Parse(GetRequestHref(r)); err == nil {
		hostname = _url.Hostname()
	}
	return
}

type getRe struct {
	Status string
	Data   string
}

func GetHttpCustomPath(roomName string) (string, error) {
	var re getRe
	outputIp := Conf().Section("rtsp").Key("out_put_ip").MustString("localhost")
	roomKeyPath := fmt.Sprintf("http://%s:8090/control/get?room=%v", outputIp, roomName)
	response, err := http.Get(roomKeyPath)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		return "", err
	}

	json.Unmarshal(body, &re)

	customPath := fmt.Sprintf("rtmp://%s:1935/%s/%s", outputIp, "live", re.Data)

	return customPath, nil
}
