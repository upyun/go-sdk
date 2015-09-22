package upyun

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
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

// Set HTTP endpoint
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

// do http form request
func (core *upYunHTTPCore) doFormRequest(url, policy, sign,
	fpath string, fd io.Reader) (*http.Response, error) {

	body := &bytes.Buffer{}
	headers := make(map[string]string)

	// generate form data
	err := func() error {
		writer := multipart.NewWriter(body)

		defer writer.Close()

		writer.WriteField("policy", policy)
		writer.WriteField("signature", sign)
		part, err := writer.CreateFormFile("file", filepath.Base(fpath))
		if err != nil {
			return err
		}

		if _, err = chunkedCopy(part, fd); err != nil {
			return err
		}
		headers["Content-Type"] = writer.FormDataContentType()

		return nil
	}()
	if err != nil {
		return nil, err
	}

	return core.doHttpRequest("POST", url, headers, body)
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
		if length != "" {
			req.ContentLength, _ = strconv.ParseInt(length, 10, 64)
		}
	}

	return core.httpClient.Do(req)
}
