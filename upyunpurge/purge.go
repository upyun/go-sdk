package upyunpurge

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	endpoint = "http://purge.upyun.com/purge/"
	timeout  = 60

	maxRefreshedURLPerMinute = 600 - 50 // 600 is the official limit
	sleepTimeInSeconds       = 80
)

type InvalidURL struct {
	Urls []string `json:"invalid_domain_of_url"`
}

type UpYunPurge struct {
	httpClient *http.Client

	Bucket   string
	Username string
	Passwd   string

	Timeout int
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

func NewUpYunPurge(bucket, username, passwd string) *UpYunPurge {
	u := new(UpYunPurge)
	u.Bucket = bucket
	u.Username = username
	u.Passwd = passwd

	u.Timeout = timeout

	u.httpClient = &http.Client{}
	u.SetTimeout(u.Timeout)

	return u
}

func (u *UpYunPurge) makeSignature(urls string, date string) (string, error) {
	passwdMD5, err := stringMD5(u.Passwd)
	if err != nil {
		return "", err
	}

	signature := []string{urls, u.Bucket, date, passwdMD5}
	signatureString := strings.Join(signature, "&")
	signatureMD5, err := stringMD5(signatureString)
	if err != nil {
		return "", err
	}

	return signatureMD5, nil
}

func (u *UpYunPurge) makeAuth(sig string) string {
	return "UpYun " + u.Bucket + ":" + u.Username + ":" + sig
}

func (u *UpYunPurge) Version() string { return "0.1.0" }

func (u *UpYunPurge) SetTimeout(t int) {
	tranport := http.Transport{
		Dial: timeoutDialer(t),
	}

	u.httpClient = &http.Client{
		Transport: &tranport,
	}
}

// Return list of invalid urls if there is no error.
// A url in the list means it is not in the upyun bucket, therefore Upyun cannot refresh its content.
func (u *UpYunPurge) doPurgeURLs(urls []string) ([]string, error) {
	method := "POST"

	urlsString := strings.Join(urls, "\n")

	body := "purge=" + urlsString
	body, err := encodeURL(body)
	if err != nil {
		return nil, err
	}

	date := genRFC1123Date()

	sig, err := u.makeSignature(urlsString, date)
	if err != nil {
		return nil, err
	}

	auth := u.makeAuth(sig)

	req, err := http.NewRequest(method, endpoint, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Expect", "")
	req.Header.Add("Date", date)
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		var obj InvalidURL
		err = json.Unmarshal(reply, &obj)
		if err != nil {
			return nil, err
		} else {
			return obj.Urls, nil
		}
	} else {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(reply))
	}
}

// Send purge request to upyun.  Because upyun limits how many URLs one can purge per minute,
// this function will send URLs in batch if the list is larger than the limit.  It will also
// wait 80 seconds before sending next batch.
// Return value:
//   invalidURLs: List of URLs that upyun cannot purge, which usually means these URLs are not
//                in current bucket.
//   purgedURLs:  List of URLs that requests were sent without error. Please note invalid URLs
//                will also appear in this list.
func (u *UpYunPurge) PurgeURLs(urls []string) (invalidURLs []string, purgedURLs []string, err error) {
	if urls == nil || len(urls) == 0 {
		return
	}

	// Separate urls in sub-slices so that each request does not go
	// over upyun limit
	urlsInGroupOfLimit := InGroupOf(urls, maxRefreshedURLPerMinute)
	numberOfGroups := len(urlsInGroupOfLimit)

	for i := 0; i < len(urlsInGroupOfLimit); i++ {
		currentURLList := urlsInGroupOfLimit[i]

		var urlsError []string
		urlsError, err = u.doPurgeURLs(currentURLList)
		if err != nil {
			return
		}
		purgedURLs = append(purgedURLs, currentURLList...)
		invalidURLs = append(invalidURLs, urlsError...)

		// need to send another batch?
		if i < numberOfGroups-1 {
			time.Sleep(sleepTimeInSeconds * time.Second)
		}
	}

	return
}
