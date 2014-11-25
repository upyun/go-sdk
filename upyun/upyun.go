package upyun

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Auto    = "v0.api.upyun.com"
	Telecom = "v1.api.upyun.com"
	Cnc     = "v2.api.upyun.com"
	Ctt     = "v3.api.upyun.com"
)

const (
	DefaultMaxChunkSize = 8192
	DefaultMinChunkSize = 1
)

var endpointList = []string{Auto, Telecom, Cnc, Ctt}

type Info struct {
	Name string
	Type string
	Size int64
	Time int64
}

func newInfo(s string) (info *Info) {
	info = new(Info)

	infoList := strings.Split(s, "\t")
	info.Name = infoList[0]
	info.Type = infoList[1]
	info.Size, _ = strconv.ParseInt(infoList[2], 10, 64)
	info.Time, _ = strconv.ParseInt(infoList[3], 10, 64)

	return
}

type RespError struct {
	err       error
	RequestId string
}

func newRespError(requestId string, respStatus string) (respError *RespError) {
	respError = new(RespError)

	respError.RequestId = requestId
	respError.err = errors.New(respStatus)

	return
}

func (r *RespError) Error() string {
	return r.err.Error()
}

type FileInfo struct {
	Type string
	Date string
	Size int64
}

func newFileInfo(s string) (fileInfo *FileInfo) {
	fileInfo = new(FileInfo)

	headers := strings.Split(s, "\n")
	for _, h := range headers {
		if h == "" {
			continue
		}

		tmp := strings.Split(h, ":")
		k, v := tmp[0], tmp[1]
		switch {
		case strings.Contains(k, "type"):
			fileInfo.Type = v
		case strings.Contains(k, "size"):
			fileInfo.Size, _ = strconv.ParseInt(v, 10, 64)
		case strings.Contains(k, "date"):
			fileInfo.Date = v
		}
	}

	return
}

type UpYun struct {
	httpClient *http.Client

	Bucket   string
	Username string
	Passwd   string
	Endpoint string

	Timeout   int
	ChunkSize int
}

