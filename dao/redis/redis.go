package redis

/**
 * @Author Zero
 * @Date 2022/4/24 17:05
 * @Version 1.0
 * @Description
 **/
import (
	"fmt"
	"github.com/go-redis/redis"
	"web_app/settings"
)

var rdb *redis.Client

func Init(cfg *settings.RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host,
			cfg.Port,
		),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	_ = rdb.Close()
}
