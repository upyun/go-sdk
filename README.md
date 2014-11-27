# UPYUN Go SDK

[![Build Status](https://travis-ci.org/upyun/go-sdk.svg?branch=master)](https://travis-ci.org/upyun/go-sdk)

### 常量
```
const (
    Auto    = "v0.api.upyun.com"
    Telecom = "v1.api.upyun.com"
    Cnc     = "v2.api.upyun.com"
    Ctt     = "v3.api.upyun.com"
)

const (
    DefaultMaxChunkSize = 8192
    DefaultMinChunkSize = 1
)

```

### 类型

```
type FileInfo struct {
    Type string
    Date string
    Size int64
}

type Info struct {
    Name string
    Type string
    Size int64
    Time int64
}

type UpYun struct {
    Bucket   string
    Username string
    Passwd   string
    Endpoint string

    Timeout   int
    ChunkSize int
}


func NewUpYun(bucket, username, passwd string) *UpYun

func (u *UpYun) Delete(key string) error

func (u *UpYun) Get(key string, value *os.File) error

func (u *UpYun) GetInfo(key string) (*FileInfo, error)

func (u *UpYun) GetList(key string) ([]Info, error)

func (u *UpYun) Mkdir(key string) error

func (u *UpYun) Put(key string, value *os.File, md5 bool, secret string) (string, error)

func (u *UpYun) SetChunkSize(chunksize int) (int, error)

func (u *UpYun) SetEndpoint(endpoint string) (string, error)

func (u *UpYun) SetTimeout(t int)

func (u *UpYun) Usage() (int64, error)

func (u *UpYun) Version() string

```

## 初始化

```
u := upyun.NewUpYun("bucket", "username", "passwd")

```

## 版本

```
u.Version()

```

## 设置

### 设置上传下载分块大小
```
u.SetChunkSize(DefaultMaxChunkSize)
```

### 设置线路

***(default: v0.api.upyun.com)***

> v0.api.upyun.com //自动判断最优线路
> 
> v1.api.upyun.com //电信线路
>
> v2.api.upyun.com //联通（网通）线路
>
> v3.api.upyun.com //移动（铁通）线路

```
u.SetEndpoint(Auto)
```
### 设置连接api超时时间
***(default: 60s)***

```
u.SetTimeout(30)
```
## API

### Usage

```
used, err := u.Usage()
```

返回已使用的空间的量

### Mkdir

```
err := u.Mkdir("/path/to/dir")
```

### GetInfo
```
fileInfo, err := u.GetInfo("/path/to/file/or/dir")
```

### GetList
```
infoList, err := u.GetList("/path/to/dir")
```

### Delete
```
err := u.Delete("/path/to/file/or/dir")
```
***删除目录的时候，目录必须为空***

### Get, Put
**目前只支持文件的上传和下载**

上传

```
func (u *UpYun) Put(key string, value *os.File, md5 bool, secret string) (string, error)
```
* md5表示是否需要md的校验，需要填入`true`，不需要填入`false`
* secret表示是否需要加密，若设置该值，则无法直接访问原图，需要在原图URL的基础上加上密钥值才能访问，若不设置，置为`""`

> 注： 设置 Content-Secret 密钥后，原图将被保护，不能被直接访问，只有缩略图是允许被直接访问的 设置密钥后，若需访问原图，需要在 URL 后加上「缩略图间隔符号」和「访问密钥」（如： 当缩略图间隔符为 !，访问密钥为 secret，那么，原图访问方式即为： http://bucket.b0.upaiyun.com/sample.jpg!secret）

```
fi, err := os.Open("foo.txt")
retHeaders, err := u.Put("/path/to/file", fi, false, "")
```

下载


```
fo, err := os.Create("foo.txt")
err = u.Get("/path/to/file", fo)
```
