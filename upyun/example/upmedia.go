package main

import (
	config "./config"
	"fmt"
	"github.com/polym/go-sdk/upyun"
	"strings"
	"time"
)

func main() {
	upm := upyun.NewUpYunMedia(config.Bucket, config.Username, config.Passwd)

	task := map[string]interface{}{
		"type":        "video",
		"format":      "flv",
		"audio_codec": "copy",
		"video_codec": "copy",
	}
	tasks := []map[string]interface{}{task}

	ids, _ := upm.PostTasks("sugar.mkv", config.Notify, tasks)

	for {
		status, _ := upm.GetProgress(strings.Join(ids, ","))
		for _, id := range ids {
			fmt.Println(id, status.Tasks[id])
		}
		time.Sleep(time.Second)
	}
}
