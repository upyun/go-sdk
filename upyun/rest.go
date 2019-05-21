package upyun

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	defaultResumePartSize = 1024 * 1024
	minResumePutFileSize  = 10 * 1024 * 1024
)

type restReqConfig struct {
	method    string
	uri       string
	query     string
	headers   map[string]string
	closeBody bool
	httpBody  io.Reader
	useMD5    bool
}

// GetObjectConfig provides a configuration to Get method.
type GetObjectConfig struct {
	Path string
	// Headers contains custom http header, like User-Agent.
	Headers   map[string]string
	LocalPath string
	Writer    io.Writer
}

// GetObjectConfig provides a configuration to List method.
type GetObjectsConfig struct {
	Path           string
	Headers        map[string]string
	ObjectsChan    chan *FileInfo
	QuitChan       chan bool
	MaxListObjects int
	MaxListTries   int
	// MaxListLevel: depth of recursion
	MaxListLevel int
	// DescOrder:  whether list objects by desc-order
	DescOrder bool

	rootDir string
	level   int
	objNum  int
	try     int
}

// RangeObjectsConfig provides a configuration to RangeList method.
type RangeObjectsConfig struct {
	StartTimestamp int64
	EndTimestamp   int64
	ObjectsChan    chan *FileInfo
	QuitChan       chan bool
	Headers        map[string]string
	MaxListTries   int
}

// PutObjectConfig provides a configuration to Put method.
type PutObjectConfig struct {
	Path            string
	LocalPath       string
	Reader          io.Reader
	Headers         map[string]string
	UseMD5          bool
	UseResumeUpload bool
	// Append Api Deprecated
	// AppendContent     bool
	ResumePartSize    int64
	MaxResumePutTries int
}

type DeleteObjectConfig struct {
	Path   string
	Async  bool
	Folder bool //optional
}

type ModifyMetadataConfig struct {
	Path      string
	Operation string
	Headers   map[string]string
}

func (up *UpYun) Usage() (n int64, err error) {
	var resp *http.Response
	resp, err = up.doRESTRequest(&restReqConfig{
		method: "GET",
		uri:    "/",
		query:  "usage",
	})

	if err == nil {
		n, err = readHTTPBodyToInt(resp)
	}

	if err != nil {
		return 0, fmt.Errorf("usage: %v", err)
	}
	return n, nil
}

func (up *UpYun) Mkdir(path string) error {
	_, err := up.doRESTRequest(&restReqConfig{
		method: "POST",
		uri:    path,
		headers: map[string]string{
			"folder":         "true",
			"x-upyun-folder": "true",
		},
		closeBody: true,
	})
	if err != nil {
		return fmt.Errorf("mkdir %s: %v", path, err)
	}
	return nil
}

// TODO: maybe directory
func (up *UpYun) Get(config *GetObjectConfig) (fInfo *FileInfo, err error) {
	if config.LocalPath != "" {
		var fd *os.File
		if fd, err = os.Create(config.LocalPath); err != nil {
			return nil, fmt.Errorf("create file: %v", err)
		}
		defer fd.Close()
		config.Writer = fd
	}

	if config.Writer == nil {
		return nil, fmt.Errorf("no writer")
	}

	resp, err := up.doRESTRequest(&restReqConfig{
		method: "GET",
		uri:    config.Path,
	})
	if err != nil {
		return nil, fmt.Errorf("doRESTRequest: %v", err)
	}
	defer resp.Body.Close()

	fInfo = parseHeaderToFileInfo(resp.Header, false)
	fInfo.Name = config.Path

	if fInfo.Size, err = io.Copy(config.Writer, resp.Body); err != nil {
		return nil, fmt.Errorf("io copy: %v", err)
	}

	return
}

