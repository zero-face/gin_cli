package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify" //监控配置文件是否变化
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

//全局变量：保存全局所有配置信息
var Conf = new(AppConfig)

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Mode         string `mapstructure:"mode"`
	Version      string `mapstructure:"version"`
	Port         int    `mapstructure:"port"`
	*LogConfig   `mapstructure:"log"`
	*MysqlConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}
type MysqlConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DbName       int    `mapstructure:"db_name"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     string `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() error {
	viper.SetConfigName("config")  //指定配置文件名称（无需后缀）
	viper.SetConfigType("yaml")    //指定配置文件类型
	viper.AddConfigPath("./conf/") //指定配置文件路径
	viper.AddConfigPath(".")       //指定配置文件路径
	err := viper.ReadInConfig()
	if err != nil {
		//读取配置文件失败
		panic(fmt.Errorf("Fatal error config file: %s\n", err))
		return err
	}
	//将读取到的信息反序列化到Conf中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err: %v\n", err)
	}
	//监听配置文件,自动更新
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("%s配置文件修改了...%v\n", in.Name, in.Op.String())
		zap.L().Info("配置文件修改了")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err: %v\n", err)
		}
	})
	return nil
}
