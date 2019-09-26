package upyun

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultPartSize      = 1024 * 1024
	MaxPartNum           = 10000
	minResumePutFileSize = 10 * 1024 * 1024
)

type restReqConfig struct {
	method    string
	uri       string
	query     string
	headers   map[string]string
	closeBody bool
	httpBody  io.Reader
	useMD5    bool
	listener  ProgressListener
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
	Listener          ProgressListener
}

//UploadFileConfig is multipart file upload config
type UploadFileConfig struct {
	Path          string
	LocalPath     string
	PartSize      int64
	Listener      ProgressListener
	Parallel      int
	CheckPointDir string
}
type UploadPartConfig struct {
	Reader   io.Reader
	PartSize int64
	PartID   int
}
type InitMultipartUploadConfig struct {
	Path          string
	PartSize      int64
	ContentLength int64 //optional
	ContentType   string
}
type InitMultipartUploadResult struct {
	UploadID string
	Path     string
	PartSize int64
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

type ListMultipartConfig struct {
	Prefix string
	Limit  int64
}
type ListMultipartPartsConfig struct {
	BeginID int
}
type MultipartUploadFile struct {
	Key       string `json:"key"`
	UUID      string `json:"uuid"`
	Completed bool   `json:"completed"`
	CreatedAt int64  `json:"created_at"`
}
type ListMultipartUploadResult struct {
	Files []*MultipartUploadFile `json:"files"`
}
type MultipartUploadedPart struct {
	Etag string `json:"etag"`
	Size int64  `json:"size"`
	Id   int    `json:"id"`
}
type ListUploadedPartsResult struct {
	Parts []*MultipartUploadedPart `json:"parts"`
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
		listener:  config.Listener,
	})
	if err != nil {
		return fmt.Errorf("doRESTRequest: %v", err)
	}
	return nil
}

func getPartInfo(partSize, fsize int64) (int64, int64, error) {
	if partSize <= 0 {
		partSize = DefaultPartSize
	}
	if partSize < DefaultPartSize {
		return 0, 0, fmt.Errorf("The minimum of part size is %d", DefaultPartSize)
	}
	if partSize%DefaultPartSize != 0 {
		return 0, 0, fmt.Errorf("The part size is a multiple of %d", DefaultPartSize)
	}

	partNum := (fsize + partSize - 1) / partSize
	if partNum > MaxPartNum {
		return 0, 0, fmt.Errorf("The maximum part number is  %d", MaxPartNum)
	}
	return partSize, partNum, nil
}
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

	partSize, partNum, err := getPartInfo(config.ResumePartSize, fsize)
	if err != nil {
		return err
	}
	maxPartID := int(partNum) - 1

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	curSize := int64(0)
	headers := config.Headers
	listener := config.Listener
	event := newProgressEvent(TransferStartedEvent, 0, fsize)
	publishProgress(listener, event)

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
				event = newProgressEvent(TransferFailedEvent, curSize, fsize)
				publishProgress(listener, event)
				return fmt.Errorf("doRESTRequest: %v", err)
			}
			fragFile.Seek(0, 0)
		}

		if config.MaxResumePutTries > 0 && try == config.MaxResumePutTries {
			event = newProgressEvent(TransferFailedEvent, curSize, fsize)
			publishProgress(listener, event)
			return err
		}

		curSize += partSize
		event = newProgressEvent(TransferDataEvent, curSize, fsize)
		publishProgress(listener, event)
		if id == 0 {
			headers["X-Upyun-Multi-UUID"] = resp.Header.Get("X-Upyun-Multi-UUID")
		} else {
			if id == maxPartID {
				break
			}
		}
	}

	event = newProgressEvent(TransferCompletedEvent, curSize, fsize)
	publishProgress(listener, event)
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

