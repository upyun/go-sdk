# upyun
--
    import "github.com/polym/go-sdk/upyun"

package upyun is used for your UPYUN bucket this sdk implement purge api, form
api, http rest api

## Example

examples are in upyun/example/

## Usage

```go
const (
	Auto    = "v0.api.upyun.com"
	Telecom = "v1.api.upyun.com"
	Cnc     = "v2.api.upyun.com"
	Ctt     = "v3.api.upyun.com"
)
```
Auto: Auto detected, based on user's internet Telecom: (ISP) China Telecom Cnc:
(ISP) China Unicom Ctt: (ISP) China Tietong purgeEndpoint: endpoint used for
purging Default(Min/Max)ChunkSize: set the buffer size when doing copy operation
defaultConnectTimeout: connection timeout when connect to upyun endpoint

```go
const (
	Version = "1.2.0"
)
```

#### func  SetChunkSize

```go
func SetChunkSize(chunksize int)
```

#### type FileInfo

```go
type FileInfo struct {
	Type string
	Date string
	Size int64
}
```

FileInfo when HEAD file

#### type FormAPIResp

```go
type FormAPIResp struct {
	Code      int    `json:"code"`
	Msg       string `json:"message"`
	Url       string `json:"url"`
	Timestamp int64  `json:"time"`
	ImgWidth  int    `json:"image-width"`
	ImgHeight int    `json:"image-height"`
	ImgFrames int    `json:"image-frames"`
	ImgType   string `json:"image-type"`
	Sign      string `json:"sign"`
}
```

Response from UPYUN Form API Server

#### type Info

```go
type Info struct {
	Size int64
	Time int64
	Name string
	Type string
}
```

FileInfo when use getlist

#### type MediaStatusResp

```go
type MediaStatusResp struct {
	Tasks map[string]interface{} `json:"tasks"`
}
```

status response

#### type MergeResp

```go
type MergeResp struct {
	Path          string      `json:"path"`
	ContentType   string      `json:"mimetype"`
	ContentLength interface{} `json:"file_size"`
	LastModify    int64       `json:"last_modified"`
	Signature     string      `json:"signature"`
	ImageWidth    int         `json:"image_width"`
	ImageHeight   int         `json:"image_height"`
	ImageFrames   int         `json:"image_frames"`
}
```

merge response body

#### type ReqError

```go
type ReqError struct {
	Headers http.Header
}
```

Request Error

#### func (*ReqError) Error

```go
func (r *ReqError) Error() string
```

#### type UpYun

```go
type UpYun struct {
	Bucket   string
	Username string
	Passwd   string

	ChunkSize int
}
```

UPYUN REST API Client

#### func  NewUpYun

```go
func NewUpYun(bucket, username, passwd string) *UpYun
```
NewUpYun return a new UPYUN REST API client given a bucket name, username,
password. As Default, endpoint is set to Auto, http client connection timeout is
set to defalutConnectionTimeout which is equal to 60 seconds.

#### func (*UpYun) Delete

```go
func (u *UpYun) Delete(key string) error
```
Delete deletes the specified **file** in UPYUN File System.

#### func (*UpYun) Get

```go
func (u *UpYun) Get(key string, value io.Writer) error
```
Get gets the specified file in UPYUN File System

#### func (*UpYun) GetInfo

```go
func (u *UpYun) GetInfo(key string) (FileInfo, error)
```
GetInfo gets information of item in UPYUN File System

#### func (*UpYun) GetList

```go
func (u *UpYun) GetList(key string) ([]Info, error)
```
GetList lists items in key. The number of items must be less then 100

#### func (*UpYun) LoopList

```go
func (u *UpYun) LoopList(key, iter string, limit int) ([]Info, string, error)
```
LoopList list items iteratively.

#### func (*UpYun) Mkdir

```go
func (u *UpYun) Mkdir(key string) error
```
Mkdir creates a directory in UPYUN File System

#### func (*UpYun) Purge

```go
func (u *UpYun) Purge(urls []string) (string, error)
```
Purge post a purge request to UPYUN Purge Server

#### func (*UpYun) Put

```go
func (u *UpYun) Put(key string, value io.Reader, useMD5 bool, secret, contentType string,
	headers map[string]string) (http.Header, error)
```
Put uploads filelike object to UPYUN File System

