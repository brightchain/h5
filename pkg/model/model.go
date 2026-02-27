package model

import (
	"errors"
	"fmt"
	"h5/pkg/config"
	"log"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 连接管理器
type RDBManager struct {
	OpenTx bool     // 是否开启事务
	DsName string   // 数据源名称
	Db     *gorm.DB // 非事务实例
	Tx     *gorm.Tx // 事务实例
	Errors []error  // 操作过程中记录的错误
}

// db连接
var (
	MASTER = "db"                     // 默认主数据源
	RDB    = map[string]*RDBManager{} // 初始化时加载数据源到集合
)

func InitDb() {
	for k := range config.GetStringMap("databases") {
		connByConf(k)
	}

}

func connByConf(key string) {
	db, err := connectDB(key)
	if err != nil {
		slog.Error("数据库链接失败!", err)
		return
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	rdb := &RDBManager{
		Db:     db,
		DsName: key,
	}
	RDB[key] = rdb
}

func connectDB(key string) (*gorm.DB, error) {
	filename := config.GetString("logger.gormName", "./storage/logs/gorm.log")
	level := config.GetString("logger.level")
	logOps := logger.Config{
		SlowThreshold:             time.Second, // Slow SQL threshold
		LogLevel:                  logger.Info, // Log level(这里记得根据需求改一下)
		IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
		Colorful:                  false,       // Disable color
	}
	switch level {
	case "debug":
		logOps.LogLevel = logger.Info
	case "info":
		logOps.LogLevel = logger.Info
	case "warn":
		logOps.LogLevel = logger.Warn
	case "error":
		logOps.LogLevel = logger.Error
	default:
		logOps.LogLevel = logger.Warn
	}
	// 初始化GORM日志配置
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("logger.Setup err: %v", err)
	}
	newLogger := logger.New(log.New(f, "\r\n", log.LstdFlags), logOps)
	configPath := "databases." + key
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&loc=Local",
		config.GetString(configPath+".username", "root"),
		config.GetString(configPath+".password", ""),
		config.GetString(configPath+".host", "127.0.0.1"),
		config.GetString(configPath+".port", "3306"),
		config.GetString(configPath+".database", ""),
		config.GetString(configPath+".charset", "utf8mb4"))
	//fmt.Printf("数据库%v,%v", key, dsn)
	gormConfig := mysql.New(mysql.Config{
		DSN: dsn,
	})
	conn, err := gorm.Open(gormConfig, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		//slog.Error("数据库连接失败", err)
		return nil, errors.New("数据库连接失败")
	}

	return conn, nil
}
