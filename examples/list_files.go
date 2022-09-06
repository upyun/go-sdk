package main

import (
	"fmt"
	"github.com/upyun/go-sdk/v3/upyun"
	"log"
)

// ListFiles ...
func ListFiles() {

	fpath := "/upyun"

	header := make(map[string]string)

	config := &upyun.ListObjectsConfig{
		Path:         fpath,
		Headers:      header,
		MaxListTries: 0,
		DescOrder:    false,
		Iter:         "",
		Limit:        30,
	}
	fileInfos, iter, err := up.ListObjects(config)
	if err != nil {
		log.Printf("ls %s: %v", fpath, err)
	}
	for _, fInfo := range fileInfos {
		fmt.Println(fInfo.Name)
	}
	fmt.Println(iter)
}
