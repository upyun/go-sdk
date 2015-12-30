package main

import (
	config "./config"
	"fmt"
	"github.com/upyun/go-sdk/upyun"
	"time"
)

func main() {
	ump := upyun.NewUpYunMultiPart(config.Bucket, config.Secret, 1024000)
	options := map[string]interface{}{
		"x-gmkerl-rotate": "90",
		"notify-url":      config.Notify,
		"ext-param":       "123456",
	}
	resp, err := ump.Put("cc.jpg", "/test/IMG-c"+fmt.Sprint(time.Now().Unix())+".jpg", 3600, options)
	fmt.Printf("%+v %v\n", resp, err)
}
