package model

import (
	"go-one/util"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 数据库连接单例
var DB *gorm.DB

// Init 初始化数据库连接
func Init(dsn, tz string) {
	database(dsn, tz)
}

// database 在中间件中初始化 postgres 链接
func database(connString, tz string) {
	// 初始化GORM日志配置
	newLogger := logger.New(
		log.New(util.LogWriter(), "\r\n", log.LstdFlags), // 复用项目日志 writer，持久化+分片
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	// 使用 postgres driver
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{
		Logger: newLogger,
		NowFunc: func() time.Time {
			return util.GetCurrentTime()
		},
	})
	// Error
	if connString == "" || err != nil {
		util.Log().Error("postgres连接失败: %v", err)
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		util.Log().Error("postgres错误: %v", err)
		panic(err)
	}

	// 设置连接池
	// 空闲
	sqlDB.SetMaxIdleConns(10)
	// 打开
	sqlDB.SetMaxOpenConns(20)
	DB = db

	// 设置数据库会话时区（默认 Asia/Shanghai，可通过 DB_TIMEZONE 配置）
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	// 简单转义单引号，避免语法错误
	safeTZ := strings.ReplaceAll(tz, "'", "''")
	if err := db.Exec("SET TIME ZONE '" + safeTZ + "'").Error; err != nil {
		util.Log().Warning("设置数据库时区失败，tz=%s, err=%v", tz, err)
	}

	migration()
	util.Log().Info("数据库连接成功")

}
