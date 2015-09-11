package upyun

import (
	//	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

//UPYUN MEDIA API
type UpYunMedia struct {
	// Core
	upYunHTTPCore

	username string
	passwd   string
}

func NewUpYunMedia(user, pass string) *UpYunMedia {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(defaultConnectTimeout),
		},
	}

	up := &UpYunMedia{
		username: user,
		passwd:   md5Str(pass),
	}

	// inherit from upYunHTTPCore
	up.endpoint = "p0.api.upyun.com"
	up.httpClient = client

	return up

}

func (upm *UpYunMedia) makeMediaAuth(kwargs map[string]string) string {
	var keys []string
	for k, _ := range kwargs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	auth := ""
	for _, k := range keys {
		auth += k + kwargs[k]
	}

	return fmt.Sprintf("UPYUN %s:%s", upm.username,
		md5Str(upm.username+auth+upm.passwd))
}

func (upm *UpYunMedia) PostTasks(bucket, src, notify,
	tasks string) ([]string, http.Header, error) {

	kwargs := map[string]string{
		"bucket_name": bucket,
		"source":      src,
		"notify_url":  notify,
		"tasks":       tasks,
	}

	resp, err := upm.doMediaRequest("POST", "/pretreatment", kwargs)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	fmt.Println(resp.Header)
	rtHeaders := filterHeaders(resp.Header)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, rtHeaders, err
	}

	if resp.StatusCode/2 == 100 {
		var ids []string
		err = json.Unmarshal(buf, &ids)
		if err != nil {
			return nil, rtHeaders, err
		}
		return ids, rtHeaders, err
	}

	return nil, rtHeaders, errors.New(string(buf))
}

func (upm *UpYunMedia) GetProgress(bucket,
	task_ids string) (string, http.Header, error) {

	kwargs := map[string]string{
		"bucket_name": bucket,
		"task_ids":    task_ids,
	}

	resp, err := upm.doMediaRequest("GET", "/status", kwargs)
	if err != nil {
		return "", nil, err
	}

	defer resp.Body.Close()

	rtHeaders := filterHeaders(resp.Header)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", rtHeaders, err
	}

	return string(buf), rtHeaders, err
}

func (upm *UpYunMedia) doMediaRequest(method, path string,
	kwargs map[string]string) (*http.Response, error) {

	// Normalize url
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := fmt.Sprintf("http://%s%s", upm.endpoint, path)

	// Set Headers
	headers := make(map[string]string)
	date := genRFC1123Date()
	headers["Date"] = date
	headers["Authorization"] = upm.makeMediaAuth(kwargs)

	// Payload
	var options []string
	for k, v := range kwargs {
		options = append(options, k+"="+v)
	}
	payload := strings.Join(options, "&")

	if method == "GET" {
		url = url + "?" + payload
		return upm.doHttpRequest(method, url, headers, nil)
	} else {
		if method == "POST" {
			headers["Content-Length"] = fmt.Sprint(len(payload))
			return upm.doHttpRequest(method, url, headers,
				strings.NewReader(payload))
		}
	}

	return nil, errors.New("Unknown method")
}
