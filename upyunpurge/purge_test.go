package upyunpurge

import (
	"os"
	"strings"
	"testing"
)

func TestUpyunPurge(t *testing.T) {
	bucket := os.Getenv("UPYUN_BUCKET")
	username := os.Getenv("UPYUN_USERNAME")
	passwd := os.Getenv("UPYUN_PASSWORD")
	urlsString := os.Getenv("UPYUN_URLS")

	urls := strings.Split(urlsString, ",")

	u := NewUpYunPurge(bucket, username, passwd)

	err := u.RefreshURLs(urls)

	if err != nil {
		t.Error(err)
	}
}