#### func (*UpYun) SetEndpoint

```go
func (u *UpYun) SetEndpoint(endpoint string) (string, error)
```
SetEndpoint sets the request endpoint to UPYUN REST Server.

#### func (*UpYun) SetTimeout

```go
func (core *UpYun) SetTimeout(timeout int)
```
Set connect timeout

#### func (*UpYun) Usage

```go
func (u *UpYun) Usage() (int64, error)
```
Usage gets the usage of the bucket in UPYUN File System

#### type UpYunForm

```go
type UpYunForm struct {
	Key    string
	Bucket string
}
```

UPYUN HTTP FORM API Client

#### func  NewUpYunForm

```go
func NewUpYunForm(bucket, key string) *UpYunForm
```
NewUpYunForm return a UPYUN Form API client given a form api key and bucket
name. As Default, endpoint is set to Auto, http client connection timeout is set
to defalutConnectionTimeout which is equal to 60 seconds.

#### func (*UpYunForm) Put

```go
func (uf *UpYunForm) Put(fpath, saveas string, expireAfter int64,
	options map[string]string) (*FormAPIResp, error)
```
Put posts a http form request given reader, save path, expiration, other options
and returns a FormAPIResp pointer.

#### func (*UpYunForm) SetTimeout

```go
func (core *UpYunForm) SetTimeout(timeout int)
```
Set connect timeout

#### type UpYunMedia

```go
type UpYunMedia struct {
	Username string
	Passwd   string
	Bucket   string
}
```

UPYUN MEDIA API

#### func  NewUpYunMedia

```go
func NewUpYunMedia(bucket, user, pass string) *UpYunMedia
```

#### func (*UpYunMedia) GetProgress

```go
func (upm *UpYunMedia) GetProgress(task_ids string) (*MediaStatusResp, error)
```
Get Task Progress

#### func (*UpYunMedia) PostTasks

```go
func (upm *UpYunMedia) PostTasks(src, notify string,
	tasks []map[string]interface{}) ([]string, error)
```
Send Media Tasks Reqeust

#### func (*UpYunMedia) SetTimeout

```go
func (core *UpYunMedia) SetTimeout(timeout int)
```
Set connect timeout

#### type UpYunMultiPart

```go
type UpYunMultiPart struct {
	Bucket    string
	APIKey    string
	BlockSize int64
}
```

UPYUN MultiPart Upload API

#### func  NewUpYunMultiPart

```go
func NewUpYunMultiPart(bucket, apikey string, blocksize int64) *UpYunMultiPart
```
NewUpYunMultiPart returns a new UPYUN Multipart Upload API client given bucket
name, form api key and blocksize.

#### func (*UpYunMultiPart) InitUpload

```go
func (ump *UpYunMultiPart) InitUpload(key string, value *os.File,
	expire int64, options map[string]interface{}) ([]byte, error)
```
InitUpload initalizes a multipart upload request

#### func (*UpYunMultiPart) MergeBlock

```go
func (ump *UpYunMultiPart) MergeBlock(saveToken, secret string,
	expire int64) ([]byte, error)
```
MergeBlock posts a merge request to merge all blocks uploaded

#### func (*UpYunMultiPart) Put

```go
func (ump *UpYunMultiPart) Put(key, fpath string,
	expire int64, options map[string]interface{}) (*MergeResp, error)
```
Put uploads a file through UPYUN MultiPart Upload API

#### func (*UpYunMultiPart) SetTimeout

```go
func (core *UpYunMultiPart) SetTimeout(timeout int)
```
Set connect timeout

#### func (*UpYunMultiPart) UploadBlock

```go
func (ump *UpYunMultiPart) UploadBlock(fd *os.File, bindex int, expire int64,
	fpath, saveToken, secret string) ([]byte, error)
```
UploadBlock uploads a block

#### type UploadResp

```go
type UploadResp struct {
	// returns after init request
	SaveToken string `json:"save_token"`
	// token_secert is equal to UPYUN Form API KEY
	Secret string `json:"token_secret"`
	// UPYUN Bucket Name
	Bucket string `json:"bucket_name"`
	// Number of Blocks
	Blocks   int   `json:"blocks"`
	Status   []int `json:"status"`
	ExpireAt int64 `json:"expire_at"`
}
```

upload response body
