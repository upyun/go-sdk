package main

import (
	config "./config"
	"fmt"
	"github.com/upyun/go-sdk/upyun"
)

func main() {
	uf := upyun.NewUpYunForm(config.Bucket, config.Secret)

	options := map[string]string{
		"x-gmkerl-rotate": "90",
		"notify-url":      config.Notify,
	}
	fmt.Print(options)

	formResp, err := uf.Put("cc.jpg", "/{year}/{mon}/{day}/upload_{filename}{.suffix}", 3600, options)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", formResp)
	}
}
