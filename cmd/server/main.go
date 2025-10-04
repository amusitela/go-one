package main

import (
	"context"
	"go-one/internal/api"
	"go-one/internal/conf"
	"go-one/internal/model"
	"go-one/internal/server"
	"go-one/util"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func main() {
	// 从配置文件读取配置
	conf.Init()

	// 设置Gin运行模式
	gin.SetMode(os.Getenv("GIN_MODE"))

	// 初始化Handler
	api.HandlerApi = api.NewHandler(model.DB)

	// 创建路由
	router := server.NewRouter()

	// 获取端口配置
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 设置信号处理，用于优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动 HTTP 服务器
	go func() {
		util.Log().Info("启动服务器在 :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.Log().Error("服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()

	// 阻塞等待退出信号
	<-sigChan
	util.Log().Info("收到退出信号，正在优雅退出...")

	// 优雅关闭 HTTP 服务器
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		util.Log().Error("服务器关闭错误: %v", err)
	}

	// Flush sentry events on shutdown
	sentry.Flush(2 * time.Second)

	// 关闭数据库连接
	if model.DB != nil {
		if sqlDB, err := model.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}

	util.Log().Info("退出完成")
}
