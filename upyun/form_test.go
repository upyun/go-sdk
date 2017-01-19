package upyun

import (
	"path"
	"testing"
)

var (
	FORM_FILE = path.Join(ROOT, "FORM", "表单_FILE")
)

func TestFormPutFile(t *testing.T) {
	resp, err := up.FormUpload(&FormUploadConfig{
		LocalPath:      LOCAL_FILE,
		SaveKey:        FORM_FILE,
		ExpireAfterSec: 60,
	})

	Nil(t, err)
	NotNil(t, resp)
}

func TestFormPutApps(t *testing.T) {
	thumb := map[string]interface{}{
		"name":           "thumb",
		"x-gmkerl-thumb": "/fw/120",
		"save_as":        "/x120.gif",
	}

	naga := map[string]interface{}{
		"name":   "naga",
		"type":   "video",
		"avopts": "/f/mp4",
	}

	spider := map[string]interface{}{
		"name": "spiderman",
		"url":  "http://www.upyun.com/index.html",
	}

	resp, err := up.FormUpload(&FormUploadConfig{
		LocalPath:      LOCAL_FILE,
		SaveKey:        FORM_FILE,
		NotifyUrl:      NOTIFY_URL,
		ExpireAfterSec: 60,
		Apps:           []map[string]interface{}{thumb, naga, spider},
	})

	NotNil(t, resp)
	Nil(t, err)
	Equal(t, len(resp.Taskids), 3)
}
