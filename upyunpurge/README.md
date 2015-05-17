#Purpose

With this library, you can call upyun cache refresh API to "purge" the old version of your files out of upyun system. 

#Features
Because upyun imposes a limit for how many urls one can purge for a minute, this library will send urls in batches and wait for 80 seconds before sending the next batch.

Currently, the limit is 600 URLs per minute, and the batch size is 550.

#How to use

```go
import (
    "github.com/luanzhu/go-sdk/upyunpurge"
    "log"
    "fmt"
)


func refreshURLs(bucket, username, passwd string, urls []string) {
	u := NewUpYunPurge(bucket, username, passwd)

	invalidURLs, purgedURLs, err := u.PurgeURLs(urls)
	if err != nil {
		// err contains the error message returned from server
		log.Fatal(err)
	} else {
		// If there is no error, invalidURLs contains a list of URLs that upyun
		// cannot process, which usually means that these URLs are not in 
        // current bucket.
		fmt.Println("invalidURLs: ", invalidURLs)

        // Because URLs may be sent in batches, purgedURLs contains the URLs that
        // were sent to upyun without error.
        // Items in invalidURLs will also appear here.
        fmt.Println("purgedURLs:", purgedURLs)
	}
}
```



