package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web_app/dao/mysql"
	"web_app/dao/redis"
	"web_app/logger"
	"web_app/routes"
	"web_app/settings"
)

/**
 * @Author Zero
 * @Date 2022/4/23 21:12
 * @Version 1.0
 * @Description go web脚手架
 **/
func main() {
	//加载配置文件
	if err := settings.Init(); err != nil {
		fmt.Printf("init setting failed, err: %v\n", err)
	}
	zap.L().Debug("settings init setting success.....")

	//初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err: %v\n", err)
	}
	//将日志落盘
	defer func() {
		err := zap.L().Sync()
		if err != nil {
			fmt.Printf("sync log faied, err: %v\n", err)
			zap.L().Fatal("sync log failed, err", zap.Error(err))
		}
	}()
	zap.L().Debug("logger init setting success.....")

	//初始化MySQL连接
	if err := mysql.Init(settings.Conf.MysqlConfig); err != nil {
		fmt.Printf("init logger failed, err: %v\n", err)
	}
	//最后释放连接
	defer mysql.Close()
	zap.L().Debug("mysql init setting success.....")

	//初始化Redis连接
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init logger failed, err: %v\n", err)
	}
	//释放redis连接
	defer redis.Close()
	zap.L().Debug("redis init setting success.....")

	//注册路由
	engine := routes.SetUp()
	//启动服务(优雅关机)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: engine,
	}
	//使用原生方式启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen: ", zap.Error(err))
		}
	}()
	//等待中断信号来优雅关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) //创建一个接收信息号的通道
	//kill默认会发送syscall.SIGTERM信号
	// kill - 2会发送syscall.SIGINT信号，我们常用的Ctrl+c就是触发的sigint信号
	//kill -9发送syscall.sigkill信号，但是不能被捕获，所以不需要添加
	//signal.notify把收到的所有信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //此处不会阻塞
	<-quit                                               //此处阻塞，当收到上述两三种信号时慈爱会往下执行
	zap.L().Info("shutdown ....")
	//创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("server shutdown", zap.Error(err))
	}
	zap.L().Info("server exiting")
}
