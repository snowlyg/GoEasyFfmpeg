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

type GetRe struct {
	Status string
	Data   string
}

func GetHttp(roomName string) (GetRe, error) {
	var re GetRe

	roomKeyPath := fmt.Sprintf("http://localhost:8090/control/get?room=%v", roomName)
	response, err := http.Get(roomKeyPath)
	if err != nil {
		return re, err
	}
	defer response.Body.Close()
	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		return re, err
	}

	json.Unmarshal(body, &re)

	return re, nil
}
