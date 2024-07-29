package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/enums"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/hibiken/asynq"
)

type AsynqService struct {
	BaseBean
	matterDao     *MatterDao
	matterService *MatterService
}

func (this *AsynqService) Init() {
	this.BaseBean.Init()
	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}
}

func (this *AsynqService) AsynqVideoTaskServiceBeforeHandle(request *http.Request, w string, uuid string) *TaskIdStuct {
	widths := strings.Split(w, "_")

	taskIdstruct := &TaskIdStuct{
		Success: false,
		TaskId:  "",
		ErrStr:  "未知错误",
	}

	user := this.checkUser(request)
	username := user.Username

	matter := this.matterService.Detail(request, uuid)
	filename := matter.Name

	filenameextname := filename[strings.LastIndex(filename, "."):]

	// if util.IsContain(enums.VideoType("").GetAllString(), filenameextname) {
	// 	this.logger.Info("视频文件在范围内:" + strings.Join(enums.VideoType("").GetAllString(), ","))
	// }

	// 判断是否为视频文件
	if util.IsContain(enums.VideoType("").GetAllString(), filenameextname) {
		// 触发视频处理任务
		// videoTaskPayload := VideoTaskPayload{FileName: "D:\\360安全浏览器下载\\【案例银行-产品管理】设置产品售价操作讲解.mp4", UserName: "admin"}
		videoTaskPayload := VideoTaskPayload{FileName: matter.AbsolutePath(), UserName: username, UserUuid: user.Uuid, Widths: widths, Matter: matter}

		uuid4str := this.AsynqVideoTaskService(videoTaskPayload)

		taskIdstruct.Success = true
		taskIdstruct.TaskId = uuid4str
		taskIdstruct.ErrStr = ""

	} else {

		// taskstuct.ErrStr = fmt.Sprintf("视频文件支持.mp4/.avi/.mov/.wmv/.flv/.mkv，%+v不是视频文件", filename[strings.LastIndex(filename, "."):])
		taskIdstruct.ErrStr = "视频文件在范围内:" + strings.Join(enums.VideoType("").GetAllString(), ",")
		taskIdstruct.Success = false
	}

	return taskIdstruct

}

func (this *AsynqService) AsynqVideoTaskService(videoTaskPayload VideoTaskPayload) string {
	client := asynq.NewClient(
		asynq.RedisClientOpt{Addr: core.CONFIG.MyRedisUrl(),
			Password: core.CONFIG.MyRedisPassword(),
			DB:       core.CONFIG.MyRedisDb()})
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: core.CONFIG.MyRedisUrl(),
		Password: core.CONFIG.MyRedisPassword(),
		DB:       core.CONFIG.MyRedisDb()})
	payload, err := json.Marshal(videoTaskPayload)
	if err != nil {
		panic(err)
	}

	t1 := asynq.NewTask("videotask", payload)
	// uuid4, err := uuid.NewV4()
	// uuid4str := fmt.Sprintf("%+v", uuid4)
	// 任务id采取物料id和处理width来表示
	taskId := fmt.Sprintf("%s_%s", videoTaskPayload.Matter.Uuid, strings.Join(videoTaskPayload.Widths, "_"))
	if err != nil {
		panic(err)
	}

	// 判断任务id是否已存在当前任务队列中
	_, err = client.Enqueue(t1, asynq.TaskID(taskId), asynq.Timeout(30*6*time.Second), asynq.Retention(24*time.Hour), asynq.MaxRetry(0))
	switch {
	case errors.Is(err, asynq.ErrTaskIDConflict):
		// handle duplicate task
		inspector.DeleteTask("default", taskId)
		_, err = client.Enqueue(t1, asynq.TaskID(taskId), asynq.Timeout(30*6*time.Second), asynq.Retention(24*time.Hour), asynq.MaxRetry(0))
	case err != nil:
		// handle other errors
		panic(err)
	}

	// 直接打印会报错
	// this.logger.Info(" [*] Successfully enqueued task: %+v", info)

	return taskId

}
