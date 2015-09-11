package upyun

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type UpYun struct {
	// Core
	upYunHTTPCore

	Bucket   string
	Username string
	Passwd   string

	ChunkSize int
}

func NewUpYun(bucket, username, passwd string) *UpYun {
	u := new(UpYun)
	u.Bucket = bucket
	u.Username = username
	u.Passwd = passwd

	u.endpoint = Auto

	u.httpClient = &http.Client{}
	u.SetTimeout(defaultConnectTimeout)

	return u
}

func (u *UpYun) makeRESTAuth(method, uri, date, lengthStr string) string {
	sign := []string{method, uri, date, lengthStr, md5Str(u.Passwd)}

	return "UpYun " + u.Username + ":" + md5Str(strings.Join(sign, "&"))
}

func (u *UpYun) makePurgeAuth(purgeList, date string) string {
	sign := []string{purgeList, u.Bucket, date, md5Str(u.Passwd)}

	return "UpYun " + u.Bucket + ":" + u.Username + ":" + md5Str(strings.Join(sign, "&"))
}

func (u *UpYun) Usage() (int64, error) {
	result, _, err := u.doRESTRequest("GET", "/?usage", nil, nil)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(result, 10, 64)
}

func (u *UpYun) Mkdir(key string) error {
	headers := make(map[string]string)

	headers["mkdir"] = "true"
	headers["folder"] = "true"

	_, _, err := u.doRESTRequest("POST", key, headers, nil)

	return err
}

func (u *UpYun) Put(key string, value io.Reader, useMD5 bool, secret string) (http.Header, error) {
	headers := make(map[string]string)
	headers["mkdir"] = "true"

	// secret

	if secret != "" {
		headers["Content-Secret"] = secret
	}

	// Get Content length

	/// if is file

	switch v := value.(type) {
	case *os.File:
		if useMD5 {
			hash := md5.New()

			_, err := chunkedCopy(hash, value)
			if err != nil {
				return nil, err
			}

			headers["Content-MD5"] = fmt.Sprintf("%x", hash.Sum(nil))

			// seek to origin of file
			_, err = v.Seek(0, 0)
			if err != nil {
				return nil, err
			}
		}

		fileInfo, err := v.Stat()
		if err != nil {
			return nil, err
		}

		headers["Content-Length"] = strconv.FormatInt(fileInfo.Size(), 10)

		_, rtHeaders, err := u.doRESTRequest("PUT", key, headers, value)

		return rtHeaders, err

	case io.Reader:
		buf, err := ioutil.ReadAll(v)
		if err != nil {
			return nil, err
		}

		headers["Content-Length"] = strconv.Itoa(len(buf))

		if useMD5 {
			headers["Content-MD5"] = fmt.Sprintf("%x", md5.Sum(buf))
		}

		_, rtHeaders, err := u.doRESTRequest("PUT", key, headers, bytes.NewReader(buf))

		return rtHeaders, err
	}

	return nil, errors.New("Invalid Reader")
}

func (u *UpYun) Get(key string, value io.Writer) error {
	_, _, err := u.doRESTRequest("GET", key, nil, value)

	return err
}

func (u *UpYun) Delete(key string) error {
	_, _, err := u.doRESTRequest("DELETE", key, nil, nil)

	return err
}

func (u *UpYun) GetList(key string) ([]Info, error) {
	ret, _, err := u.doRESTRequest("GET", key, nil, nil)
	if err != nil {
		return nil, err
	}

	list := strings.Split(ret, "\n")
	var infoList []Info

	for _, v := range list {
		if len(v) == 0 {
			continue
		}
		infoList = append(infoList, newInfo(v))
	}

	return infoList, nil
}

func (u *UpYun) LoopList(key, iter string, limit int) ([]Info, string, error) {
	headers := map[string]string{
		"X-List-Limit": fmt.Sprint(limit),
		"X-List-Order": "asc",
	}
	if iter != "" {
		headers["X-List-Iter"] = iter
	}

	ret, rtHeaders, err := u.doRESTRequest("GET", key, headers, nil)
	if err != nil {
		return nil, "", err
	}

	list := strings.Split(ret, "\n")
	var infoList []Info
	for _, v := range list {
		if len(v) == 0 {
			continue
		}
		infoList = append(infoList, newInfo(v))
	}

	nextIter := ""
	if _, ok := rtHeaders["X-Upyun-List-Iter"]; ok {
		nextIter = rtHeaders["X-Upyun-List-Iter"][0]
	}

	if nextIter == "g2gCZAAEbmV4dGQAA2VvZg" {
		nextIter = ""
	}

	return infoList, nextIter, nil
}

func (u *UpYun) GetInfo(key string) (FileInfo, error) {
	_, headers, err := u.doRESTRequest("HEAD", key, nil, nil)
	if err != nil {
		return FileInfo{}, err
	}

	fileInfo := newFileInfo(headers)

	return fileInfo, nil
}

func (u *UpYun) Purge(urls []string) (string, error) {
	purge := fmt.Sprintf("http://%s/purge/", purgeEndpoint)

	date := genRFC1123Date()
	purgeList := strings.Join(urls, "\n")

	headers := make(map[string]string)
	headers["Date"] = date
	headers["Authorization"] = u.makePurgeAuth(purgeList, date)
	headers["Content-Type"] = "application/x-www-form-urlencoded;charset=utf-8"

	form := make(url.Values)
	form.Add("purge", purgeList)

	body := strings.NewReader(form.Encode())
	resp, err := u.doHttpRequest("POST", purge, headers, body)
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode/100 == 2 {
		result := make(map[string][]string)
		json.Unmarshal(content, result)

		return strings.Join(result["invalid_domain_of_url"], ","), nil
	}

	return "", errors.New(string(content))
}

func (u *UpYun) doRESTRequest(method, uri string, headers map[string]string,
	value interface{}) (result string, rtHeaders http.Header, err error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	// Normalize url
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	uri = "/" + u.Bucket + uri
	url := fmt.Sprintf("http://%s%s", u.endpoint, uri)

	// date
	date := genRFC1123Date()

	// auth
	lengthStr, ok := headers["Content-Length"]
	if !ok {
		lengthStr = "0"
	}

	headers["Date"] = date
	headers["Authorization"] = u.makeRESTAuth(method, uri, date, lengthStr)

	// Get method
	rc, ok := value.(io.Reader)
	if !ok {
		rc = nil
	}

	resp, err := u.doHttpRequest(method, url, headers, rc)
	if err != nil {
		return "", nil, err
	}

	if _, ok := value.(io.Closer); ok {
		defer resp.Body.Close()
	}

	// retrive request id
	requestId := "Unknown"

	requestIds, ok := resp.Header[http.CanonicalHeaderKey("X-Request-Id")]
	if ok {
		requestId = strings.Join(requestIds, ",")
	}

	if (resp.StatusCode / 100) == 2 {
		if method == "GET" && value != nil {
			written, err := chunkedCopy(value.(io.Writer), resp.Body)

			return strconv.FormatInt(written, 10), filterHeaders(resp.Header), err
		} else if method == "GET" && value == nil {
			body, err := ioutil.ReadAll(resp.Body)
			return string(body[:]), resp.Header, err
		} else if method == "PUT" || method == "HEAD" {
			return "", filterHeaders(resp.Header), nil
		} else {
			return "", nil, nil
		}
	}

	return "", filterHeaders(resp.Header), newRespError(requestId, resp.Status)
}