func (up *UpYun) put(config *PutObjectConfig) error {
	/* Append Api Deprecated
	if config.AppendContent {
		if config.Headers == nil {
			config.Headers = make(map[string]string)
		}
		config.Headers["X-Upyun-Append"] = "true"
	}
	*/
	_, err := up.doRESTRequest(&restReqConfig{
		method:    "PUT",
		uri:       config.Path,
		headers:   config.Headers,
		closeBody: true,
		httpBody:  config.Reader,
		useMD5:    config.UseMD5,
	})
	if err != nil {
		return fmt.Errorf("doRESTRequest: %v", err)
	}
	return nil
}

// TODO: progress
func (up *UpYun) resumePut(config *PutObjectConfig) error {
	f, ok := config.Reader.(*os.File)
	if !ok {
		return fmt.Errorf("resumePut: type != *os.File")
	}

	fileinfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Stat: %v", err)
	}

	fsize := fileinfo.Size()
	if fsize < minResumePutFileSize {
		return up.put(config)
	}

	if config.ResumePartSize == 0 {
		config.ResumePartSize = defaultResumePartSize
	}
	maxPartID := int((fsize+config.ResumePartSize-1)/config.ResumePartSize - 1)

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	curSize, partSize := int64(0), config.ResumePartSize
	headers := config.Headers
	for id := 0; id <= maxPartID; id++ {
		if curSize+partSize > fsize {
			partSize = fsize - curSize
		}
		headers["Content-Length"] = fmt.Sprint(partSize)
		headers["X-Upyun-Part-ID"] = fmt.Sprint(id)

		switch id {
		case 0:
			headers["X-Upyun-Multi-Type"] = headers["Content-Type"]
			headers["X-Upyun-Multi-Length"] = fmt.Sprint(fsize)
			headers["X-Upyun-Multi-Stage"] = "initiate,upload"
		case int(maxPartID):
			headers["X-Upyun-Multi-Stage"] = "upload,complete"
			if config.UseMD5 {
				f.Seek(0, 0)
				headers["X-Upyun-Multi-MD5"], _ = md5File(f)
			}
		default:
			headers["X-Upyun-Multi-Stage"] = "upload"
		}

		fragFile, err := newFragmentFile(f, curSize, partSize)
		if err != nil {
			return fmt.Errorf("newFragmentFile: %v", err)
		}

		try := 0
		var resp *http.Response
		for ; config.MaxResumePutTries == 0 || try < config.MaxResumePutTries; try++ {
			resp, err = up.doRESTRequest(&restReqConfig{
				method:    "PUT",
				uri:       config.Path,
				headers:   headers,
				closeBody: true,
				useMD5:    config.UseMD5,
				httpBody:  fragFile,
			})
			if err == nil {
				break
			}
			if _, ok := err.(net.Error); !ok {
				return fmt.Errorf("doRESTRequest: %v", err)
			}
			fragFile.Seek(0, 0)
		}

		if config.MaxResumePutTries > 0 && try == config.MaxResumePutTries {
			return err
		}

		if id == 0 {
			headers["X-Upyun-Multi-UUID"] = resp.Header.Get("X-Upyun-Multi-UUID")
		} else {
			if id == maxPartID {
				return nil
			}
		}

		curSize += partSize
	}

	return nil
}

func (up *UpYun) Put(config *PutObjectConfig) (err error) {
	if config.LocalPath != "" {
		var fd *os.File
		if fd, err = os.Open(config.LocalPath); err != nil {
			return fmt.Errorf("open file: %v", err)
		}
		defer fd.Close()
		config.Reader = fd
	}

	if config.UseResumeUpload {
		return up.resumePut(config)
	}
	return up.put(config)
}

func (up *UpYun) Delete(config *DeleteObjectConfig) error {
	headers := map[string]string{}
	if config.Async == true {
		headers["x-upyun-async"] = "true"
	}
	if config.Folder {
		headers["x-upyun-folder"] = "true"
	}
	_, err := up.doRESTRequest(&restReqConfig{
		method:    "DELETE",
		uri:       config.Path,
		headers:   headers,
		closeBody: true,
	})
	if err != nil {
		if e, ok := err.(Error); ok {
			e.error = fmt.Errorf("delete %s: %v", config.Path, err)
			return e
		}
		return fmt.Errorf("delete %s: %v", config.Path, err)
	}
	return nil
}

