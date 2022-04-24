package mysql

/**
 * @Author Zero
 * @Date 2022/4/24 16:54
 * @Version 1.0
 * @Description
 **/
import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"web_app/settings"
)

//声明一个全局的db对象
var db *sqlx.DB

func Init(cfg *settings.MysqlConfig) (err error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName)
	//默认是open + ping
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("connect db failed, err: %v\n", zap.Error(err))
		return
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns) //最大连接数
	db.SetMaxIdleConns(cfg.MaxIdleConns) //最大空闲连接数
	return
}
func Close() {
	_ = db.Close()
}
