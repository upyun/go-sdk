package upyun

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	ROOT       = MakeTmpPath()
	NOTIFY_URL = os.Getenv("UPYUN_NOTIFY")
	TEMP_DIR   = "./.temp"
)

var up = NewUpYun(&UpYunConfig{
	Bucket:   os.Getenv("UPYUN_BUCKET"),
	Operator: os.Getenv("UPYUN_USERNAME"),
	Password: os.Getenv("UPYUN_PASSWORD"),
	Secret:   os.Getenv("UPYUN_SECRET"),
})

func MakeTmpPath() string {
	return "/go-sdk/" + fmt.Sprint(time.Now().UnixNano())
}
func TempKey(t *testing.T) string {
	return path.Join(ROOT, fmt.Sprint(time.Now().UnixNano()))
}
func TempLocalFile(t *testing.T) string {
	name := "go-sdk" + "-" + fmt.Sprint(time.Now().UnixNano())
	name = strings.Replace(name, "/", "_", -1)
	d := path.Join(TEMP_DIR, name)
	os.MkdirAll(TEMP_DIR, 0755)
	return d
}
func TempLocalDir(t *testing.T) string {
	name := "go-sdk" + "-" + fmt.Sprint(time.Now().UnixNano())
	name = strings.Replace(name, "/", "_", -1)
	d := path.Join(TEMP_DIR, name)
	os.MkdirAll(d, 0755)
	return d
}

func Equal(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\tnexp: %#v\n\n\tgot:  %#v\033[39m\n\n",
			filepath.Base(file), line, expected, actual)
		t.FailNow()
	}
}

func NotEqual(t *testing.T, actual, expected interface{}) {
	if reflect.DeepEqual(actual, expected) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\tnexp: %#v\n\n\tgot:  %#v\033[39m\n\n",
			filepath.Base(file), line, expected, actual)
		t.FailNow()
	}
}

func Nil(t *testing.T, object interface{}) {
	if !isNil(object) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\t   <nil> (expected)\n\n\t!= %+v (actual)\033[39m\n\n",
			filepath.Base(file), line, object)
		t.FailNow()
	}
}

func NotNil(t *testing.T, object interface{}) {
	if isNil(object) {
		_, file, line, _ := runtime.Caller(1)
		t.Logf("\033[31m%s:%d:\n\n\tExpected value not to be <nil>\033[39m\n\n",
			filepath.Base(file), line)
		t.FailNow()
	}
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false

}

func computeMD5(filePath string, resume bool) (string, error) {
	hash := md5.New()
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	if resume {
		stat, err := f.Stat()
		if err != nil {
			return "", err
		}
		size := stat.Size()
		count := (size + DefaultPartSize - 1) / DefaultPartSize
		for i := int64(0); i < count; i++ {
			innerHash := md5.New()
			if _, err := io.Copy(innerHash, io.LimitReader(f, DefaultPartSize)); err != nil {
				return "", err
			}
			hash.Write([]byte(fmt.Sprintf("%x", innerHash.Sum(nil))))
		}
	} else {
		if _, err := io.Copy(hash, f); err != nil {
			return "", err
		}
	}
	hash.Sum(nil)
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func TestMain(m *testing.M) {
	_, err := up.Usage()
	if err != nil {
		fmt.Println("failed to login. Have set UPYUN_BUCKET UPYUN_USERNAME UPYUN_PASSWORD UPYUN_SECRET UPYUN_NOTIFY?", err)
		os.Exit(-1)
	}
	clean := func() {
		objs := make(chan *FileInfo, 20)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for obj := range objs {
				up.Delete(&DeleteObjectConfig{
					Path: path.Join(ROOT, obj.Name),
				})
			}
			up.Delete(&DeleteObjectConfig{
				Path: ROOT,
			})
			wg.Done()
		}()

		up.List(&GetObjectsConfig{
			Path:         ROOT,
			ObjectsChan:  objs,
			MaxListLevel: -1,
		})
		wg.Wait()

		if _, err := up.GetInfo(ROOT); err == nil {
			fmt.Println("Not cleanup")
			os.Exit(-1)
		}
	}

	flag.Parse()
	code := m.Run()

	clean()
	os.Exit(code)
}
