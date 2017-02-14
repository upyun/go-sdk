package upyun

import (
	"path"
	"testing"
)

var (
	MP4_URL      = "http://prog-test.b0.upaiyun.com/const/kai.3gp"
	MP4_SAVE_AS  = path.Join(ROOT, "kai.3gp")
	MP4_TASK_IDS []string
)

func TestSpider(t *testing.T) {
	task := map[string]interface{}{
		"url":     MP4_URL,
		"save_as": MP4_SAVE_AS,
	}
	ids, err := up.CommitTasks(&CommitTasksConfig{
		AppName:   "spiderman",
		NotifyUrl: NOTIFY_URL,
		Tasks:     []interface{}{task},
	})

	Nil(t, err)
	Equal(t, len(ids), 1)
}

func TestNagaCommit(t *testing.T) {
	task := map[string]interface{}{
		"type":   "video",
		"avopts": "/f/mp4",
	}
	task2 := map[string]interface{}{
		"type":   "video",
		"avopts": "/f/mp3",
	}

	ids, err := up.CommitTasks(&CommitTasksConfig{
		AppName:   "naga",
		NotifyUrl: NOTIFY_URL,
		Tasks:     []interface{}{task, task2},
	})

	Nil(t, err)
	Equal(t, len(ids), 2)

	MP4_TASK_IDS = ids
}

func TestNagaProgress(t *testing.T) {
	res, err := up.GetProgress(MP4_TASK_IDS)
	Nil(t, err)
	Equal(t, len(res), 2)
}

func TestNagaResult(t *testing.T) {
	res, err := up.GetResult(MP4_TASK_IDS)
	Nil(t, err)
	Equal(t, len(res), 2)
}
