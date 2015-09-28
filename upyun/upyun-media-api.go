package upyun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// UPYUN MEDIA API
type UpYunMedia struct {
	upYunHTTPCore // HTTP Core

	username string
	passwd   string
}

// status response
type MediaStatusResp struct {
	Tasks map[string]interface{} `json:"tasks"`
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

// Send Media Tasks Reqeust
func (upm *UpYunMedia) PostTasks(bucket, src, notify string,
	tasks []map[string]interface{}) ([]string, error) {
	data, err := json.Marshal(tasks)
	if err != nil {
		return nil, err
	}

	kwargs := map[string]string{
		"bucket_name": bucket,
		"source":      src,
		"notify_url":  notify,
		"tasks":       base64Str(data),
	}

	resp, err := upm.doMediaRequest("POST", "/pretreatment", kwargs)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/2 == 100 {
		var ids []string
		err = json.Unmarshal(buf, &ids)
		if err != nil {
			return nil, err
		}
		return ids, err
	}

	return nil, newRespError(string(buf), resp.Header)
}

// Get Task Progress
func (upm *UpYunMedia) GetProgress(bucket,
	task_ids string) (*MediaStatusResp, error) {

	kwargs := map[string]string{
		"bucket_name": bucket,
		"task_ids":    task_ids,
	}

	resp, err := upm.doMediaRequest("GET", "/status", kwargs)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode/2 == 100 {
		var status MediaStatusResp
		if err := json.Unmarshal(buf, &status); err != nil {
			return nil, err
		}
		return &status, nil
	}

	return nil, newRespError(string(buf), resp.Header)
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
		return upm.doHTTPRequest(method, url, headers, nil)
	} else {
		if method == "POST" {
			headers["Content-Length"] = fmt.Sprint(len(payload))
			return upm.doHTTPRequest(method, url, headers,
				strings.NewReader(payload))
		}
	}

	return nil, errors.New("Unknown method")
}
