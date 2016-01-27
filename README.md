# UPYUN Go SDK

[![Build Status](https://travis-ci.org/upyun/go-sdk.svg?branch=master)](https://travis-ci.org/upyun/go-sdk)

    import "github.com/upyun/go-sdk/upyun"

UPYUN Go SDK, 集成：
- [UPYUN HTTP REST 接口](http://docs.upyun.com/api/rest_api/)
- [UPYUN HTTP FORM 接口](http://docs.upyun.com/api/form_api/)
- [UPYUN 缓存刷新接口](http://docs.upyun.com/api/purge/)
- [UPYUN 分块上传接口](http://docs.upyun.com/api/multipart_upload/)
- [UPYUN 视频处理接口](http://docs.upyun.com/api/av_pretreatment/)

Table of Contents
=================

  * [UPYUN Go SDK](#upyun-go-sdk)
    * [Examples](#examples)
    * [Projects using this SDK](#projects-using-this-sdk)
    * [Usage](#usage)
      * [UPYUN HTTP REST 接口](#upyun-http-rest-接口)
        * [UpYun](#upyun)
        * [初始化 UpYun](#初始化-upyun)
        * [设置 API 访问域名](#设置-api-访问域名)
        * [获取空间存储使用量](#获取空间存储使用量)
        * [创建目录](#创建目录)
        * [上传](#上传)
        * [下载](#下载)
        * [删除](#删除)
        * [获取文件信息](#获取文件信息)
        * [获取文件列表](#获取文件列表)
      * [UPYUN 缓存刷新接口](#upyun-缓存刷新接口)
      * [UPYUN HTTP 表单上传接口](#upyun-http-表单上传接口)
        * [UpYunForm](#upyunform)
        * [初始化 UpYunForm](#初始化-upyunform)
        * [FormAPIResp](#formapiresp)
        * [设置 API 访问域名](#设置-api-访问域名-1)
        * [上传文件](#上传文件)
      * [UPYUN 分块上传接口](#upyun-分块上传接口)
        * [UpYunMultiPart](#upyunmultipart)
        * [UploadResp](#uploadresp)
        * [MergeResp](#mergeresp)
        * [初始化 UpYunMultiPart](#初始化-upyunmultipart)
        * [上传](#上传-1)
      * [UPYUN 音视频处理接口](#upyun-音视频处理接口)
        * [UpYunMedia](#upyunmedia)
        * [MediaStatusResp](#mediastatusresp)
        * [初始化 UpYunMedia](#初始化-upyunmedia)
        * [提交任务](#提交任务)
        * [查询进度](#查询进度)

## Examples

示例代码见 `examples/`。

## Projects using this SDK

- [UPYUN Command Tool](https://github.com/polym/upx) by [polym](https://github.com/polym)

## Usage

### UPYUN HTTP REST 接口

#### UpYun

```go
type UpYun struct {
    Bucket    string    // 空间名（即服务名称）
    Username  string    // 操作员
    Passwd    string    // 密码
    ChunkSize int       // 块读取大小, 默认32KB
}
```

#### 初始化 UpYun

```go
func NewUpYun(bucket, username, passwd string) *UpYun
```

#### 设置 API 访问域名

```go
// Auto: Auto detected, based on user's internet
// Telecom: (ISP) China Telecom
// Cnc:     (ISP) China Unicom
// Ctt:     (ISP) China Tietong
const (
    Auto = iota
    Telecom
    Cnc
    Ctt
)

func (u *UpYun) SetEndpoint(ed int) error
```

#### 获取空间存储使用量

```go
func (u *UpYun) Usage() (int64, error)
```

#### 创建目录

```go
func (u *UpYun) Mkdir(key string) error
```

#### 上传

```go
func (u *UpYun) Put(key string, value io.Reader, useMD5 bool,
        headers map[string]string) (http.Header, error)
```

`key` 为 UPYUN 上的存储路径，`value` 既可以是文件，也可以是 `buffer`，`useMD5` 是否 MD5 校验，`headers` 自定义上传参数，除 [上传参数](https://docs.upyun.com/api/rest_api/#_4)，还可以设置 `Content-Length`，支持流式上传。流式上传需要指定 `Contnet-Length`，如需 MD5 校验，需要设置 `Content-MD5`。

#### 下载

```go
func (u *UpYun) Get(key string, value io.Writer) (int, error)
```

此方法返回文件大小

#### 删除

```go
func (u *UpYun) Delete(key string) error
```

#### 获取文件信息

```go
type FileInfo struct {
    Size int64         // 文件大小
    Time time.Time     // 修改时间
    Name string        // 文件名
    Type string        // 类型，folder 或者 file
}

func (u *UpYun) GetInfo(key string) (*FileInfo, error)
```

#### 获取文件列表

```go
// 少量文件
func (u *UpYun) GetList(key string) ([]*FileInfo, error)

// 大量文件
func (u *UpYun) GetLargeList(key string, recursive bool) chan *FileInfo
```

`key` 必须为目录。对于目录下有大量文件的，建议使用 `GetLargeList`。

---

### UPYUN 缓存刷新接口

```go
func (u *UpYun) Purge(urls []string) (string, error)
```

---

### UPYUN HTTP 表单上传接口

#### UpYunForm

```go
type UpYunForm struct {
    Secret    string    // 表单密钥
    Bucket    string    // 空间名（即服务名称）
}
```

#### 初始化 UpYunForm

```go
func NewUpYunForm(bucket, key string) *UpYunForm
```

#### FormAPIResp

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

#### 设置 API 访问域名

```go
func (u *UpYunForm) SetEndpoint(ed int) error
```

#### 上传文件

```go
func (uf *UpYunForm) Put(fpath, saveas string, expireAfter int64,
    options map[string]string) (*FormAPIResp, error)
```

`fpath` 上传文件名，`saveas` UPYUN 存储保存路径，`expireAfter` 过期时间长度，`options` 上传参数。

---

### UPYUN 分块上传接口

#### UpYunMultiPart

```go
type UpYunMultiPart struct {
    Bucket    string        // 空间名（即服务名称）
    Secret    string        // 表单密钥
    BlockSize int64         // 分块大小，单位字节, 建议 1024000
}
```

#### UploadResp

```go
type UploadResp struct {
    // returns after init request
    SaveToken string `json:"save_token"`
    // token_secert is equal to UPYUN Form API Secret
    Secret string `json:"token_secret"`
    // UPYUN Bucket Name
    Bucket string `json:"bucket_name"`
    // Number of Blocks
    Blocks   int   `json:"blocks"`
    Status   []int `json:"status"`
    ExpireAt int64 `json:"expire_at"`
}
```

#### MergeResp

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

#### 初始化 UpYunMultiPart

```go
func NewUpYunMultiPart(bucket, secret string, blocksize int64) *UpYunMultiPart
```

#### 上传

```go
func (ump *UpYunMultiPart) Put(fpath, saveas string,
    expireAfter int64, options map[string]interface{}) (*MergeResp, error)
```

---

### UPYUN 音视频处理接口

#### UpYunMedia

```go
type UpYunMedia struct {
    Username  string    // 操作员
    Passwd    string    // 密码
    Bucket    string    // 空间名（即服务名称）
}
```

#### MediaStatusResp

```go
type MediaStatusResp struct {
    Tasks map[string]interface{} `json:"tasks"`
}
```

#### 初始化 UpYunMedia

```go
func NewUpYunMedia(bucket, user, pass string) *UpYunMedia
```

#### 提交任务

```go
func (upm *UpYunMedia) PostTasks(src, notify string,
    tasks []map[string]interface{}) ([]string, error)
```

`src` 音视频文件 UPYUN 存储路径，`notify` 回调URL，`tasks` 任务列表，返回结果为任务 id 列表。

#### 查询进度

```go
func (upm *UpYunMedia) GetProgress(task_ids string) (*MediaStatusResp, error)
```

`task_ids` 是多个 `task_id` 用 `,` 连接起来。
