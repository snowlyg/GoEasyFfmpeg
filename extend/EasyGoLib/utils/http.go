package utils

import (
	"encoding/json"
	"errors"
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
	customPath := ""
	if getServerType() == "rtmp" {
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
		customPath = fmt.Sprintf("rtmp://%s:1935/%s/%s", outputIp, "live", re.Data)
		return customPath, nil
	} else if getServerType() == "rtsp" {
		customPath = fmt.Sprintf("rtsp://%s:8554/%s", outputIp, roomName)
		return customPath, nil
	} else {
		return "", errors.New("错误推流服务类型")

	}
}

func getServerType() string {
	paramStr := Conf().Section("rtsp").Key("decoder").MustString("-strict -2 -threads 2 -c:v copy -c:a copy -f rtsp")
	paramStrs := strings.Split(paramStr, " ")
	serverType := paramStrs[len(paramStrs)-1]
	return serverType
}

func GetOutPutUrl(roomName, transType string) string {
	outputIp := Conf().Section("rtsp").Key("out_put_ip").MustString("localhost")
	url := fmt.Sprintf("rtsp://%s:8554/%v", outputIp, roomName)
	if getServerType() == "rtmp" {
		if transType == "RTMP" {
			url = fmt.Sprintf("rtmp://%s:1935/live/%v", outputIp, roomName)
		} else if transType == "HLS" {
			url = fmt.Sprintf("rtmp://%s:7002/live/%v.mu38", outputIp, roomName)
		} else if transType == "FLV" {
			url = fmt.Sprintf("rtmp://%s:7001/live/%v.flv", outputIp, roomName)
		}
		return url
	} else if getServerType() == "rtsp" {
		url = fmt.Sprintf("rtsp://%s:8554/%v", outputIp, roomName)
		return url
	} else {

		return ""
	}
}
