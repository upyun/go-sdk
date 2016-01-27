// SEE upyun_test.go
package main

import (
	config "./config"
	"fmt"
	"github.com/upyun/go-sdk/upyun"
	"os"
	"time"
)

func main() {
	up := upyun.NewUpYun(config.Bucket, config.Username, config.Passwd)
	headers := map[string]string{
		"x-gmkerl-watermark-type":   "text",
		"x-gmkerl-watermark-font":   "simhei",
		"x-gmkerl-watermark-color":  "#faf1fb",
		"x-gmkerl-watermark-size":   "20",
		"x-gmkerl-watermark-text":   "UPYUN",
		"x-gmkerl-watermark-border": "#40404085",
		"x-gmkerl-watermark-margin": "10,10",
	}

	fd, _ := os.Open("cc.jpg")
	x := fmt.Sprintf("/wm/cc%d.jpg", time.Now().Unix()%10000)
	fmt.Println(up.Put(x, fd, false, headers))

	c := up.GetLargeList("/", true)
	for {
		v, more := <-c
		if !more {
			break
		}
		fmt.Println(v)
	}
}
