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
	"time"
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

	err := up.Put(&PutObjectConfig{
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
	time.Sleep(time.Second)
	err := up.Put(&PutObjectConfig{
		Path:   REST_FILE_BUF,
		Reader: r,
		Headers: map[string]string{
			"Content-Length": fmt.Sprint(len(s)),
		},
		UseMD5: true,
	})
	Nil(t, err)
}

func TestCopyMove(t *testing.T) {
	s := BUF_CONTENT
	r := strings.NewReader(s)
	time.Sleep(time.Second)
	srcPath := path.Join(REST_DIR, "src_file")
	err := up.Put(&PutObjectConfig{
		Path:   srcPath,
		Reader: r,
		Headers: map[string]string{
			"Content-Length": fmt.Sprint(len(s)),
		},
		UseMD5: true,
	})
	Nil(t, err)

	time.Sleep(time.Second)
	copyPath := path.Join(REST_DIR, "copy_dest_file")
	err = up.Copy(&CopyObjectConfig{
		SrcPath:  srcPath,
		DestPath: copyPath,
	})
	Nil(t, err)

	movePath := path.Join(REST_DIR, "move_dest_file")
	err = up.Move(&MoveObjectConfig{
		SrcPath:  srcPath,
		DestPath: movePath,
	})
	Nil(t, err)

	time.Sleep(time.Second)
	err = up.Delete(&DeleteObjectConfig{
		Path: copyPath,
	})
	Nil(t, err)
	time.Sleep(time.Second)
	err = up.Delete(&DeleteObjectConfig{
		Path: movePath,
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
func testMultiUpload(t *testing.T, key string, data []byte, partSize int64, parts []int, completed bool) *InitMultipartUploadResult {
	uploadResult, err := up.InitMultipartUpload(&InitMultipartUploadConfig{
		Path:     key,
		PartSize: partSize,
	})
	Nil(t, err)
	for _, partId := range parts {
		start := int64(partId) * partSize
		end := start + partSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}
		err := up.UploadPart(uploadResult, &UploadPartConfig{
			PartID:   partId,
			PartSize: end - start,
			Reader:   bytes.NewReader(data[start:end]),
		})
		Nil(t, err)
	}
	if completed {
		err := up.CompleteMultipartUpload(uploadResult, nil)
		Nil(t, err)
	}
	return uploadResult
}
func TestMultiListParts(t *testing.T) {
	data10m := make([]byte, 10*1024*1024)
	partSize := int64(3 * 1024 * 1024)
	prefixKey := TempKey(t)

	key := path.Join(prefixKey, "upload.txt")
	initResult := testMultiUpload(t, key, data10m, partSize, []int{1, 2}, false)
	result, err := up.ListMultipartParts(initResult, &ListMultipartPartsConfig{})
	Nil(t, err)
	Equal(t, len(result.Parts), 2)

	result, err = up.ListMultipartParts(initResult, &ListMultipartPartsConfig{BeginID: 2})
	Nil(t, err)
	Equal(t, len(result.Parts), 1)
}
func TestMultiGetUpload(t *testing.T) {
	data10m := make([]byte, 10*1024*1024)
	partSize := int64(3 * 1024 * 1024)
	prefixKey := TempKey(t)
	var key string
	keyMap := make(map[string]bool)

	key = path.Join(prefixKey, "init.txt")
	keyMap[key] = true
	testMultiUpload(t, key, data10m, partSize, nil, false)
	key = path.Join(prefixKey, "upload.txt")
	keyMap[key] = true
	testMultiUpload(t, key, data10m, partSize, []int{1, 2}, false)
	key = path.Join(prefixKey, "complete.txt")
	testMultiUpload(t, key, data10m, partSize, []int{0, 1, 2, 3}, true)
	result, err := up.ListMultipartUploads(&ListMultipartConfig{
		Prefix: prefixKey,
	})
	Nil(t, err)

	Equal(t, len(result.Files), len(keyMap))
}

func TestMultipartUpload(t *testing.T) {
	fname := "10M"
	fd, _ := os.Create(fname)
	NotNil(t, fd)
	kb := strings.Repeat("U", 1024)
	for i := 0; i < (minResumePutFileSize/1024 + 2); i++ {
		fd.WriteString(kb)
	}

	defer os.RemoveAll(fname)
	defer fd.Close()

	path := REST_FILE_1M
	config := &PutObjectConfig{
		Path:                   path,
		Reader:                 fd,
		LocalPath:              fname,
		UseMD5:                 true,
		Headers:                make(map[string]string),
		MultipartUpload:        true,
		MultipartUploadWorkers: 2,
	}
	err := up.Put(config)
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
	name := fInfo.Meta["x-upyun-meta-filename"]
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
	time.Sleep(time.Second)
	err := up.Delete(&DeleteObjectConfig{
		Path: REST_DIR,
	})
	NotNil(t, err)

	for _, obj := range REST_OBJS {
		time.Sleep(time.Second)
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

func putTestFilesToBucket(t *testing.T, remotePath string) []FileInfo {
	var files []FileInfo

	// 创建测试文件
	msg := strings.Repeat("hello", 5)
	for i := 0; i < 10; i++ {
		file := fmt.Sprintf("%d.txt", i)
		fd, err := os.Create(file)
		Nil(t, err)
		_, err = fd.WriteString(msg)
		Nil(t, err)
		stat, err := fd.Stat()
		Nil(t, err)

		hash, err := md5File(fd)
		Nil(t, err)
		files = append(files, FileInfo{
			Name:  file,
			Size:  stat.Size(),
			IsDir: stat.IsDir(),
			MD5:   hash,
		})

		err = up.Put(&PutObjectConfig{
			Path:      path.Join(remotePath, file),
			LocalPath: file,
			UseMD5:    true,
		})
		Nil(t, err)
		fd.Close()
		err = os.RemoveAll(file)
		Nil(t, err)
	}
	return files
}

func TestListObjects(t *testing.T) {
	remotePath := "/go-sdk/lb/"
	limit := 1

	files := putTestFilesToBucket(t, remotePath)
	config := &ListObjectsConfig{
		Path:         remotePath,
		MaxListTries: 0,
		DescOrder:    false,
		Iter:         "",
		Limit:        limit,
	}

	var fileInfos []*FileInfo
	for {
		fileInfo, iter, err := up.ListObjects(config)
		Nil(t, err)
		NotNil(t, fileInfo)
		NotNil(t, iter)
		t.Log(iter) // 间隔
		for _, item := range fileInfo {
			fileInfos = append(fileInfos, item)
			t.Log(item.Name)
		}
		if iter == "" {
			break
		}
		config.Iter = iter
	}

	count := len(fileInfos)
	for i := 0; i < count; i++ {
		Equal(t, files[i].Name, fileInfos[i].Name)
		Equal(t, files[i].IsDir, fileInfos[i].IsDir)
		Equal(t, files[i].Size, fileInfos[i].Size)
	}

}

func TestGetDisorderResumeProcess(t *testing.T) {
	fname := "10M"
	fd, _ := os.Create(fname)
	NotNil(t, fd)
	kb := strings.Repeat("U", 1024)
	for i := 0; i < minResumePutFileSize/1024+2; i++ {
		fd.WriteString(kb)
	}
	fileInfo, err := fd.Stat()
	Nil(t, err)
	defer fd.Close()
	defer os.RemoveAll(fname)

	path := "Disorder.txt"

	// file config
	config := &PutObjectConfig{
		Path:            path,
		Reader:          fd,
		LocalPath:       fname,
		UseMD5:          true,
		UseResumeUpload: true,
		Headers:         make(map[string]string),
	}

	// init
	result, err := up.InitMultipartUpload(&InitMultipartUploadConfig{
		Path:          path,
		PartSize:      DefaultPartSize,
		ContentType:   config.Headers["Content-Type"],
		ContentLength: fileInfo.Size(),
		OrderUpload:   false,
	})
	Nil(t, err)

	// imitate upload part failed
	testBreak := 10
	var curSize int64 = 0
	fileSize := fileInfo.Size()

	// imitate upload some part and failed after testBreak part
	for i := 0; i <= testBreak; i++ {
		fragFile, err := newFragmentFile(fd, curSize, DefaultPartSize)
		Nil(t, err)
		partSize := int64(DefaultPartSize)
		res := fileSize - curSize
		if res < DefaultPartSize {
			partSize = res
		}
		err = up.UploadPart(result, &UploadPartConfig{
			Reader:   fragFile,
			PartSize: partSize,
			PartID:   i,
		})
		Nil(t, err)

		curSize += partSize
	}
	resp, err := up.GetResumeProcess(result.Path)

	Nil(t, err)
	Equal(t, testBreak+1, len(resp.Parts))
	Equal(t, curSize, fileSize)
	err = up.CompleteMultipartUpload(result, &CompleteMultipartUploadConfig{})
	Nil(t, err)

	err = up.Delete(&DeleteObjectConfig{
		Path: path,
	})
	Nil(t, err)

}
