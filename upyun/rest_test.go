package upyun

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"testing"
)

var (
	REST_DIR      = path.Join(ROOT, "REST")
	REST_FILE_1   = path.Join(REST_DIR, "FILE_1")
	REST_FILE_BUF = path.Join(REST_DIR, "FILE_BUF")
	REST_FILE_1M  = path.Join(REST_DIR, "FILE_1M")
	REST_OBJS     = []string{"FILE_1", "FILE_1M", "FILE_BUF"}

	BUF_CONTENT     = "UPYUN GO SDK"
	LOCAL_FILE      = "./rest.go"
	LOCAL_SAVE_FILE = LOCAL_FILE + "_bak"
)

func TestPrintEndpoint(t *testing.T) {
	c, err := net.Dial("tcp", "v0.api.upyun.com:80")
	Nil(t, err)
	fmt.Printf("v0.api: %s, client_ip: %s\n", c.RemoteAddr(), c.LocalAddr())
}

func TestUsage(t *testing.T) {
	n, err := up.Usage()
	Nil(t, err)
	Equal(t, n > 0, true)
}

func TestGetInfoDir(t *testing.T) {
	fInfo, err := up.GetInfo("/")
	Nil(t, err)
	NotNil(t, fInfo)
	Equal(t, fInfo.IsDir, true)
}

func TestMkdir(t *testing.T) {
	err := up.Mkdir(REST_DIR)
	Nil(t, err)
}

func TestPutWithFileReader(t *testing.T) {
	fd, _ := os.Open(LOCAL_FILE)
	NotNil(t, fd)
	defer fd.Close()

	_, err := up.Put(&PutObjectConfig{
		Path:   REST_FILE_1,
		Reader: fd,
		Headers: map[string]string{
			"X-Upyun-Meta-Filename": LOCAL_FILE,
		},
		UseMD5: true,
	})
	Nil(t, err)
}

func TestPutWithBuffer(t *testing.T) {
	s := BUF_CONTENT
	r := strings.NewReader(s)

	_, err := up.Put(&PutObjectConfig{
		Path:   REST_FILE_BUF,
		Reader: r,
		Headers: map[string]string{
			"Content-Length": fmt.Sprint(len(s)),
		},
		UseMD5: true,
	})
	Nil(t, err)
}

/*
func TestPutWithBufferAppend(t *testing.T) {
	s := BUF_CONTENT
	for k := 0; k < 3; k++ {
		r := strings.NewReader(s)
		err := up.Put(&PutObjectConfig{
			Path:   REST_FILE_BUF_BUF,
			Reader: r,
			Headers: map[string]string{
				"Content-Length": fmt.Sprint(len(s)),
			},
			AppendContent: true,
			UseMD5:        true,
		})
		if k != 0 {
			NotNil(t, err)
		} else {
			Nil(t, err)
		}
	}
}
*/

func TestResumePut(t *testing.T) {
	fname := "1M"
	fd, _ := os.Create(fname)
	NotNil(t, fd)
	kb := strings.Repeat("U", 1024)
	for i := 0; i < (minResumePutFileSize/1024 + 2); i++ {
		fd.WriteString(kb)
	}
	fd.Close()

	defer os.RemoveAll(fname)

	_, err := up.Put(&PutObjectConfig{
		Path:            REST_FILE_1M,
		LocalPath:       fname,
		UseMD5:          true,
		UseResumeUpload: true,
	})
	Nil(t, err)
}

func TestGetWithWriter(t *testing.T) {
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)
	fInfo, err := up.Get(&GetObjectConfig{
		Path:   REST_FILE_BUF,
		Writer: buf,
	})
	Nil(t, err)
	NotNil(t, fInfo)
	Equal(t, fInfo.IsDir, false)
	Equal(t, fInfo.Size, int64(len(BUF_CONTENT)))
	Equal(t, buf.String(), BUF_CONTENT)
}

func TestGetWithLocalPath(t *testing.T) {
	defer os.Remove(LOCAL_SAVE_FILE)
	fInfo, err := up.Get(&GetObjectConfig{
		Path:      REST_FILE_1,
		LocalPath: LOCAL_SAVE_FILE,
	})
	Nil(t, err)
	NotNil(t, fInfo)
	Equal(t, fInfo.IsDir, false)

	NotNil(t, fInfo.Meta)
	name, _ := fInfo.Meta["x-upyun-meta-filename"]
	Equal(t, name, LOCAL_FILE)

	_, err = os.Stat(LOCAL_SAVE_FILE)
	Nil(t, err)

	b1, err := ioutil.ReadFile(LOCAL_FILE)
	Nil(t, err)

	b2, err := ioutil.ReadFile(LOCAL_SAVE_FILE)
	Nil(t, err)

	Equal(t, string(b1), string(b2))
}

func TestGetInfoFile(t *testing.T) {
	fInfo, err := up.GetInfo(REST_FILE_BUF)
	Nil(t, err)
	NotNil(t, fInfo)
	Equal(t, fInfo.IsDir, false)
	Equal(t, fInfo.Name, REST_FILE_BUF)
	// as append interface
	Equal(t, fInfo.Size, int64(len(BUF_CONTENT)))
	Equal(t, fInfo.ContentType, "application/octet-stream")
}

func TestList(t *testing.T) {
	ch := make(chan *FileInfo, 10)
	files := []string{}

	go func() {
		err := up.List(&GetObjectsConfig{
			Path:        REST_DIR,
			ObjectsChan: ch,
		})
		Nil(t, err)
	}()

	for fInfo := range ch {
		files = append(files, fInfo.Name)
	}

	Equal(t, len(files), len(REST_OBJS))
	sort.Strings(files)
	sort.Strings(REST_OBJS)
	for k := range REST_OBJS {
		Equal(t, REST_OBJS[k], files[k])
	}
}

func TestIsNotExist(t *testing.T) {
	_, err := up.GetInfo("/NotExist")
	Equal(t, IsNotExist(err), true)
}

func TestModifyMetadata(t *testing.T) {
	//	time.Sleep(10 * time.Second)
	err := up.ModifyMetadata(&ModifyMetadataConfig{
		Path:      REST_FILE_1,
		Operation: "replace",
		Headers: map[string]string{
			"X-Upyun-Meta-Filename": LOCAL_SAVE_FILE,
		},
	})

	Nil(t, err)
}

func TestDelete(t *testing.T) {
	err := up.Delete(&DeleteObjectConfig{
		Path: REST_DIR,
	})
	NotNil(t, err)

	for _, obj := range REST_OBJS {
		err := up.Delete(&DeleteObjectConfig{
			Path: path.Join(REST_DIR, obj),
		})
		Nil(t, err)
	}

	err = up.Delete(&DeleteObjectConfig{
		Path: REST_DIR,
	})
	Nil(t, err)
}
