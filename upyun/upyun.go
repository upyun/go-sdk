package upyun

import (
	"net"
	"net/http"
	"time"
)

const (
	version = "3.0.1"

	defaultChunkSize      = 32 * 1024
	defaultConnectTimeout = time.Second * 60
)

type UpYunConfig struct {
	Bucket    string
	Operator  string
	Password  string
	Secret    string // deprecated
	Hosts     map[string]string
	UserAgent string
	UseHTTP   bool
}

type UpYun struct {
	UpYunConfig
	httpc      *http.Client
	deprecated bool
	Recorder
	stopChan chan struct{}
}

func NewUpYun(config *UpYunConfig) *UpYun {
	up := &UpYun{}
	up.Bucket = config.Bucket
	up.Operator = config.Operator
	up.Password = md5Str(config.Password)
	up.Secret = config.Secret
	up.Hosts = config.Hosts
	up.UseHTTP = config.UseHTTP
	if config.UserAgent != "" {
		up.UserAgent = config.UserAgent
	} else {
		up.UserAgent = makeUserAgent(version)
	}

	up.httpc = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (c net.Conn, err error) {
				return net.DialTimeout(network, addr, defaultConnectTimeout)
			},
		},
	}

	return up
}

func (up *UpYun) SetHTTPClient(httpc *http.Client) {
	up.httpc = httpc
}

func (up *UpYun) UseDeprecatedApi() {
	up.deprecated = true
}

func (up *UpYun) SetRecorder(recoder Recorder) {
	if recoder != nil {
		up.Recorder = recoder
		up.SetTimedTask(up.Recorder.TimedClearance)
	}
	up.stopChan = make(chan struct{})
}

func (up *UpYun) SetTimedTask(task func()) {
	t := time.NewTicker(24 * time.Hour)
	go func(t *time.Ticker) {
		defer t.Stop()
		for {
			select {
			case <-t.C:
				task()
			case <-up.stopChan:
				return
			}
		}
	}(t)
}

func (up *UpYun) Close() {
	close(up.stopChan)
}
