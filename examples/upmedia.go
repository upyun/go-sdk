package main

import (
	config "./config"
	"fmt"
	"github.com/upyun/go-sdk/upyun"
	"strings"
	"time"
)

func main() {
	upm := upyun.NewUpYunMedia(config.Bucket, config.Username, config.Passwd)

	task := map[string]interface{}{
		"type":         "thumbnail",
		"thumb_single": true,
	}
	tasks := []map[string]interface{}{task}

	ids, _ := upm.PostTasks("kai.3gp", config.Notify, "json", tasks)

	for {
		status, _ := upm.GetProgress(strings.Join(ids, ","))
		for _, id := range ids {
			fmt.Println(id, status.Tasks[id])
		}
		time.Sleep(time.Second)
	}
}