func (up *UpYun) GetInfo(path string) (*FileInfo, error) {
	resp, err := up.doRESTRequest(&restReqConfig{
		method:    "HEAD",
		uri:       path,
		closeBody: true,
	})
	if err != nil {
		if e, ok := err.(Error); ok {
			e.error = fmt.Errorf("getinfo %s: %v", path, err)
			return nil, e
		}
		return nil, fmt.Errorf("getinfo %s: %v", path, err)
	}
	fInfo := parseHeaderToFileInfo(resp.Header, true)
	fInfo.Name = path
	return fInfo, nil
}

func (up *UpYun) List(config *GetObjectsConfig) error {
	if config.ObjectsChan == nil {
		return fmt.Errorf("ObjectsChan == nil")
	}
	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}
	if config.QuitChan == nil {
		config.QuitChan = make(chan bool)
	}
	// 50 is nice value
	if _, exist := config.Headers["X-List-Limit"]; !exist {
		config.Headers["X-List-Limit"] = "50"
	}

	if config.DescOrder {
		config.Headers["X-List-Order"] = "desc"
	}

	config.Headers["X-UpYun-Folder"] = "true"

	// 1st level
	if config.level == 0 {
		defer close(config.ObjectsChan)
	}

	for {
		resp, err := up.doRESTRequest(&restReqConfig{
			method:  "GET",
			uri:     config.Path,
			headers: config.Headers,
		})

		if err != nil {
			if _, ok := err.(net.Error); ok {
				config.try++
				if config.MaxListTries == 0 || config.try < config.MaxListTries {
					continue
				}
			}
			return err
		}

		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("ioutil ReadAll: %v", err)
		}

		for _, fInfo := range parseBodyToFileInfos(b) {
			if fInfo.IsDir && (config.level+1 < config.MaxListLevel || config.MaxListLevel == -1) {
				rConfig := &GetObjectsConfig{
					Path:           path.Join(config.Path, fInfo.Name),
					QuitChan:       config.QuitChan,
					ObjectsChan:    config.ObjectsChan,
					MaxListTries:   config.MaxListTries,
					MaxListObjects: config.MaxListObjects,
					DescOrder:      config.DescOrder,
					MaxListLevel:   config.MaxListLevel,
					level:          config.level + 1,
					rootDir:        path.Join(config.rootDir, fInfo.Name),
					try:            config.try,
					objNum:         config.objNum,
				}
				if err = up.List(rConfig); err != nil {
					return err
				}
				// empty folder
				if config.objNum == rConfig.objNum {
					fInfo.IsEmptyDir = true
				}
				config.try, config.objNum = rConfig.try, rConfig.objNum
			}
			if config.rootDir != "" {
				fInfo.Name = path.Join(config.rootDir, fInfo.Name)
			}

			select {
			case <-config.QuitChan:
				return nil
			default:
				config.ObjectsChan <- fInfo
			}

			config.objNum++
			if config.MaxListObjects > 0 && config.objNum >= config.MaxListObjects {
				return nil
			}

		}

		config.Headers["X-List-Iter"] = resp.Header.Get("X-Upyun-List-Iter")
		if config.Headers["X-List-Iter"] == "g2gCZAAEbmV4dGQAA2VvZg" {
			return nil
		}
	}
}

