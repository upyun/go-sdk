package main

/**
 * 又拍云同步处理使用例子
 * 测试与运行例子的时，用户需要根据自己的需求填写对应的配置(user-profle.go)，参数
 */

import (
	"fmt"

	"github.com/upyun/go-sdk/upyun"
)

/**
 * 同步视频拼接
 * http://docs.upyun.com/cloud/sync_video/#m3u8
 */
func concatM3U8() {
	// kwargs 参考又拍云文档说明
	kwargs := map[string]interface{}{
		"m3u8s":   []string{SAVE_KEY_M3U8, SAVE_KEY_M3U8, SAVE_KEY_M3U8},
		"save_as": SAVE_AS_M3U8,
	}
	result, err := up.CommitSyncTasks(upyun.SyncCommonTask{
		TaskUri: "/m3u8er/concat",
		Kwargs:  kwargs,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

/**
 * 同步视频剪辑
 * http://docs.upyun.com/cloud/sync_video/#m3u8_1
 */
func clipM3U8() {
	// kwargs 参考又拍云文档说明
	kwargs := map[string]interface{}{
		"m3u8":    SAVE_KEY_M3U8,
		"save_as": SAVE_AS_M3U8,
		"include": []int{0, 10},
	}
	result, err := up.CommitSyncTasks(upyun.SyncCommonTask{
		TaskUri: "/m3u8er/clip",
		Kwargs:  kwargs,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

/**
 * 视频截图
 * http://docs.upyun.com/cloud/sync_video/#m3u8_2
 */
func snapshot() {
	// kwargs 参考又拍云文档说明
	kwargs := map[string]interface{}{
		"source":  SAVE_KEY_VIDEO,
		"save_as": SAVE_AS_IMAGE,
		"point":   "00:00:05",
	}
	result, err := up.CommitSyncTasks(upyun.SyncCommonTask{
		TaskUri: "/snapshot",
		Kwargs:  kwargs,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

/**
 * 获取 M3U8 信息
 * http://docs.upyun.com/cloud/sync_video/#m3u8_3
 */
func getM3U8Meta() {
	// kwargs 参考又拍云文档说明
	kwargs := map[string]interface{}{
		"m3u8": SAVE_KEY_M3U8,
	}
	result, err := up.CommitSyncTasks(upyun.SyncCommonTask{
		TaskUri: "/m3u8er/get_meta",
		Kwargs:  kwargs,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

/**
 * 获取音视频元信息
 * http://docs.upyun.com/cloud/sync_video/#_14
 */
func getAvMeta() {
	// kwargs 参考又拍云文档说明
	kwargs := map[string]interface{}{
		"source": SAVE_KEY_MP3,
	}
	result, err := up.CommitSyncTasks(upyun.SyncCommonTask{
		TaskUri: "/avmeta/get_meta",
		Kwargs:  kwargs,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
