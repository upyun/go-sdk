package upyun

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"testing"
)

var (
	MP4_URL      = "http://prog-test.b0.upaiyun.com/const/kai.3gp"
	MP4_SAVE_AS  = path.Join(ROOT, "kai.3gp")
	MP4_TASK_IDS []string

	JPG_URL     = "http://upyun-xiang-1.b0.upaiyun.com/const/pu.jpg"
	FACE_URL    = "http://upyun-xiang-1.b0.upaiyun.com/const/face.jpg"
	JPG_SOURCE  = "/source/pu.jpg"
	MP4_SOURCE  = "/source/kai.3gp"
	JPG_SAVE_AS = path.Join(ROOT, "foo_{index}.jpg")
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

//由于是异步操作，不能确保文件已存在
func TestImgaudit(t *testing.T) {
	task := map[string]interface{}{
		"url":     JPG_URL,
		"save_as": JPG_SOURCE,
	}
	ids, err := up.CommitTasks(&CommitTasksConfig{
		AppName:   "spiderman",
		NotifyUrl: NOTIFY_URL,
		Tasks:     []interface{}{task},
	})

	Nil(t, err)
	Equal(t, len(ids), 1)

	task = map[string]interface{}{
		"source": JPG_SOURCE,
	}

	ids, err = up.CommitTasks(&CommitTasksConfig{
		AppName:   "imgaudit",
		NotifyUrl: NOTIFY_URL,
		Tasks:     []interface{}{task},
	})

	Nil(t, err)
	Equal(t, len(ids), 1)

}

//由于是异步操作，不能确保文件已存在
func TestVideoaudit(t *testing.T) {
	task := map[string]interface{}{
		"url":     MP4_URL,
		"save_as": MP4_SOURCE,
	}
	ids, err := up.CommitTasks(&CommitTasksConfig{
		AppName:   "spiderman",
		NotifyUrl: NOTIFY_URL,
		Tasks:     []interface{}{task},
	})

	Nil(t, err)
	Equal(t, len(ids), 1)

	task = map[string]interface{}{
		"source":  MP4_SOURCE,
		"save_as": JPG_SAVE_AS,
	}

	ids, err = up.CommitTasks(&CommitTasksConfig{
		AppName:   "videoaudit",
		NotifyUrl: NOTIFY_URL,
		Tasks:     []interface{}{task},
	})

	Nil(t, err)
	Equal(t, len(ids), 1)
}

func TestLiveaudit(t *testing.T) {
	result, err := up.CommitSyncTasks(SyncTaskConfig{
		Param: map[string]interface{}{
			"source":     "rtmp://live.hkstv.hk.lxdns.com/live/hks",
			"save_as":    JPG_SAVE_AS,
			"notify_url": NOTIFY_URL,
		}}, "/liveaudit/create")

	Nil(t, err)
	Equal(t, result["status"], float64(200))

	if result["status"] == float64(200) {
		result, err := up.CommitSyncTasks(SyncTaskConfig{
			Param: map[string]interface{}{
				"task_id": result["task_id"].(string),
			}}, "/liveaudit/cancel")

		Nil(t, err)
		Equal(t, result["status"], float64(200))
	}
}

func TestFaceDetect(t *testing.T) {
	resp, err := http.Get(FACE_URL + "!/face/detection")
	Nil(t, err)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	Nil(t, err)
	if err != nil {
		var result map[string]interface{}
		err = json.Unmarshal(b, &result)
		Nil(t, err)
		NotEqual(t, result["face"], nil)
	}

}
