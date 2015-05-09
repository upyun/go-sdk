package upyunpurge

import (
	"fmt"
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

	invalidURLs, purgedURLs, err := u.PurgeURLs(urls)

	if err != nil {
		t.Error(err)
	} else {
		fmt.Println("invalidURLs:", invalidURLs)
		fmt.Println("purgedURLs:", purgedURLs)
	}
}
