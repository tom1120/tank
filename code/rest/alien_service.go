package rest

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
)

// @Service
type AlienService struct {
	BaseBean
	matterDao         *MatterDao
	matterService     *MatterService
	userDao           *UserDao
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	shareService      *ShareService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
	asynqService      *AsynqService
	taskHandler       *MyTaskHandler
}

func (this *AlienService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if c, ok := b.(*ShareService); ok {
		this.shareService = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}

	b = core.CONTEXT.GetBean(this.asynqService)
	if c, ok := b.(*AsynqService); ok {
		this.asynqService = c
	}

	b = core.CONTEXT.GetBean(this.taskHandler)
	if c, ok := b.(*MyTaskHandler); ok {
		this.taskHandler = c
	}
}

func (this *AlienService) PreviewOrDownload(
	writer http.ResponseWriter,
	request *http.Request,
	uuid string,
	filename string,
	withContentDisposition bool) {

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.Name != filename {
		panic(result.BadRequest("filename in url incorrect"))
	}

	//only private file need auth.
	if matter.Privacy {

		//1.use downloadToken to auth.
		downloadTokenUuid := request.FormValue("downloadTokenUuid")
		if downloadTokenUuid != "" {
			downloadToken := this.downloadTokenDao.CheckByUuid(downloadTokenUuid)
			if downloadToken.ExpireTime.Before(time.Now()) {
				panic(result.BadRequest("downloadToken has expired"))
			}

			if downloadToken.MatterUuid != uuid {
				panic(result.BadRequest("token and file info not match"))
			}

			tokenUser := this.userDao.CheckByUuid(downloadToken.UserUuid)
			if matter.UserUuid != tokenUser.Uuid {
				panic(result.UNAUTHORIZED)
			}

			//TODO: expire the download token. If download by chunk, do this later.
			downloadToken.ExpireTime = time.Now()
			this.downloadTokenDao.Save(downloadToken)

		} else {

			operator := this.findUser(request)

			//use share code to auth.
			shareUuid := request.FormValue("shareUuid")
			shareCode := request.FormValue("shareCode")
			shareRootUuid := request.FormValue("shareRootUuid")

			this.shareService.ValidateMatter(request, shareUuid, shareCode, operator, shareRootUuid, matter)

		}
	}

	//download directory
	if matter.Dir {

		this.matterService.DownloadZip(writer, request, []*Matter{matter})

	} else {

		//handle the image operation.
		needProcess, imageResizeM, imageResizeW, imageResizeH := this.imageCacheService.ResizeParams(request)
		if needProcess {

			//if image, try to use cache.
			mode := fmt.Sprintf("%s_%d_%d", imageResizeM, imageResizeW, imageResizeH)
			imageCache := this.imageCacheDao.FindByMatterUuidAndMode(matter.Uuid, mode)
			if imageCache == nil {
				imageCache = this.imageCacheService.cacheImage(writer, request, matter)
			}

			//download the cache image file.
			this.matterService.DownloadFile(writer, request, GetUserCacheRootDir(imageCache.Username)+imageCache.Path, imageCache.Name, withContentDisposition)

		} else {
			this.matterService.DownloadFile(writer, request, matter.AbsolutePath(), matter.Name, withContentDisposition)
		}

	}

	//async increase the download times.
	go core.RunWithRecovery(func() {
		this.matterDao.TimesIncrement(uuid)
	})

}

func (this *AlienService) VideoPreviewAndHandle(writer http.ResponseWriter, request *http.Request, uuid string) {
	// 获取尺寸
	width := request.FormValue("w")
	if width == "" {
		panic(result.BadRequest("width is empty"))
	}
	// 获取用户
	// user := this.userDao.checkUser(request)
	// 需要做获取其物料信息
	matter := this.matterDao.CheckByUuid(uuid)
	// 判断其1280或640文件是否存在
	filenamenoext := matter.Name[0:strings.LastIndex(matter.Name, ".")]
	filenameext := matter.Name[strings.LastIndex(matter.Name, "."):]
	video1280 := this.matterDao.FindByUserUuidAndPuuidAndDirAndName("", matter.Puuid, false, filenamenoext+"_"+width+filenameext)
	if video1280 != nil {
		this.matterService.DownloadFile(writer, request, video1280.AbsolutePath(), video1280.Name, true)
	} else {
		// 触发视频处理后台任务
		taskIdStuct := this.asynqService.AsynqVideoTaskServiceBeforeHandle(request, width, uuid)
		// panic(result.BadRequest("video_" + width + " not found"))
		panic(&result.WebResult{Code: result.BAD_REQUEST.Code, Msg: "video_" + width + " not found", Data: taskIdStuct})

	}

}

func (this *AlienService) VideoCoverPngPreviewHandle(writer http.ResponseWriter, request *http.Request, uuid string) {

	// 获取用户
	// user := this.userDao.checkUser(request)
	// 需要做获取其物料信息
	matter := this.matterDao.CheckByUuid(uuid)
	// 判断其1280或640文件是否存在
	filenamenoext := matter.Name[0:strings.LastIndex(matter.Name, ".")]
	// filenameext := matter.Name[strings.LastIndex(matter.Name, "."):]
	pngcover := this.matterDao.FindByUserUuidAndPuuidAndDirAndName("", matter.Puuid, false, filenamenoext+".png")
	if pngcover != nil {
		this.matterService.DownloadFile(writer, request, pngcover.AbsolutePath(), pngcover.Name, true)
	} else {
		// 触发视频处理
		pngmater, _ := this.taskHandler.GetSnapshot(matter, 1)
		// panic(result.BadRequest("video_" + width + " not found"))
		// panic(&result.WebResult{Code: result.BAD_REQUEST.Code, Msg: "封面已生成", Data: ""})
		this.matterService.DownloadFile(writer, request, pngmater.AbsolutePath(), pngmater.Name, true)

	}

}