func stringMD5(s string) (string, error) {
	hasher := md5.New()
	if _, err := hasher.Write([]byte(s)); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func genRFC1123Date() string { return time.Now().UTC().Format(time.RFC1123) }

func encodeURL(uri string) (string, error) {
	Url, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	return Url.String(), nil
}

func timeoutDialer(timeout int) func(string, string) (net.Conn, error) {
	return func(network, addr string) (c net.Conn, err error) {
		c, err = net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
		if err != nil {
			return nil, err
		}
		return c, err
	}
}

func NewUpYun(bucket, username, passwd string) *UpYun {
	u := new(UpYun)
	u.Bucket = bucket
	u.Username = username
	u.Passwd = passwd

	u.Timeout = 60
	u.ChunkSize = 8192
	u.Endpoint = Auto

	u.httpClient = &http.Client{}
	u.SetTimeout(u.Timeout)

	return u
}

func (u *UpYun) makeSignature(method, uri, date, contentLen_str string) (string, error) {
	passwdMD5, err := stringMD5(u.Passwd)
	if err != nil {
		return "", err
	}

	signature := []string{method, uri, date, contentLen_str, passwdMD5}
	signatureMD5, err := stringMD5(strings.Join(signature, "&"))
	if err != nil {
		return "", err
	}

	return signatureMD5, nil
}

func (u *UpYun) makeAuth(sig string) string { return "UpYun " + u.Username + ":" + sig }

func (u *UpYun) makeContentMD5(value *os.File) (string, error) {
	hasher := md5.New()
	chunk := make([]byte, u.ChunkSize)
	for {
		n, err := value.Read(chunk)

		if err != nil && err != io.EOF {
			return "", err
		}

		if n == 0 {
			break
		}
		hasher.Write(chunk)
	}

	if _, err := value.Seek(0, 0); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (u *UpYun) Version() string { return "0.1.0" }

func (u *UpYun) SetTimeout(t int) {
	tranport := http.Transport{
		Dial: timeoutDialer(t),
	}

	u.httpClient = &http.Client{
		Transport: &tranport,
	}
}

func (u *UpYun) SetEndpoint(endpoint string) (string, error) {
	for _, v := range endpointList {
		if v == endpoint {
			u.Endpoint = endpoint
			return endpoint, nil
		}
	}

	return u.Endpoint, errors.New("Invalid endpoint")
}

func (u *UpYun) SetChunkSize(chunksize int) (int, error) {
	if chunksize <= DefaultMaxChunkSize && chunksize > DefaultMinChunkSize {
		u.ChunkSize = chunksize
		return chunksize, nil
	}

	return u.ChunkSize, errors.New("Invalid chunksize")
}

func (u *UpYun) Usage() (int64, error) {
	content, err := u.doHttpRequest("GET", "/", nil, "?usage", nil)

	if err != nil {
		return -1, err
	}

	return strconv.ParseInt(content, 10, 64)
}

func (u *UpYun) Mkdir(key string) error {
	headers := make(map[string]string)

	headers["mkdir"] = "true"
	headers["folder"] = "true"

	_, err := u.doHttpRequest("POST", key, nil, "", headers)

	return err
}

func (u *UpYun) Put(key string, value *os.File, md5 bool, secret string) (string, error) {
	headers := make(map[string]string)
	headers["mkdir"] = "true"

	fi, err := value.Stat()
	if err != nil {
		return "", err
	}

	headers["Content-Length"] = strconv.FormatInt(fi.Size(), 10)

	if secret != "" {
		headers["Content-Secret"] = secret
	}

	if md5 {
		contentMD5, err := u.makeContentMD5(value)
		if err != nil {
			return "", err
		}

		headers["Content-MD5"] = contentMD5
	}

	rtHeaders, err := u.doHttpRequest("PUT", key, value, "", headers)

	return rtHeaders, err
}

func (u *UpYun) Get(key string, value *os.File) error {
	_, err := u.doHttpRequest("GET", key, value, "", nil)

	return err
}

func (u *UpYun) Delete(key string) error {
	_, err := u.doHttpRequest("DELETE", key, nil, "", nil)

	return err
}

func (u *UpYun) GetList(key string) ([]Info, error) {
	ret, err := u.doHttpRequest("GET", key, nil, "", nil)
	if err != nil {
		return nil, err
	}

	list := strings.Split(ret, "\n")
	infoList := make([]Info, len(list))
	for i, v := range list {
		infoList[i] = *newInfo(v)
	}

	return infoList, nil
}

func (u *UpYun) GetInfo(key string) (*FileInfo, error) {
	ret, err := u.doHttpRequest("HEAD", key, nil, "", nil)
	if err != nil {
		return nil, err
	}

	fileInfo := newFileInfo(ret)

	return fileInfo, nil
}

func (u *UpYun) doHttpRequest(method, uri string, value *os.File,
	args string, headers map[string]string) (string, error) {
	var _uri string

	if uri[:1] != "/" {
		_uri = "/" + u.Bucket + "/" + uri
	} else {
		_uri = "/" + u.Bucket + uri
	}

	if args != "" {
		_uri = _uri + args
	}
	_uri, err := encodeURL(_uri)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("http://%s%s", u.Endpoint, _uri)
	date := genRFC1123Date()

	contentLen_str := headers["Content-Length"]

	if contentLen_str == "" {
		contentLen_str = "0"
	}

	contentLen_int, err := strconv.ParseInt(contentLen_str, 0, 64)
	if err != nil {
		return "", err
	}

	sig, err := u.makeSignature(method, _uri, date, contentLen_str)
	if err != nil {
		return "", err
	}

	auth := u.makeAuth(sig)

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", err
	}

	if method == "PUT" {
		req.Body = value
		req.ContentLength = contentLen_int
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	req.Header.Add("Date", date)
	req.Header.Add("Authorization", auth)

	resp, err := u.httpClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var requestId string
	requestIds, ok := resp.Header[http.CanonicalHeaderKey("X-Request-Id")]
	if !ok {
		requestId = "Unknown"
	} else {
		requestId = strings.Join(requestIds, ",")
	}

	if (resp.StatusCode / 100) == 2 {
		if method == "GET" && value != nil {
			var written int64 = 0
			chunk := make([]byte, u.ChunkSize)

			for {
				n, err := resp.Body.Read(chunk)
				if err != nil && err != io.EOF {
					return "", err
				}

				if n == 0 {
					break
				}

				value.Write(chunk[:n])
				written += int64(n)
			}
			return "", nil
		} else if method == "GET" && value == nil {
			body, err := ioutil.ReadAll(resp.Body)
			return string(body[:]), err
		} else if method == "PUT" || method == "HEAD" {
			var headerStrings string
			for k, v := range resp.Header {
				if strings.Contains(strings.ToLower(k), "x-upyun-") {
					headerStrings += fmt.Sprintf("%s:%s\n", strings.ToLower(k), v[0])
				}
			}
			return headerStrings, nil
		} else {
			return "", nil
		}
	}

	return "", newRespError(requestId, resp.Status)
}