func (up *UpYun) InitMultipartUpload(config *InitMultipartUploadConfig) (*InitMultipartUploadResult, error) {
	partSize, _, err := getPartInfo(config.PartSize, config.ContentLength)
	if err != nil {
		return nil, err
	}
	headers := make(map[string]string)
	headers["X-Upyun-Multi-Type"] = config.ContentType
	if config.ContentLength > 0 {
		headers["X-Upyun-Multi-Length"] = strconv.FormatInt(config.ContentLength, 10)
	}
	headers["X-Upyun-Multi-Stage"] = "initiate"
	headers["X-Upyun-Multi-Disorder"] = "true"
	headers["X-Upyun-Multi-Part-Size"] = strconv.FormatInt(partSize, 10)
	resp, err := up.doRESTRequest(&restReqConfig{
		method:    "PUT",
		uri:       config.Path,
		headers:   headers,
		closeBody: true,
	})
	if err != nil {
		return nil, err
	}
	return &InitMultipartUploadResult{
		UploadID: resp.Header.Get("X-Upyun-Multi-Uuid"),
		Path:     config.Path,
		PartSize: partSize,
	}, err
}
func (up *UpYun) UploadPart(initResult *InitMultipartUploadResult, part *UploadPartConfig) error {
	headers := make(map[string]string)
	headers["X-Upyun-Multi-Stage"] = "upload"
	headers["X-Upyun-Multi-Uuid"] = initResult.UploadID
	headers["X-Upyun-Part-Id"] = fmt.Sprint(part.PartID)
	headers["Content-Length"] = fmt.Sprint(part.PartSize)

	_, err := up.doRESTRequest(&restReqConfig{
		method:    "PUT",
		uri:       initResult.Path,
		headers:   headers,
		closeBody: true,
		useMD5:    false,
		httpBody:  part.Reader,
	})
	if err != nil {
		return fmt.Errorf("doRESTRequest: %v", err)
	}
	return nil
}
func (up *UpYun) ListMultipartUploads(config *ListMultipartConfig) (*ListMultipartUploadResult, error) {
	headers := make(map[string]string)
	headers["X-Upyun-List-Type"] = "multi"
	if config.Prefix != "" {
		headers["X-Upyun-List-Prefix"] = base64.StdEncoding.EncodeToString([]byte(config.Prefix))
	}
	if config.Limit > 0 {
		headers["X-Upyun-List-Limit"] = strconv.FormatInt(config.Limit, 10)
	}

	res, err := up.doRESTRequest(&restReqConfig{
		method:    "GET",
		headers:   headers,
		uri:       "/",
		closeBody: false,
		useMD5:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("ListMultipartUploads: %v", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ListMultipartUploads reader body: %v", err)
	}

	result := &ListMultipartUploadResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, fmt.Errorf("ListMultipartUploads json unmarshal: %v", err)
	}
	return result, nil
}

func (up *UpYun) ListMultipartParts(intiResult *InitMultipartUploadResult, config *ListMultipartPartsConfig) (*ListUploadedPartsResult, error) {
	headers := make(map[string]string)
	headers["X-Upyun-Multi-Uuid"] = intiResult.UploadID

	if config.BeginID > 0 {
		headers["X-Upyun-Part-Id"] = fmt.Sprint(config.BeginID)
	}
	res, err := up.doRESTRequest(&restReqConfig{
		method:    "GET",
		headers:   headers,
		uri:       intiResult.Path,
		closeBody: false,
		useMD5:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("ListUploadedParts: %v", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ListUploadedParts reader body: %v", err)
	}

	result := &ListUploadedPartsResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, fmt.Errorf("ListUploadedParts json unmarshal: %v", err)
	}
	return result, nil
}
func (up *UpYun) UploadPartFromFile(initResult *InitMultipartUploadResult, filePath string, startPosition, partSize int64, partID int) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	fd.Seek(startPosition, os.SEEK_SET)
	headers := make(map[string]string)
	headers["X-Upyun-Multi-Stage"] = "upload"
	headers["X-Upyun-Multi-Uuid"] = initResult.UploadID
	headers["X-Upyun-Part-Id"] = fmt.Sprint(partID)
	headers["Content-Length"] = fmt.Sprint(partSize)

	_, err = up.doRESTRequest(&restReqConfig{
		method:    "PUT",
		uri:       initResult.Path,
		headers:   headers,
		closeBody: true,
		useMD5:    false,
		httpBody:  &io.LimitedReader{R: fd, N: partSize},
	})
	if err != nil {
		return fmt.Errorf("doRESTRequest: %v", err)
	}
	return nil
}
func (up *UpYun) CompleteMultipartUpload(initResult *InitMultipartUploadResult) error {
	headers := make(map[string]string)
	headers["X-Upyun-Multi-Stage"] = "complete"
	headers["X-Upyun-Multi-Uuid"] = initResult.UploadID
	_, err := up.doRESTRequest(&restReqConfig{
		method:  "PUT",
		uri:     initResult.Path,
		headers: headers,
	})
	if err != nil {
		return fmt.Errorf("doRESTRequest: %v", err)
	}
	return nil
}
func (up *UpYun) UploadFile(config *UploadFileConfig) error {
	if config.LocalPath == "" {
		return fmt.Errorf("resumePut: should using local path")
	}

	var err error
	config.LocalPath, err = filepath.Abs(config.LocalPath)
	if err != nil {
		return err
	}

	filePath := config.LocalPath
	cp, err := newCheckpoint(filePath, config.Path, config.CheckPointDir)
	if err != nil {
		return err
	}
	var initResult *InitMultipartUploadResult
	if cp.UploadID == "" { //should init upload
		initResult, err = up.InitMultipartUpload(&InitMultipartUploadConfig{
			Path:          config.Path,
			ContentLength: cp.FileStat.Size,
			PartSize:      config.PartSize,
		})
		if err != nil {
			return err
		}

		err := cp.init(initResult.PartSize)
		if err != nil {
			return err
		}
		cp.UploadID = initResult.UploadID
	} else {
		initResult = &InitMultipartUploadResult{
			Path:     config.Path,
			UploadID: cp.UploadID,
		}
	}

	workers := config.Parallel
	if workers == 0 {
		workers = 4
	}
	jobChan := make(chan *uploadPart, workers*2)
	results := make(chan *uploadPart, workers*2)
	failed := make(chan error, 1)
	dieCh := make(chan struct{})

	for i := 0; i < workers; i++ {
		go func() {
			for part := range jobChan {
				err := up.UploadPartFromFile(initResult, filePath, part.Offset, part.PartSize, part.PartID)
				if err != nil {
					select {
					case <-dieCh:
					case failed <- err:
					}
					return
				}
				select {
				case <-dieCh:
					return
				case results <- part:
				}
			}
		}()
	}

	todoParts := cp.todoParts()
	completedBytes := cp.getCompletedBytes()
	listener := config.Listener
	event := newProgressEvent(TransferStartedEvent, completedBytes, cp.FileStat.Size)
	publishProgress(listener, event)
	go func() {
		defer close(jobChan)
		for _, part := range todoParts {
			select {
			case jobChan <- part:
			case <-dieCh:
				return
			}
		}
	}()

	for range todoParts {
		select {
		case r := <-results:
			r.Completed = true
			cp.completePart(r.PartID)
			cp.dump()
			completedBytes += r.PartSize
			event = newProgressEvent(TransferDataEvent, completedBytes, cp.FileStat.Size)
			publishProgress(listener, event)
		case err := <-failed:
			close(dieCh)
			event = newProgressEvent(TransferFailedEvent, completedBytes, cp.FileStat.Size)
			publishProgress(listener, event)
			return err
		}
	}

	event = newProgressEvent(TransferCompletedEvent, completedBytes, cp.FileStat.Size)
	publishProgress(listener, event)

	err = up.CompleteMultipartUpload(initResult)
	if err != nil {
		return err
	}
	cp.remove()
	return nil
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

	resp, err := up.doHTTPRequest(config.method, url, headers, config.httpBody, config.listener)
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
