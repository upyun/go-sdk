package main

import (
	config "./config"
	"fmt"
	"github.com/polym/go-sdk/upyun"
	"strings"
	"time"
)

func main() {
	upm := upyun.NewUpYunMedia(config.Username, config.Passwd)

	task := map[string]interface{}{
		"type":   "video",
		"format": "flv",
	}
	tasks := []map[string]interface{}{task, task}

	ids, _ := upm.PostTasks("bigfile", "/fox.mp4", config.Notify, tasks)

	for {
		status, _ := upm.GetProgress("bigfile", strings.Join(ids, ","))
		for _, id := range ids {
			fmt.Println(id, status.Tasks[id])
		}
		time.Sleep(time.Second)
	}
}
