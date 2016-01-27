package upyun

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	username      = os.Getenv("UPYUN_USERNAME")
	password      = os.Getenv("UPYUN_PASSWORD")
	bucket        = os.Getenv("UPYUN_BUCKET")
	apikey        = os.Getenv("UPYUN_SECRET")
	up            = NewUpYun(bucket, username, password)
	upf           = NewUpYunForm(bucket, apikey)
	ump           = NewUpYunMultiPart(bucket, apikey, 1024000)
	upm           = NewUpYunMedia(bucket, username, password)
	testPath      = "/gosdk"
	upload        = "upyun-rest-api.go"
	uploadInfo, _ = os.Lstat(upload)
	uploadSize    = uploadInfo.Size()
	download      = "/tmp/xxx.go"

	length    int
	err       error
	fd        *os.File
	upInfo    *FileInfo
	upInfos   []*FileInfo
	formResp  *FormAPIResp
	mergeResp *MergeResp
)

func TestUsage(t *testing.T) {
	if _, err := up.Usage(); err != nil {
		fmt.Println(err)
		t.Errorf("failed to get Usage. %v", err)
	}
}

func TestSetEndpoint(t *testing.T) {
	for _, ed := range []int{Telecom, Cnc, Ctt, Auto} {
		if err = up.SetEndpoint(ed); err == nil {
			_, err = up.Usage()
		}
		if err != nil {
			t.Errorf("failed to SetEndpoint. %v", ed, err)
		}
	}
	if err = up.SetEndpoint(5); err == nil {
		t.Errorf("invalid SetEndpoint")
	}
}

func TestMkdir(t *testing.T) {
	if _, err = up.GetInfo(testPath); err == nil {
		t.Error(testPath, "already exists")
		//		t.Fail()
	}
	if err = up.Mkdir(testPath); err == nil {
		_, err = up.GetInfo(testPath)
	}
	if err != nil {
		t.Errorf("failed to Mkdir. %v", err)
	}
}

func TestPut(t *testing.T) {
	// put file
	if fd, err = os.Open(upload); err != nil {
		t.Skipf("failed to open %s %v", upload, err)
	}

	_, err = up.Put(testPath+"/"+upload, fd, false, nil)
	if err != nil {
		t.Errorf("failed to put %v", err)
	}

	fd, _ = os.Open(upload)
	_, err = up.Put(testPath+"/dir2/"+upload, fd, true, map[string]string{"Content-Type": "video/mp4"})
	if err != nil {
		t.Errorf("failed to put %v", err)
	}

	// put buf
	b := bytes.NewReader([]byte("UPYUN GO SDK"))
	_, err = up.Put(testPath+"/"+upload+".buf", b, false, nil)
	if err != nil {
		t.Errorf("failed to put %v", err)
	}
}

func TestGet(t *testing.T) {
	fd, err = os.OpenFile(download, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Skipf("failed to open %s %v", download, err)
	}

	defer os.Remove(download)

	if length, err = up.Get(testPath+"/"+upload, fd); err != nil {
		t.Errorf("failed to get %s %v", testPath+"/"+upload, err)
	}

	if length != int(uploadSize) {
		t.Errorf("size not equal %d != %d", length, uploadSize)
	}

	dInfo, _ := fd.Stat()
	if dInfo.Size() != uploadSize {
		t.Errorf("size not equal %d != %d", dInfo.Size, uploadSize)
	}
}

func TestGetInfo(t *testing.T) {
	if upInfo, err = up.GetInfo(testPath); err != nil {
		t.Errorf("failed to GetInfo %s %v", testPath, err)
	}
	if upInfo.Type != "folder" {
		t.Errorf("%s not folder", testPath)
	}

	if upInfo, err = up.GetInfo(testPath + "/" + upload); err != nil {
		t.Errorf("failed to GetInfo %s %v", testPath+"/"+upload, err)
	} else {
		if upInfo.Type != "file" {
			t.Errorf("%s not file", testPath+"/"+upload)
		}
		if upInfo.Size != uploadSize {
			t.Errorf("size not equal %d != %d", upInfo.Size, uploadSize)
		}
	}

	if upInfo, err = up.GetInfo(testPath + "/up"); upInfo != nil || err == nil {
		t.Errorf("%s should not exist", testPath+"/up")
	}
}

