package mysql

import (
	"alivePlatform/order_service/config"

	"fmt"
	"gorm.io/gorm/logger"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// https://gorm.io/zh_CN/docs/connecting_to_the_database.html

var db *gorm.DB

// Init 初始化MySQL连接
func Init(cfg *config.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Println("failed to connect database")
		return err
	}

	// 额外的连接配置
	sqlDB, err := db.DB() // database/sql.DB, 为标准库的原生DB，由此可设置最大连接数等配置信息
	if err != nil {
		return
	}

	// 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// 设置连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return
}
