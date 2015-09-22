// package upyun is used for your UPYUN bucket
// this sdk implement purge api, form api, http rest api
package upyun

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "1.1.0"
)

// Auto: Auto detected, based on user's internet
// Telecom: (ISP) China Telecom
// Cnc:     (ISP) China Unicom
// Ctt:     (ISP) China Tietong
// purgeEndpoint: endpoint used for purging
// Default(Min/Max)ChunkSize: set the buffer size when doing copy operation
// defaultConnectTimeout: connection timeout when connect to upyun endpoint
const (
	Auto    = "v0.api.upyun.com"
	Telecom = "v1.api.upyun.com"
	Cnc     = "v2.api.upyun.com"
	Ctt     = "v3.api.upyun.com"

	purgeEndpoint = "purge.upyun.com"

	defaultChunkSize      = 32 * 1024
	defaultConnectTimeout = 60
)

// chunkSize: chunk size when copy
var (
	chunkSize = defaultChunkSize
	endpoints = [...]string{
		Auto, Telecom, Cnc, Ctt,
	}
)

// Util functions

// User Agent
func makeUserAgent() string {
	return fmt.Sprintf("UPYUN Go SDK %s", Version)
}

// Greenwich Mean Time
func genRFC1123Date() string {
	return time.Now().UTC().Format(time.RFC1123)
}

// make md5 from string
func md5Str(s string) (ret string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// URL encode
func encodeURL(uri string) string {
	return base64.URLEncoding.EncodeToString([]byte(uri))
}

// filter useless headers
func filterHeaders(headers http.Header) http.Header {
	for k, _ := range headers {
		if strings.Contains(strings.ToLower(k), "x-") == false {
			headers.Del(k)
		}
	}
	return headers
}

// headers to string
func parseHeaders(headers http.Header) string {
	result := ""
	for k, v := range headers {
		result += fmt.Sprintf("%s:%s;", k, v[0])
	}
	return result
}

// Because of io.Copy use a 32Kb buffer, and, it is hard coded
// user can specify a chunksize with upyun.SetChunkSize
func chunkedCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, chunkSize)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])

			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return
}

// Use for http connection timeout
func timeoutDialer(timeout int) func(string, string) (net.Conn, error) {
	return func(network, addr string) (c net.Conn, err error) {
		c, err = net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
		if err != nil {
			return nil, err
		}
		return
	}
}

func SetChunkSize(chunksize int) {
	chunkSize = chunksize
}

// FileInfo when use getlist
type Info struct {
	Size int64
	Time int64
	Name string
	Type string
}

func newInfo(s string) Info {
	infoList := strings.Split(s, "\t")
	size, _ := strconv.ParseInt(infoList[2], 10, 64)
	time, _ := strconv.ParseInt(infoList[3], 10, 64)

	return Info{
		Name: infoList[0],
		Type: infoList[1],
		Size: size,
		Time: time,
	}
}

// FileInfo when HEAD file
type FileInfo struct {
	Type string
	Date string
	Size int64
}

func newFileInfo(headers http.Header) (fileInfo FileInfo) {
	for k, v := range headers {
		switch {
		case strings.Contains(k, "Type"):
			fileInfo.Type = v[0]
		case strings.Contains(k, "Size"):
			fileInfo.Size, _ = strconv.ParseInt(v[0], 10, 64)
		case strings.Contains(k, "Date"):
			fileInfo.Date = v[0]
		}
	}

	return
}

// Request Error
type ReqError struct {
	err       error
	RequestId string
}

func newRespError(requestId string, respStatus string) *ReqError {
	return &ReqError{
		RequestId: requestId,
		err:       errors.New(respStatus),
	}
}

func (r *ReqError) Error() string {
	return r.err.Error()
}

func md5sum(fd io.Reader) (string, int64, error) {
	var result []byte
	hash := md5.New()
	if written, err := io.Copy(hash, fd); err != nil {
		return "", written, err
	} else {
		return hex.EncodeToString(hash.Sum(result)), written, nil
	}
}
