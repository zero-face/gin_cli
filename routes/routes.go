package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"web_app/logger"
)

/**
 * @Author Zero
 * @Date 2022/4/24 17:13
 * @Version 1.0
 * @Description
 **/
func SetUp() *gin.Engine {
	engine := gin.New()
	//全局注册zap的logger
	engine.Use(logger.GinLogger(), logger.GinRecovery(true))
	engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return engine
}
