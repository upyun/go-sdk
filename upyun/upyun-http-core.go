package upyun

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type upYunHTTPCore struct {
	endpoint   string
	httpClient *http.Client
}

// Set connect timeout
func (core *upYunHTTPCore) SetTimeout(timeout int) {
	core.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(timeout),
		},
	}
}

// Set http endpoint
func (core *upYunHTTPCore) SetEndpoint(endpoint string) (string, error) {
	for _, v := range endpoints {
		if v == endpoint {
			core.endpoint = endpoint
			return endpoint, nil
		}
	}

	err := fmt.Sprintf("Invalid endpoint, pick from Auto, Telecom, Cnc, Ctt")
	return core.endpoint, errors.New(err)
}

// do http request
func (core *upYunHTTPCore) doHttpRequest(method, url string, headers map[string]string,
	body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// User Agent
	req.Header.Set("User-Agent", makeUserAgent())

	// https://code.google.com/p/go/issues/detail?id=6738
	if method == "PUT" || method == "POST" {
		length := req.Header.Get("Content-Length")
		req.ContentLength, _ = strconv.ParseInt(length, 10, 64)
	}

	return core.httpClient.Do(req)
}
