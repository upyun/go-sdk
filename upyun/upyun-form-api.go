package upyun

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

type FormAPIResp struct {
	Code      int    `json:"code"`
	Msg       string `json:"message"`
	Url       string `json:"url"`
	Timestamp int64  `json:"time"`
	ImgWidth  int    `json:"image-width"`
	ImgHeight int    `json:"image-height"`
	ImgFrames int    `json:"image-frames"`
	ImgType   string `json:"image-type"`
	Sign      string `json:"sign"`
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
	options map[string]string) (*FormAPIResp, error) {
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

	url := fmt.Sprintf("http://%s/%s", uf.endpoint, uf.Bucket)
	resp, err := uf.doFormRequest(url, policy, sig, path, file)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode/100 == 2 {
		var formResp FormAPIResp
		if err := json.Unmarshal(buf, &formResp); err != nil {
			return nil, err
		}
		return &formResp, nil
	}

	return nil, newRespError(string(buf), resp.Header)
}
