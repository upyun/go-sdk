package main

import (
	config "./config"
	"fmt"
	"github.com/polym/go-sdk/upyun"
	"time"
)

func main() {
	ump := upyun.NewUpYunMultiPart(config.Bucket, config.Apikey, 1024000)
	options := map[string]interface{}{
		"x-gmkerl-rotate": "90",
		"notify-url":      config.Notify,
		"ext-param":       "123456",
	}
	resp, err := ump.Put("/test/IMG-c"+fmt.Sprint(time.Now().Unix())+".jpg", "example/cc.jpg", 3600, options)
	fmt.Printf("%+v\n", resp, err)
}
