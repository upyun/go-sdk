package main

/**
 * 又拍云人工智能使用例子
 * 测试与运行例子的时，用户需要根据自己的需求填写对应的配置(user-profle.go)，参数
 */

import (
	"fmt"

	"github.com/upyun/go-sdk/upyun"
)

func asyncAudit(appName string, tasks []interface{}) {
	ids, err := up.CommitTasks(
		&upyun.CommitTasksConfig{
			AppName:   appName,
			NotifyUrl: NOTIFY_URL,
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
 * 内容识别-图片
 * http://docs.upyun.com/ai/audit/
 */
func imageAsyncAudit() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"source": SAVE_KEY_IMAGE,
		},
	}
	asyncAudit("imgaudit", tasks)
}

/**
 * 内容识别-点播
 * http://docs.upyun.com/ai/audit/
 */
func videoAsyncAudit() {
	// tasks 参考又拍云文档说明
	tasks := []interface{}{
		map[string]interface{}{
			"source": SAVE_KEY_VIDEO,
		},
	}
	asyncAudit("videoaudit", tasks)
}

/**
 * 内容识别-直播创建
 * http://docs.upyun.com/ai/audit/
 */
func liveAudit() {
	// 参考又拍云文档说明
	result, err := up.CommitSyncTasks(upyun.LiveauditCreateTask{
		Source:    RTMP_SOURCE,
		SaveAs:    SAVE_AS_IMAGE,
		NotifyUrl: NOTIFY_URL,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

/**
 * 内容识别-直播取消
 * http://docs.upyun.com/ai/audit/
 */
func liveAuditCancel() {
	// 参考又拍云文档说明
	result, err := up.CommitSyncTasks(upyun.LiveauditCancelTask{
		TaskId: "e94523760efd30f9c7a68666b46cba70",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
