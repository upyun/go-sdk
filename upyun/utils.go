package upyun

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	escape []uint32 = []uint32{
		0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */

		/*             ?>=< ;:98 7654 3210  /.-, +*)( '&%$ #"!  */
		0xfc001fff, /* 1111 1100 0000 0000  0001 1111 1111 1111 */

		/*             _^]\ [ZYX WVUT SRQP  ONML KJIH GFED CBA@ */
		0x78000001, /* 0111 1000 0000 0000  0000 0000 0000 0001 */

		/*              ~}| {zyx wvut srqp  onml kjih gfed cba` */
		0xb8000001, /* 1011 1000 0000 0000  0000 0000 0000 0001 */

		0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
		0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
		0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
		0xffffffff, /* 1111 1111 1111 1111  1111 1111 1111 1111 */
	}
	hexMap = "0123456789ABCDEF"
)

func makeRFC1123Date(d time.Time) string {
	utc := d.UTC().Format(time.RFC1123)
	return strings.ReplaceAll(utc, "UTC", "GMT")
}

func makeUserAgent(version string) string {
	return fmt.Sprintf("UPYUN Go SDK V2/%s", version)
}

func md5Str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func base64ToStr(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func hmacSha1(key string, data []byte) []byte {
	hm := hmac.New(sha1.New, []byte(key))
	hm.Write(data)
	return hm.Sum(nil)
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func escapeUri(s string) string {
	size := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escape[c>>5]&(1<<(c&0x1f)) > 0 {
			size += 3
		} else {
			size++
		}
	}

	ret := make([]byte, size)
	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escape[c>>5]&(1<<(c&0x1f)) > 0 {
			ret[j] = '%'
			ret[j+1] = hexMap[c>>4]
			ret[j+2] = hexMap[c&0xf]
			j += 3
		} else {
			ret[j] = c
			j += 1
		}
	}
	return string(ret)
}

func unescapeUri(s string) string {
	n := 0
	for i := 0; i < len(s); n++ {
		switch s[i] {
		case '%':
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				// if not correct, return original string
				return s
			}
			i += 3
		default:
			i++
		}
	}

	t := make([]byte, n)
	j := 0
	for i := 0; i < len(s); j++ {
		switch s[i] {
		case '%':
			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])
			i += 3
		default:
			t[j] = s[i]
			i++
		}
	}
	return string(t)
}

var readHTTPBody = ioutil.ReadAll

func readHTTPBodyToStr(resp *http.Response) (string, error) {
	b, err := readHTTPBody(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", errorOperation("read body", err)
	}
	return string(b), nil
}

func addQueryToUri(rawurl string, kwargs map[string]string) string {
	u, _ := url.ParseRequestURI(rawurl)
	q := u.Query()
	for k, v := range kwargs {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func encodeQueryToPayload(kwargs map[string]string) string {
	payload := url.Values{}
	for k, v := range kwargs {
		payload.Set(k, v)
	}
	return payload.Encode()
}

func readHTTPBodyToInt(resp *http.Response) (int64, error) {
	b, err := readHTTPBody(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, errorOperation("read body", err)
	}

	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0, errorOperation("parse int", err)
	}
	return n, nil
}

func parseStrToInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func md5File(f io.ReadSeeker) (string, error) {
	offset, _ := f.Seek(0, 0)
	defer f.Seek(offset, 0)
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

type JsonFileInfo struct {
	ContentType  string `json:"type"`
	Name         string `json:"name"`
	Length       int64  `json:"length"`
	LastModified int64  `json:"last_modified"`
}
type JsonFiles struct {
	Files []*JsonFileInfo `json:"files"`
	Iter  string          `json:"iter"`
}

func parseBodyToFileInfos(b []byte) (iter string, fInfos []*FileInfo, err error) {
	files := &JsonFiles{}
	err = json.Unmarshal(b, files)
	if err != nil {
		return
	}

	iter = files.Iter
	fInfos = make([]*FileInfo, len(files.Files))

	for i, f := range files.Files {
		fInfos[i] = &FileInfo{
			Name:        f.Name,
			IsDir:       f.ContentType == "folder",
			ContentType: f.ContentType,
			Size:        f.Length,
			Time:        time.Unix(f.LastModified, 0),
		}
	}
	return
}