func TestGetList(t *testing.T) {
	if upInfos, err = up.GetList(testPath); err != nil {
		t.Errorf("failed to GetList %s %v", testPath, err)
	}

	if len(upInfos) != 3 {
		t.Errorf("failed to GetList %s %d != 3", testPath, len(upInfos))
	}
}

func TestGetLargeList(t *testing.T) {
	ch := up.GetLargeList(testPath, false)
	count := 0
	for {
		var more bool
		upInfo, more = <-ch
		if !more {
			break
		}
		count++
	}
	if count != 3 {
		t.Errorf("GetLargeList %d != 3", count)
	}

	ch = up.GetLargeList(testPath, true)
	count = 0
	for {
		var more bool
		upInfo, more = <-ch
		if !more {
			break
		}
		count++
	}
	if count != 4 {
		t.Errorf("GetLargeList recursive %d != 4", count)
	}
}

func TestDelete(t *testing.T) {
	// delete file
	path := testPath + "/" + upload
	if err = up.Delete(path); err != nil {
		t.Errorf("failed to Delete %s %v", path, err)
	}

	path = testPath + "/" + upload + ".buf"
	if err = up.Delete(path); err != nil {
		t.Errorf("failed to Delete %s %v", path, err)
	}

	path = testPath + "/dir2/" + upload
	if err = up.Delete(path); err != nil {
		t.Errorf("failed to Delete %s %v", path, err)
	}

	// delete not empty folder
	path = testPath
	if err = up.Delete(path); err == nil {
		t.Errorf("Delete no-empty folder should failed %s", path)
	}
	// delete empty folder
	path = testPath + "/dir2"
	if err = up.Delete(path); err != nil {
		t.Errorf("failed to Delete empty folder %s %v", path, err)
	}

	path = testPath
	if err = up.Delete(path); err != nil {
		t.Errorf("failed to Delete empty folder %s %v", path, err)
	}
}

func TestPurge(t *testing.T) {
	var s string
	s, err = up.Purge([]string{"http://www.baidu.com",
		fmt.Sprintf("http://%s.b0.upaiyun.com/%s", up.Bucket, testPath+"/"+upload)})
	if err != nil {
		t.Errorf("failed to Purge %v", err)
	}

	if s != "http://www.baidu.com" {
		t.Errorf("%s != baidu", s)
	}
}

func TestFormAPI(t *testing.T) {
	formResp, err = upf.Put(upload,
		testPath+"/upload_{filename}{.suffix}", 3600, nil)
	if err != nil {
		t.Errorf("failed to put %s %v", upload, err)
		return
	}
	if err = up.Delete(formResp.Url); err == nil {
		err = up.Delete(testPath)
	}
	if err != nil {
		t.Errorf("failed to remove %s %v", formResp.Url, err)
	}
}

func TestMultiPart(t *testing.T) {
	mergeResp, err = ump.Put(upload, testPath+"/multipart", 3600, nil)
	if err != nil {
		t.Errorf("failed to put %s %v", upload, err)
		return
	}

	if err = up.Delete(mergeResp.Path); err == nil {
		err = up.Delete(testPath)
	}
	if err != nil {
		t.Errorf("failed to remove %s %v", mergeResp.Path, err)
	}
}

func TestMedia(t *testing.T) {
	task := map[string]interface{}{
		"type":         "thumbnail",
		"thumb_single": true,
	}
	tasks := []map[string]interface{}{task}

	if ids, err := upm.PostTasks("kai.3gp", "http://www.upyun.com/notify", tasks); err != nil {
		t.Errorf("failed to post tasks %v %v", tasks, err)
	} else {
		if _, err = upm.GetProgress(strings.Join(ids, ",")); err != nil {
			t.Errorf("failed to get progress %v %v", ids, err)
		}
	}
}
