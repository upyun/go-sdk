package upyun

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// UPYUN HTTP FORM API

type UpYunForm struct {
	// Core
	upYunHTTPCore

	Key    string
	Bucket string
}

func NewUpYunForm(bucket, key string) *UpYunForm {
	up := &UpYunForm{
		Key:    key,
		Bucket: bucket,
	}

	up.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(defaultConnectTimeout),
		},
	}

	up.endpoint = Auto

	return up
}

func (uf *UpYunForm) Put(saveas, path string, expireAfter int64,
	options map[string]string) (http.Header, error) {
	if options == nil {
		options = make(map[string]string)
	}

	options["bucket"] = uf.Bucket
	options["save-key"] = saveas
	options["expiration"] = strconv.FormatInt(time.Now().Unix()+expireAfter, 10)

	args, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	policy := base64.StdEncoding.EncodeToString(args)
	sig := md5Str(policy + "&" + uf.Key)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("policy", policy)
	writer.WriteField("signature", sig)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return nil, err
	}

	if _, err = chunkedCopy(part, file); err != nil {
		return nil, err
	}

	writer.Close()

	url := fmt.Sprintf("http://%s/%s", uf.endpoint, uf.Bucket)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", makeUserAgent())

	resp, err := uf.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	rtHeaders := filterHeaders(resp.Header)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rtHeaders, err
	}

	return rtHeaders, errors.New(string(buf))
}
