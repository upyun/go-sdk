package main

/**
 * 又拍云上传预处理使用例子
 * 测试与运行例子的时，用户需要根据自己的需求填写对应的配置(user-profle.go)，参数
 */

import (
	"fmt"

	"github.com/upyun/go-sdk/v3/upyun"
)

func asyncPreProcess(localPath string, saveKey string, apps []map[string]interface{}) {
	resp, err := up.FormUpload(&upyun.FormUploadConfig{
		LocalPath:      localPath,
		SaveKey:        saveKey,
		NotifyUrl:      NOTIFY_URL,
		ExpireAfterSec: 60,
		Apps:           apps,
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Code, resp.Msg, resp.Taskids)
	}
}

func syncPreProcess(localPath string, saveKey string, options map[string]interface{}) {
	resp, err := up.FormUpload(&upyun.FormUploadConfig{
		LocalPath:      localPath,
		SaveKey:        saveKey,
		NotifyUrl:      NOTIFY_URL,
		ExpireAfterSec: 60,
		Options:        options,
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Code, resp.Msg, resp.Taskids)
	}
}

/**
 * 图片上传异步预处理
 * http://docs.upyun.com/cloud/image/
 */
func imageAsyncPreProcess() {
	// apps 参考又拍云文档说明
	apps := []map[string]interface{}{
		map[string]interface{}{
			"name":           "thumb",
			"x-gmkerl-thumb": "/format/png",
			"save_as":        SAVE_AS_IMAGE,
		},
	}
	asyncPreProcess(LOCAL_IMAGE, SAVE_KEY_IMAGE, apps)
}

/**
 * 异步音视频上传预处理
 * http://docs.upyun.com/cloud/av/
 */
func videoAsyncPreProcess() {
	// apps 参考又拍云文档说明
	apps := []map[string]interface{}{
		map[string]interface{}{
			"name":    "naga",
			"type":    "video",
			"avopts":  "/s/128x96",
			"save_as": SAVE_AS_VIDEO,
		},
	}
	asyncPreProcess(LOCAL_VIDEO, SAVE_KEY_VIDEO, apps)
}

/**
 * 文档转换上传预处理
 * http://docs.upyun.com/cloud/uconvert/
 */
func fileConvertAsyncPreProcess() {
	// apps 参考又拍云文档说明
	apps := []map[string]interface{}{
		map[string]interface{}{
			"name":    "uconvert",
			"save_as": SAVE_AS_DOC,
		},
	}
	asyncPreProcess(LOCAL_DOC, SAVE_KEY_DOC, apps)
}

/**
 * 图片内容识别上传预处理
 * http://docs.upyun.com/ai/audit/
 */
func imageAuditAsyncPreProcess() {
	// apps 参考又拍云文档说明
	apps := []map[string]interface{}{
		map[string]interface{}{
			"name": "imgaudit",
		},
	}
	asyncPreProcess(LOCAL_IMAGE, SAVE_KEY_IMAGE, apps)
}

/**
 * 点播内容识别上传预处理
 * http://docs.upyun.com/ai/audit/
 */
func videoAuditAsyncPreProcess() {
	// apps 参考又拍云文档说明
	apps := []map[string]interface{}{
		map[string]interface{}{
			"name": "videoaudit",
		},
	}
	asyncPreProcess(LOCAL_VIDEO, SAVE_KEY_VIDEO, apps)
}

/**
 * 图片上传同步预处理
 * http://docs.upyun.com/cloud/image/
 */
func imageSyncPreProcess() {
	// options 参考又拍云文档说明
	options := map[string]interface{}{
		"x-gmkerl-thumb": "/format/png",
	}
	syncPreProcess(LOCAL_IMAGE, SAVE_KEY_IMAGE, options)
}

/**
 * 同步音频处理
 * http://docs.upyun.com/cloud/sync_audio/
 */
func voiceSyncPreProcess() {
	// options 参考又拍云文档说明
	options := map[string]interface{}{
		"x-audio-avopts": "/ab/48/ac/5/f/ogg",
	}
	syncPreProcess(LOCAL_MP3, SAVE_KEY_MP3, options)
}
