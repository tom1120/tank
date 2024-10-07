package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/disintegration/imaging"
	"github.com/eyebluecn/tank/code/core"
	uuid "github.com/nu7hatch/gouuid"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"

	"github.com/hibiken/asynq"
)

// MyTaskHandler 定义了如何处理任务
type MyTaskHandler struct {
	BaseBean
	matterDao *MatterDao
}

func (this *MyTaskHandler) Init() {
	this.BaseBean.Init()
	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

}

// ProcessTask 实现了 asynq.TaskHandler 接口
func (this *MyTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ProcessTask", r)
			rbytes, err := json.Marshal(r)
			if _, err = task.ResultWriter().Write(rbytes); err != nil {
				panic(err)
			}
		}
	}()

	// 因为与主线程是不同协程，这里需要单独初始化，才不会引起空指针调用
	this.Init()
	// 这里是处理任务的逻辑
	fmt.Printf("Processing task of type %q with payload: %+v\n", task.Type(), string(task.Payload()))
	switch task.Type() {
	case "videotask":
		// 处理 mytask 类型的任务
		return this.videoProcessTask(ctx, task)
	default:
		// 处理其他类型的任务
		return fmt.Errorf("unexpected task type: %q", task.Type())
	}

	// 返回 nil 表示任务处理成功
	return nil
}

func (this *MyTaskHandler) videoProcessTask(ctx context.Context, task *asynq.Task) error {
	this.logger.Info("videoProcessTask=====start")
	taskresult := VideoTaskResult{Success: false,
		TaskId: task.ResultWriter().TaskID()}

	paymodel := &VideoTaskPayload{}
	if err := json.Unmarshal(task.Payload(), paymodel); err != nil {
		taskresult.ErrStr = err.Error()
		panic(taskresult)
	}

	// time.Sleep(5 * time.Second)
	// FileName为绝对路径
	dotpos := strings.LastIndex(paymodel.FileName, ".")
	filenamenotext := paymodel.FileName[0:dotpos]

	relativeFileNameNoExt := filenamenotext[strings.LastIndex(filenamenotext, "/")+1:]

	// 截取第一帧作为封面
	_, err := this.GetSnapshot(paymodel.Matter, 1)

	// 处理宽度
	widths := paymodel.Widths
	// 默认初始化为2的数组，默认元素值为nil
	// outarr := make([]*ffmpeg_go.Stream, 0)

	// 数据库事务开始
	db := core.CONTEXT.GetDB()

	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			this.logger.Error("recovered from videoProcessTask", r)
			// 数据库回滚
			tx.Rollback()
			// 回写任务结果
			rbytes, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}
			if _, err := task.ResultWriter().Write(rbytes); err != nil {
				panic(err)
			}

		}
	}()

	// width判断
	sourcefi, err := ffmpeg_go.Probe(paymodel.FileName)

	this.logger.Info(sourcefi)

	if err != nil {
		taskresult.ErrStr = err.Error()
		panic(taskresult)
	}

	for i := 0; i < len(widths); i++ {
		// 获取视频宽度
		jsonobj, err := simplejson.NewJson([]byte(sourcefi))
		if err != nil {
			taskresult.ErrStr = err.Error()
			panic(taskresult)
		}
		width := jsonobj.Get("streams").GetIndex(0).Get("width").MustInt()
		task_width, err := strconv.Atoi(widths[i])
		if width < task_width {
			taskresult.ErrStr = "原视频宽度" + strconv.Itoa(width) + "小于任务视频宽度" + widths[i] + "，无法生成" + widths[i] + "的缩放"
			panic(taskresult)
		}

		matter := paymodel.Matter
		matter.Name = relativeFileNameNoExt + "_" + widths[i] + ".mp4"
		matter.Path = matter.Path[0:strings.LastIndex(matter.Path, "/")] + "/" + matter.Name
		matter_temp := this.matterDao.FindByUserUuidAndPathPublic(paymodel.UserUuid, matter.Path)
		if matter_temp == nil {
			input := ffmpeg_go.Input(paymodel.FileName).Split()
			// out1 := input.Get(string(i)).Filter("scale", ffmpeg_go.Args{widths[i] + ":-1"}).
			// 	Output(filenamenotext + "_1280.mp4")
			// out2 := input.Get("1").Filter("scale", ffmpeg_go.Args{"640:-1"}).
			// 	Output(filenamenotext + "_640.mp4")

			out := input.Get(string(i)).Filter("scale", ffmpeg_go.Args{widths[i] + ":-1"}).
				Output(filenamenotext + "_" + widths[i] + ".mp4")

			// outarr = append(outarr, out)

			// 执行缩放命令
			err := out.OverWriteOutput().ErrorToStdOut().Run()

			if err != nil {
				taskresult.ErrStr = err.Error() + "  给定宽高不能等比例缩放，请输入合理宽度值 或 ffmpeg转换其他错误"
				panic(taskresult)
			}

			// 写入企业网盘访问
			matter.Uuid = ""

			fi, err := os.Stat(filenamenotext + "_" + widths[i] + ".mp4")
			if err != nil {
				taskresult.ErrStr = err.Error()
				panic(taskresult)
			}

			matter.Size = fi.Size() // todo get file size
			// 创建完成后，invalid memory address or nil pointer dereference  调用函数内this为nil

			// 自带事务，不能使用此方法
			// this.matterDao.Create(matter)

			timeUUID, _ := uuid.NewV4()
			matter.Uuid = string(timeUUID.String())
			matter.CreateTime = time.Now()
			matter.UpdateTime = time.Now()
			matter.Sort = time.Now().UnixNano() / 1e6
			db := tx.Create(matter)

			if db.Error != nil {
				taskresult.ErrStr = db.Error.Error()
				panic(taskresult)
			}

		}

	}

	// panic("测试事务回滚")

	tx.Commit()

	// 合并执行缩放命令
	// if len(outarr) > 0 {
	// 	err := ffmpeg_go.MergeOutputs(outarr[0:]...).OverWriteOutput().ErrorToStdOut().Run()
	// 	if err != nil {
	// 		tx.Rollback()
	// 		taskresult.ResultStr = fmt.Sprintf("原文件%+v.mp4生成%+v个大小尺寸%+v的失败", relativeFileNameNoExt, len(widths), strings.Join(widths, "_"))
	// 		taskresult.ErrStr = err.Error()
	// 		panic(taskresult)
	// 	} else {
	// 		tx.Commit()
	// 	}
	// } else {
	// 	tx.Commit()
	// }

	this.logger.Info("videoProcessTask=====end")

	// 写入企业网盘以便访问

	// for i := 0; i < len(widths); i++ {
	// 	matter := paymodel.Matter

	// 	// 检测1280是否重复
	// 	matter.Name = relativeFileNameNoExt + "_" + widths[i] + ".mp4"
	// 	matter.Path = matter.Path[0:strings.LastIndex(matter.Path, "/")] + "/" + matter.Name
	// 	matter_temp := this.matterDao.FindByUserUuidAndPathPublic(paymodel.UserUuid, matter.Path)
	// 	if matter_temp == nil {

	// 	}

	// }

	// fmt.Printf("videoProcessTask=====matter:%+v", matter)

	// 回写任务结果
	taskresult.Success = true
	taskresult.ResultStr = fmt.Sprintf("原文件%+v.mp4生成%+v个大小尺寸%+v的成功", relativeFileNameNoExt, len(widths), strings.Join(widths, "_"))

	taskresultbytes, err := json.Marshal(taskresult)
	if _, err = task.ResultWriter().Write(taskresultbytes); err != nil {
		taskresult.ErrStr = err.Error()
		panic(taskresult)
	}

	return nil
}

