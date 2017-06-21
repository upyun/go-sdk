package upyun

import (
	"fmt"
	"testing"
)

func TestPurge(t *testing.T) {
	fails, err := up.Purge([]string{
		fmt.Sprintf("http://%s.b0.upaiyun.com/demo.jpg", up.Bucket),
	})

	Nil(t, err)
	Equal(t, len(fails), 0)

	fails, err = up.Purge([]string{
		fmt.Sprintf("http://%s.b0.upaiyun.com/demo.jpg", up.Bucket),
		fmt.Sprintf("http://%s-t.b0.upaiyun.com/demo.jpg", up.Bucket),
	})

	Nil(t, err)
	Equal(t, len(fails), 1)
	Equal(t, fails[0], fmt.Sprintf("http://%s-t.b0.upaiyun.com/demo.jpg", up.Bucket))

	fails, err = up.Purge([]string{
		fmt.Sprintf("http://%s.b0.upaiyun.com/测试.jpg", up.Bucket),
	})

	Nil(t, err)
	Equal(t, len(fails), 0)
}
