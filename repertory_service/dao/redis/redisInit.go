package redis

import (
	"alivePlatform/repertory_service/config"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

var (
	Rs *redsync.Redsync
)

func Init(cfg *config.RedisConfig) error {
	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password, // 密码
		DB:       cfg.DB,       // 数据库
		PoolSize: cfg.PoolSize, // 连接池大小
	})

	err := rc.Ping(context.Background()).Err()
	if err != nil {
		return err
	}

	pool := goredis.NewPool(rc) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	Rs = redsync.New(pool)
	return nil
}
