package main

import (
	config "./config"
	"fmt"
	"github.com/polym/go-sdk/upyun"
)

func main() {
	ump := upyun.NewUpYunMultiPart(config.Bucket, config.Apikey, 1024000)
	options := map[string]interface{}{
		//		"x-gmkerl-rotate": "90",
		"notify-url": config.Notify,
	}
	resp, err := ump.Put("/test/IMG-c10.jpg", "example/cc.jpg", 3600, options)
	fmt.Printf("%+v\n", resp, err)
}
