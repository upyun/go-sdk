package main

/**
 * 又拍云异步处理使用例子
 * 测试与运行例子的时，用户需要根据自己的需求填写对应的配置(user-profle.go)，参数
 */

import (
	"fmt"

	"github.com/upyun/go-sdk/v3/upyun"
)

func asyncProcess(appName string, accept string, source string, tasks []interface{}) {
	ids, err := up.CommitTasks(
		&upyun.CommitTasksConfig{
			AppName:   appName,
			NotifyUrl: NOTIFY_URL,
			Accept:    accept,
			Source:    source,
			Tasks:     tasks,
		},
	)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ids)
	}
}

/**
 * 异步音视频处理
 * http://docs.upyun.com/cloud/av/
 */
func videoAsyncProcess() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"type":    "video",
			"avopts":  "/s/128x96",
			"save_as": SAVE_AS_VIDEO,
		},
	}
	asyncProcess("naga", "json", SAVE_KEY_VIDEO, tasks)
}

/**
 * 压缩
 * http://docs.upyun.com/cloud/unzip/
 */
func compressProcess() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"sources": []string{SAVE_KEY_IMAGE, SAVE_KEY_VIDEO, SAVE_KEY_DOC},
			"save_as": SAVE_AS_ZIP,
		},
	}
	asyncProcess("compress", "", "", tasks)
}

/**
 * 解压缩
 * http://docs.upyun.com/cloud/unzip/
 */
func depressProcess() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"sources": SAVE_AS_ZIP,
			"save_as": SAVE_AS_DIR,
		},
	}
	asyncProcess("depress", "", "", tasks)
}

/**
 * 文件拉取
 * http://docs.upyun.com/cloud/spider/
 */
func spidermanProcess() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"url":     URL,
			"save_as": SAVE_AS_IMAGE,
		},
	}
	asyncProcess("spiderman", "", "", tasks)
}

/**
 * 文档转换
 * http://docs.upyun.com/cloud/uconvert/
 */
func uconvertProcess() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"source":  SAVE_KEY_DOC,
			"save_as": SAVE_AS_DOC,
		},
	}
	asyncProcess("uconvert", "", "", tasks)
}

/**
 * 异步图片拼接
 * http://docs.upyun.com/cloud/async_image/
 */
func jigsawProcess() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"image_matrix": [][]string{{SAVE_KEY_IMAGE, SAVE_KEY_IMAGE}, {SAVE_KEY_IMAGE, SAVE_KEY_IMAGE}},
			"save_as":      SAVE_AS_IMAGE,
		},
	}
	asyncProcess("jigsaw", "", "", tasks)
}
