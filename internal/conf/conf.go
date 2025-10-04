package conf

import (
	"fmt"
	"go-one/internal/cache"
	"go-one/internal/model"
	"go-one/internal/service"
	"go-one/util"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
)

// Init 初始化配置项，使用默认的 .env 文件
func Init() {
	InitWithEnvFile(".env")
}

// InitWithEnvFile 使用指定的环境文件初始化配置项
func InitWithEnvFile(envFile string) {
	// 从指定文件读取环境变量
	if err := godotenv.Load(envFile); err != nil {
		util.Log().Panic("加载 %s 文件失败: %s", envFile, err.Error())
	}

	// 先设置时区，避免日志中使用时间时 Location 为空
	tz := os.Getenv("SERVER_TIMEZONE")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	location, err := time.LoadLocation(tz)
	if err != nil {
		// 如果在程序启动时就无法加载时区，这通常是一个严重的环境配置问题
		// 在这种情况下，直接panic 让程序崩溃并暴露问题，是比静默失败更好的选择
		log.Fatalf("无法加载 '%s' 时区: %v", tz, err)
	}
	time.Local = location
	util.Location = location

	// 设置日志级别（在 Location 设置之后再构建，避免时间格式化使用空 Location）
	util.BuildLogger(os.Getenv("LOG_LEVEL"))

	// 初始化Redis
	if err := cache.InitRedis(); err != nil {
		util.Log().Panic("初始化Redis失败: %v", err)
	}

	// 连接数据库
	// 从环境变量获取数据库连接信息
	url := os.Getenv("POSTGRES_URL")
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	port := os.Getenv("DB_PORT")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("%s%s:%s@%s:%s/%s", url, user, password, host, port, dbname)

	model.Init(dsn, tz)

	// 初始化JWT配置
	service.InitJWT()

	// 初始化 Sentry（可选）
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		tracesRate := 0.0
		if v, err := strconv.ParseFloat(os.Getenv("SENTRY_TRACES_SAMPLE_RATE"), 64); err == nil {
			tracesRate = v
		}
		profilesRate := 0.0
		if v, err := strconv.ParseFloat(os.Getenv("SENTRY_PROFILES_SAMPLE_RATE"), 64); err == nil {
			profilesRate = v
		}
		_ = sentry.Init(sentry.ClientOptions{
			Dsn:                dsn,
			Environment:        os.Getenv("SENTRY_ENVIRONMENT"),
			Release:            os.Getenv("SENTRY_RELEASE"),
			AttachStacktrace:   true,
			EnableTracing:      tracesRate > 0,
			TracesSampleRate:   tracesRate,
			ProfilesSampleRate: profilesRate,
		})
		util.Log().Info("Sentry 初始化完成")
	}
}
