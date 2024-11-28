package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ImageInfo struct {
	Name    string `json:"name"`
	NewName string `json:"newname"`
	Status  string `json:"status"`
	Date    string `json:"date"`
}

type ImageList struct {
	Waitlist   []*ImageInfo
	Finishlist []*ImageInfo
}

type HttpServer struct {
	Router    *gin.Engine
	imagelist ImageList
	mtxList   sync.Mutex
}

func (h *HttpServer) Api_GetImageList(c *gin.Context) {
	// 将结构体响应为 JSON
	h.mtxList.Lock()
	defer h.mtxList.Unlock()
	c.JSON(http.StatusOK, h.imagelist)
}

// 如果Waitlist 大于20，不允许添加, 否则添加到Waitlist
func (h *HttpServer) Api_AddTask(c *gin.Context) {
	ii := &ImageInfo{}

	// 绑定JSON请求体到结构体
	if err := c.ShouldBindJSON(ii); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(ii.Name) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is err"})
		return
	}

	h.mtxList.Lock()
	defer h.mtxList.Unlock()
	if len(h.imagelist.Waitlist) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "waitlist is full"})
		return
	}

	ii.NewName = GetDestName(ii.Name)
	h.imagelist.Waitlist = append(h.imagelist.Waitlist, ii)

	fmt.Println("add task success:", ii.Name)

	// 响应成功
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (h *HttpServer) DoTask() {
	for {
		time.Sleep(time.Second * 1)
		if len(h.imagelist.Waitlist) == 0 {
			time.Sleep(time.Second * 3)
			continue
		}
		ii := h.imagelist.Waitlist[0]
		err := DoOne(ii.Name, ii.NewName)
		ii.Status = "success"
		if err != nil {
			fmt.Println("do task error:", err)
			ii.Status = "failed"
		}
		h.mtxList.Lock()
		// 把ii从Waitlist中移除，并添加到Finishlist中
		h.imagelist.Waitlist = h.imagelist.Waitlist[1:]
		h.imagelist.Finishlist = append(h.imagelist.Finishlist, ii)
		h.mtxList.Unlock()
	}
}

func (h *HttpServer) Run() {
	// 开启处理器
	go h.DoTask()

	// 创建一个 Gin 路由器
	h.Router = gin.Default()

	// 设置静态文件服务，www 目录作为根目录
	//h.Router.Static("/static", "./www")
	// 设置根目录和/index.html指向www目录下的index.html
	h.Router.GET("/", func(c *gin.Context) {
		indexPath := filepath.Join("www", "index.html")
		http.ServeFile(c.Writer, c.Request, indexPath)
	})

	h.Router.GET("/index.html", func(c *gin.Context) {
		indexPath := filepath.Join("www", "index.html")
		http.ServeFile(c.Writer, c.Request, indexPath)
	})

	h.Router.GET("/favicon.ico", func(c *gin.Context) {
		indexPath := filepath.Join("www", "favicon.ico")
		http.ServeFile(c.Writer, c.Request, indexPath)
	})

	// 设置静态文件服务，www目录作为根目录
	h.Router.Static("/static", "./www")

	apiGroup := h.Router.Group("/api")
	{
		apiGroup.GET("/getlist", h.Api_GetImageList)
		apiGroup.POST("/addTask", h.Api_AddTask)
	}

	// POST 接口示例
	h.Router.POST("/postExample", func(c *gin.Context) {
		// 定义一个请求数据结构体
		var requestData struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}

		// 绑定 JSON 请求体到结构体
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 响应相同的数据
		c.JSON(http.StatusOK, requestData)
	})

	// 启动服务器
	h.Router.Run(":3080") // 默认在 8080 端口
}