func (up *UpYun) RangeList(config *RangeObjectsConfig) error {
	if config.ObjectsChan == nil {
		return fmt.Errorf("ObjectsChan == nil")
	}
	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}
	if config.QuitChan == nil {
		config.QuitChan = make(chan bool)
	}
	if _, exist := config.Headers["X-List-Limit"]; !exist {
		config.Headers["X-List-Limit"] = "50"
	}
	if config.StartTimestamp != 0 {
		config.Headers["X-List-Start"] = fmt.Sprint(config.StartTimestamp)
	}
	if config.EndTimestamp != 0 {
		config.Headers["X-List-End"] = fmt.Sprint(config.EndTimestamp)
	}
	defer close(config.ObjectsChan)

	try := 0
	for {
		resp, err := up.doRESTRequest(&restReqConfig{
			method:  "GET",
			uri:     "/?files",
			headers: config.Headers,
		})
		if err != nil {
			if _, ok := err.(net.Error); ok {
				try++
				if config.MaxListTries == 0 || try < config.MaxListTries {
					continue
				}
			}
			return err
		}

		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("ioutil ReadAll: %v", err)
		}

		for _, fInfo := range parseRangeListToFileInfos(b) {
			select {
			case config.ObjectsChan <- fInfo:
			case <-config.QuitChan:
				return nil
			}
		}

		config.Headers["X-List-Iter"] = resp.Header.Get("X-Upyun-List-Iter")
		if config.Headers["X-List-Iter"] == "g2gCZAAEbmV4dGQAA2VvZg" {
			return nil
		}
	}
}

func (up *UpYun) ModifyMetadata(config *ModifyMetadataConfig) error {
	if config.Operation == "" {
		config.Operation = "merge"
	}
	_, err := up.doRESTRequest(&restReqConfig{
		method:    "PATCH",
		uri:       config.Path,
		query:     "metadata=" + config.Operation,
		headers:   config.Headers,
		closeBody: true,
	})
	return err
}

func (up *UpYun) doRESTRequest(config *restReqConfig) (*http.Response, error) {
	escUri := path.Join("/", up.Bucket, escapeUri(config.uri))
	if strings.HasSuffix(config.uri, "/") {
		escUri += "/"
	}
	if config.query != "" {
		escUri += "?" + config.query
	}

	headers := map[string]string{}
	hasMD5 := false
	for k, v := range config.headers {
		if strings.ToLower(k) == "content-md5" && v != "" {
			hasMD5 = true
		}
		headers[k] = v
	}

	headers["Date"] = makeRFC1123Date(time.Now())
	headers["Host"] = "v0.api.upyun.com"

	if !hasMD5 && config.useMD5 {
		switch v := config.httpBody.(type) {
		case *os.File:
			headers["Content-MD5"], _ = md5File(v)
		case UpYunPutReader:
			headers["Content-MD5"] = v.MD5()
		}
	}

	if up.deprecated {
		if _, ok := headers["Content-Length"]; !ok {
			size := int64(0)
			switch v := config.httpBody.(type) {
			case *os.File:
				if fInfo, err := v.Stat(); err == nil {
					size = fInfo.Size()
				}
			case UpYunPutReader:
				size = int64(v.Len())
			}
			headers["Content-Length"] = fmt.Sprint(size)
		}
		headers["Authorization"] = up.MakeRESTAuth(&RESTAuthConfig{
			Method:    config.method,
			Uri:       escUri,
			DateStr:   headers["Date"],
			LengthStr: headers["Content-Length"],
		})
	} else {
		headers["Authorization"] = up.MakeUnifiedAuth(&UnifiedAuthConfig{
			Method:     config.method,
			Uri:        escUri,
			DateStr:    headers["Date"],
			ContentMD5: headers["Content-MD5"],
		})
	}

	endpoint := up.doGetEndpoint("v0.api.upyun.com")
	url := fmt.Sprintf("http://%s%s", endpoint, escUri)

	resp, err := up.doHTTPRequest(config.method, url, headers, config.httpBody)
	if err != nil {
		// Don't modify net error
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, Error{
			fmt.Errorf("%s %d %s", config.method, resp.StatusCode, string(body)),
			resp.StatusCode,
		}
	}

	if config.closeBody {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}

	return resp, nil
}
