package upyun

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (up *UpYun) doHTTPRequest(method, url string, headers map[string]string,
	body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}

	req.Header.Set("User-Agent", up.UserAgent)
	if method == "PUT" || method == "POST" {
		found := false
		length := req.Header.Get("Content-Length")
		if length != "" {
			req.ContentLength, _ = strconv.ParseInt(length, 10, 64)
			found = true
		} else {
			switch v := body.(type) {
			case *os.File:
				if fInfo, err := v.Stat(); err == nil {
					req.ContentLength = fInfo.Size()
					found = true
				}
			case UpYunPutReader:
				req.ContentLength = int64(v.Len())
				found = true
			case *bytes.Buffer:
				req.ContentLength = int64(v.Len())
				found = true
			case *bytes.Reader:
				req.ContentLength = int64(v.Len())
				found = true
			case *strings.Reader:
				req.ContentLength = int64(v.Len())
				found = true
			case *io.LimitedReader:
				req.ContentLength = v.N
				found = true
			}
		}
		if found && req.ContentLength == 0 {
			req.Body = nil
		}
	}

	//	fmt.Printf("%+v\n", req)

	resp, err = up.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	err = checkResponse(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (up *UpYun) doGetEndpoint(host string) string {
	s := up.Hosts[host]
	if s != "" {
		return s
	}
	return host
}

func (up *UpYun) getEndpoint(defaultHost string) string {
	value := up.Hosts["host"]
	if value == "" {
		value = defaultHost
	}
	s := strings.TrimSpace(value)
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	if s == "" {
		s = defaultHost
	}
	scheme := "https://"
	if up.UseHTTP {
		scheme = "http://"
	}
	return fmt.Sprintf("%s%s", scheme, s)
}
