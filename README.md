# UPYUN Go SDK

[![Build Status](https://travis-ci.org/upyun/go-sdk.svg?branch=master)](https://travis-ci.org/upyun/go-sdk)

UPYUN Go SDK make it easy to use UPYUN API!

### Example

```
u := upyun.NewUpYun("bucket", "username", "password")

// Get bucket usage
usage, err := u.Usage()

// Make dir
err := u.Mkdir("/foo/bar")

// Delete
err := u.Delete("/foo/bar1.txt")

// Get dir info list
info, err := u.GetList("/foo")

// Download
// io.Writer
buf := &bytes.Buffer{}
err := u.Get("/foo/bar2.txt", buf)
// os.File
file, err := os.Open("./abc.txt")
err := u.Get("/foo/bar3.txt", file)

// Upload
// os.File
file, err := os.Open("./abc.txt")
u.Put("/foo/bar2.txt", file, false, "")
// io.Reader
buf := &bytes.Buffer{}
_, err := io.Copy(buf, file)
u.Put("/foo/bar2.txt", buf, false, "")

// Purge
resp, err := u.Purge([]string{"/foo/bar.txt", "/foo/bar1.txt"})

// Form API
uf := upyun.NewUpYunForm("bucket", "form_api_key")
uf.Put("/foo/bar.txt", "./abc.txt", 100, nil)
```

### Installation

```
go install github.com/upyun/go-sdk/upyun
```

### Set chunk size
(Default: 32kb)

Chunk size is a size of a buffer which user use it to do copy operation, Golang's io.Copy have it hard coded at 32kb. This may not good for some cases,  you can change it with anysize.

**Example**

```
var chunksize int = 1024
upyun.SetChunkSize(chunksize)
```

### Set Endpoints
(Default: Auto)

1. `Auto`: Auto detect by user network
2. `Telecom`: (ISP) China Telecom
3. `Cnc`: (ISP) China Unicom
4. `Ctt`: (ISP) China Tietong

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")
u.SetEndpoint(upyun.Auto)

uf := upyun.NewUpYunForm("bucket", "form_api_key")
uf.SetEndpoint(upyun.Auto)
```

### Set Connect Timeout
(Default: 60)

Set the connection timeout when connect to endpoint

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")
u.SetTimeout(30)

uf := upyun.NewUpYunForm("bucket", "form_api_key")
uf.SetTimeout(30)
```

## HTTP REST API

### New

```func NewUpYun(bucket, username, passwd string) *UpYun```

Create a UpYun instance with your bucket infomation(bucketname, username, password), using this instance, you can upload, download, get file info, etc.

Try to reuse one instance will make it faster!

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")
```

### Usage

```func (u *UpYun) Usage() (int64, error)```

Get the usage of a bucket, if bucket infomation incorrect or err's not nil, it will return zero, otherwise, it return how many storage have been used by this bucket.

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")
usage, err := u.Usage()
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(usage)
}
```

### Mkdir

```func (u *UpYun) Mkdir(key string) error```

As it means, it will create a directory **recursively**.

**Example**

```
remote_dir := "/foo/bar/foo"

u := upyun.NewUpYun("bucket", "username", "password")
err := u.Mkdir(remote_dir)
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println("Successful make directory " + remote_dir)
}

```

### Delete

```func (u *UpYun) Delete(key string) error```

Delete a file or an **empty** directory.

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")

u.Delete("/foo")
```

### GetList

```func (u *UpYun) GetList(key string) ([]Info, error)```

Get a list of file information of the specified directory.

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")

list, err := u.GetList("/foo")
if err != nil {
	fmt.Println(err)
} else {
	for _, fi := range list {
		fmt.Println(fi)
	}
}
```

### GetInfo

```func (u *UpYun) GetInfo(key string) (FileInfo, error)```

Get file information of the specified file.

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")

info, err := u.GetInfo("/foo/bar1.txt")
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(info)
}
```


### Put

```func (u *UpYun) Put(key string, value io.Reader, md5 bool, secret string) (string, error)```

Upload io.Reader to remote file.

* `key`: remote file path
* `value`: a io.Reader where data is stored
* `md5`: set md5 to `true` to enable remote server's md5 chunksum, otherwise(`false`) not.
* `secret`: encrypt picture, with specified this argument, origin picture is no longer available, you should add `!secret` after origin picture's URL, like this, http://bucket.b0.upaiyun.com/sample.jpg!secret(origin picture is http://bucket.b0.upaiyun.com/sample.jpg). **Zero value of string("") means no encrypt.**

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")
file, err := os.Open("./abc.txt")

// resp will have some response headers which include origin picture's args if upload picture.
resp, err := u.Put("/foo/bar.txt", file, false, "")
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(resp)
}
```



### Get

```func (u *UpYun) Get(key string, value io.Writer) error```

Download remote file to a io.Writer

* `key`: remote file path
* `value`: a io.Writer use to store data

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")

buf := &bytes.Buffer{}
err := u.Get("/foo/bar.txt", buf)
if err != nil {
	fmt.Println(err)
}
```



### Purge

```func (u *UpYun) Purge(urls []string) (string, error)```

Purge files cache.

When Purge successfully, return with invalid urls and nil.

otherwise, return "" and error.

**Example**

```
u := upyun.NewUpYun("bucket", "username", "password")

invalidURL, err := u.Purge([]string{"http://bucket.b0.upaiyun.com/sample.jpg", "http://bucket.b0.upaiyun.com/sample1.jpg"})
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(invalidURL)
}
```

## HTTP FORM API

### New

```func NewUpYunForm(bucket, key string) *UpYunForm```

Create a instance of form api, the reason why separate it and REST API is the auth is different.

**Example**

```
uf := upyun.NewUpYunForm("chengzi", "your_bucket_form_key")
```

### Put

```func (uf *UpYunForm) Put(saveas, path string, expireAfter int64, options map[string]string) error```

Upload a file to remote path

* `saveas`: remote file path
* `path`: local file path
* `expireAfter`: request will expire after this time
* `options`: nil if nothing have to specified, otherwise look up [here](http://docs.upyun.com/api/form_api/#api_1).

**Example**

```
uf := upyun.NewUpYunForm("chengzi", "your_bucket_form_key")

err := uf.Put("/foo/bar1.txt", "./abc.txt", 100, nil)
if err != nil {
	fmt.Println(err)
}
```

## Contributor

Thanks for these guys' contribution!

* [luanzhu](https://github.com/luanzhu)

* [Wine93](https://github.com/Wine93)