func (this *MyTaskHandler) GetSnapshot(fileMatter *Matter, frameNum int) (matter *Matter, err error) {
	filename := fileMatter.Name[0:strings.LastIndex(fileMatter.Name, ".")] + ".png"
	fileRelativePath := fileMatter.Path[0:strings.LastIndex(fileMatter.Path, "/")] + "/" + filename

	fileAbsoluteVideoPath := fileMatter.AbsolutePath()

	fileAbsolutePath := fileAbsoluteVideoPath[0:strings.LastIndex(fileAbsoluteVideoPath, "/")] + "/" + filename

	matter = &Matter{
		Puuid:    fileMatter.Puuid,
		UserUuid: fileMatter.UserUuid,
		Username: fileMatter.Username,
		Dir:      false,
		Name:     filename,
		Md5:      "",
		Size:     0,
		Privacy:  false,
		Path:     fileRelativePath,
	}

	buf := bytes.NewBuffer(nil)
	err = ffmpeg_go.Input(fileAbsoluteVideoPath).Filter("select", ffmpeg_go.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg_go.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()

	if err != nil {
		log.Fatal("生成缩略图失败：", err)
		return nil, err
	}

	img, err := imaging.Decode(buf)
	if err != nil {
		log.Fatal("生成缩略图失败：", err)
		return nil, err
	}

	err = imaging.Save(img, fileAbsolutePath)
	if err != nil {
		log.Fatal("生成缩略图失败：", err)
		return nil, err
	}

	fi, err := os.Stat(fileAbsolutePath)

	if err != nil {
		log.Fatal("生成的缩略图文件无法统计大小", err)
		return nil, err
	}

	matter.Size = fi.Size()

	this.matterDao.Create(matter)

	return matter, nil
}
