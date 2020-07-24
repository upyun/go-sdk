package main

/**
 * 又拍云 GO-SDK examle 用户配置文件
 * 测试与运行例子的时，用户需要根据自己的需求填写对应的配置，参数
 */

import (
	"os"

	"github.com/upyun/go-sdk/v3/upyun"
)

var (
	up = upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   os.Getenv("UPYUN_BUCKET"),
		Operator: os.Getenv("UPYUN_USERNAME"),
		Password: os.Getenv("UPYUN_PASSWORD"),
	})
)

const (

	// 指定的通知URL
	NOTIFY_URL = ""

	// 本地图片路径，适用于图片文件上传，预处理
	LOCAL_IMAGE = "./sample/sample.jpg"

	// 本地视频路径，适用于视频文件上传，预处理
	LOCAL_VIDEO = "./sample/sample.mp4"

	// 本地文档路径，包括PDF，PPT，WORD，EXCEL，适用于文档文件上传，预处理
	LOCAL_DOC = "./sample/sample.pptx"

	// 本地音频路径，适用于同步音频处理
	LOCAL_MP3 = "./sample/sample.mp3"

	// 云存储中保存的图片文件路径，适用于图片相关上传，预处理，图片内容识别
	SAVE_KEY_IMAGE = "/save_key.jpg"

	// 云存储中保存的视频文件路径，适用于视频相关上传，预处理，视频内容识别，同步音视频处理
	SAVE_KEY_VIDEO = "/save_key.mp4"

	// 云存储中保存的文档文件路径，适用于文档相关上传，预处理，文档转换
	SAVE_KEY_DOC = "/save_key.pptx"

	// 云存储中保存的音频文件路径，适用于同步音视频处理
	SAVE_KEY_MP3 = "/save_key.mp3"

	// 云存储中保存的 M3U8 文件路径，适用于同步音视频处理
	SAVE_KEY_M3U8 = "/save_key.m3u8"

	// 云存储中 save_as 参数指定的图片路径，适用于图片相关
	SAVE_AS_IMAGE = "/save_as.jpg"

	// 云存储中 save_as 参数指定的视频路径，适用于视频相关
	SAVE_AS_VIDEO = "/save_as.mp4"

	// 云存储中 save_as 参数指定的文档路径，适用于文档转换
	SAVE_AS_DOC = "/save_as"

	// 云存储中 save_as 参数指定的压缩文件路径，适用于文件压缩，解压
	SAVE_AS_ZIP = "/save_as.zip"

	// 云存储中目录，适用于文件解压
	SAVE_AS_DIR = "/save_as/"

	// 云存储中 save_as 参数指定的M3U8文件路径，适用于同步音视频处理
	SAVE_AS_M3U8 = "/save_as.m3u8"

	// 文件URL，适用于文件拉取
	URL = "http://p07vpkunh.bkt.clouddn.com/aaaaa/image.png"

	// RTMP源，适用于内容识别-直播
	RTMP_SOURCE = "rtmp://live.hkstv.hk.lxdns.com/live/hks"
)
