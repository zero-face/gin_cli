package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"web_app/settings"
)

/**
 * @Author Zero
 * @Date 2022/4/24 16:24
 * @Version 1.0
 * @Description
 **/

var (
	logger *zap.Logger
)

func Init(cfg *settings.LogConfig) (err error) {
	//写同步器
	writeSyncer := getLogWriter(cfg.Filename,
		cfg.MaxSize,
		cfg.MaxAge,
		cfg.MaxBackups,
	)
	//编码器
	encoder := getEncoder()
	//日志级别（经过UnmarshalText（）会将level设置成level类型的对应界别，其值其实是整型，类型为int8而已，只是被封装用来表示日志级别）
	level := new(zapcore.Level)
	err = level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return err
	}
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger = zap.New(core, zap.AddCaller())
	//sugaredLogger = logger.Sugar()
	//替换掉zap中全局的logger
	zap.ReplaceGlobals(logger)
	return nil
}

//获取编码器
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder, //人类可读的方式输出时间
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)

}

//获取日志的写同步器
func getLogWriter(fileName string, maxSize int, maxAge int, maxBackups int) zapcore.WriteSyncer {
	//记录到(一个)日志中
	//file, _ := os.Create("./test.log")
	//return zapcore.AddSync(file)
	//日志归档
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName,   //路径
		MaxSize:    maxSize,    //单个日志文件大小
		MaxBackups: maxBackups, //备份数量
		MaxAge:     maxAge,     //最大备份天数
		Compress:   false,      //是否压缩
	}
	//将实现了io.writer的类型传入，然后返回一个写同步器
	//内部实现是将writer传入，组装成一个WriteWrapper（写构造器）的结构体
	//写构造器类型实现了写同步器接口中的方法，所以就是一个写同步器
	//相当于一个实现io.writer的结构，就能经过封装编程一个写同步器，结构内部就需要实现日志写出的方式，路径，日志大小等等（这里使用lumberJackLogger组件实现的）
	return zapcore.AddSync(lumberJackLogger)
}

//替换gin中默认的日志的
//中间件-记录每次请求的日志信息
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start)
		zap.L().Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

//收集项目可能出现的panic
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
