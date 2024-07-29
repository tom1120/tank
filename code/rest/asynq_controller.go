package rest

import (
	"encoding/json"
	"net/http"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/hibiken/asynq"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type AsynqController struct {
	BaseController
	matterDao     *MatterDao
	matterService *MatterService
	asynqService  *AsynqService
}

func (this *AsynqController) Init() {
	this.BaseController.Init()
	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}
	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}
	b = core.CONTEXT.GetBean(this.asynqService)
	if b, ok := b.(*AsynqService); ok {
		this.asynqService = b
	}

}

func (this *AsynqController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {
	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))
	routeMap["/api/asynq/hello"] = this.Wrap(this.Hello, USER_ROLE_USER)
	routeMap["/api/asynq/asynqVideoTask"] = this.Wrap(this.AsynqVideoTask, USER_ROLE_USER)
	routeMap["/api/asynq/asynqTaskResultFetch"] = this.Wrap(this.AsynqTaskResultFetch, USER_ROLE_USER)
	return routeMap
}

func (this *AsynqController) Hello(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	input := ffmpeg_go.Input("D:\\360安全浏览器下载\\【案例银行-产品管理】设置产品售价操作讲解.mp4").Split()
	out1 := input.Get("0").Filter("scale", ffmpeg_go.Args{"1280:-1"}).
		Output("D:\\360安全浏览器下载\\【案例银行-产品管理】设置产品售价操作讲解_1280.mp4")
	out2 := input.Get("1").Filter("scale", ffmpeg_go.Args{"640:-1"}).
		Output("D:\\360安全浏览器下载\\【案例银行-产品管理】设置产品售价操作讲解_640.mp4")
	err := ffmpeg_go.MergeOutputs(out1, out2).OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		panic(err)
	}
	return this.Success("hello")
}

func (this *AsynqController) AsynqVideoTask(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	fileuuid := request.FormValue("uuid")
	widthstr := request.FormValue("w")
	// 自己建的，不是取的系统初始化的实例，会失去依赖
	// asynqInstance := &AsynqService{}
	taskIdStuct := this.asynqService.AsynqVideoTaskServiceBeforeHandle(request, widthstr, fileuuid)
	return this.Success(taskIdStuct)
}

func (this *AsynqController) AsynqTaskResultFetch(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	taskId := request.FormValue("taskId")

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: core.CONFIG.MyRedisUrl(),
		Password: core.CONFIG.MyRedisPassword(),
		DB:       core.CONFIG.MyRedisDb()})

	taskinfo, err := inspector.GetTaskInfo("default", taskId)
	if err != nil {
		panic(err)
	}
	// this.logger.Info(" [*] Successfully get task info: %+v", taskinfo)

	// if taskinfo.State == asynq.TaskStateCompleted {
	// 	this.logger.Info(" [*] Successfully get task info: %+v", taskinfo)
	// }
	videoTaskPayload := &VideoTaskPayload{}
	videoTaskResult := &VideoTaskResult{}
	err = json.Unmarshal(taskinfo.Payload, videoTaskPayload)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(taskinfo.Result, videoTaskResult)
	if err != nil {
		panic(err)
	}

	videoTaskQueryResult := &VideoTaskQueryResult{VideoTaskPayload: *videoTaskPayload, VideoTaskResult: *videoTaskResult}
	return this.Success(videoTaskQueryResult)
}
