package controller

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
)

//go:embed dist
var content embed.FS

type HttpServer struct {
	handlerServer HandlerServer
}

func NewHttpService(handlerServer HandlerServer) *HttpServer {
	return &HttpServer{
		handlerServer: handlerServer,
	}
}

func (h *HttpServer) StartHttpServer() {
	r := gin.Default()
	api := r.Group("/api")
	//create pipeline job
	api.POST("/pipeline", h.handlerServer.createPipeline)
	//update pipeline job
	api.PUT("/pipeline/:oldName", h.handlerServer.updatePipeline)
	//get pipeline job
	api.GET("/pipeline/:name", h.handlerServer.getPipeline)
	//delete pipeline job and pipeline job detail
	api.DELETE("/pipeline/:name", h.handlerServer.deletePipeline)
	//get pipeline job list
	api.GET("/pipeline", h.handlerServer.pipelineList)
	//get pipeline job detail info
	api.GET("/pipeline/:name/detail/:id", h.handlerServer.getPipelineDetail)
	//delete pipeline job detail
	api.DELETE("/pipeline/:name/detail/:id", h.handlerServer.deleteJobDetail)
	//get pipeline job detail list
	api.GET("/pipeline/:name/details", h.handlerServer.getPipelineDetailList)
	//exec pipeline job
	api.POST("/pipeline/:name/exec", h.handlerServer.execPipeline)
	//re exec pipeline detail job
	api.POST("/pipeline/:name/:id/re-exec", h.handlerServer.reExecuteJob)
	//stop pipeline job
	api.POST("/pipeline/:name/:id/stop", h.handlerServer.stopJobDetail)
	api.GET("/pipeline/:name/logs/:id", h.handlerServer.getJobLog)
	api.GET("/pipeline/:name/logs/:id/:stagename", h.handlerServer.getJobStageLog)
	// get template list
	api.GET("/pipeline/templates", h.handlerServer.getTemplates)
	api.GET("/pipeline/template/:id", h.handlerServer.getTemplateDetail)
	api.GET("/ping", func(c *gin.Context) {
		//输出json结果给调用方
		Success("", c)
	})
	fe, _ := fs.Sub(content, "dist")
	r.NoRoute(gin.WrapH(http.FileServer(http.FS(fe))))
	r.Run(":8080") // listen and serve on
}
