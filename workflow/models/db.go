package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
	dbErr  error
)

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
}

func InitDB(cfg DBConfig) error {
	dbOnce.Do(func() {
		level := cfg.LogLevel
		if level == 0 {
			level = logger.Warn
		}
		gormDB, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(level),
		})
		if err != nil {
			dbErr = fmt.Errorf("open mysql failed: %w", err)
			return
		}
		sqlDB, err := gormDB.DB()
		if err != nil {
			dbErr = fmt.Errorf("get sql.DB failed: %w", err)
			return
		}
		if cfg.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		}
		if cfg.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		}
		if cfg.ConnMaxLifetime > 0 {
			sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		}
		db = gormDB
	})
	return dbErr
}

func DB() *gorm.DB {
	return db
}

func SetDB(d *gorm.DB) {
	db = d
}

// Init 加载 conf/<env>/app.toml，初始化全局 *gorm.DB 并 ping 一次。
// 失败返回 error，调用方（init.go / main）自行决定 panic 还是降级。
func Init() error {
	cfg, err := LoadConfig("")
	if err != nil {
		return fmt.Errorf("models.Init load config: %w", err)
	}
	mc := cfg.Mysql
	if mc.DSN == "" {
		return fmt.Errorf("models.Init: mysql.dsn is empty")
	}
	if err := InitDB(DBConfig{
		DSN:             mc.DSN,
		MaxOpenConns:    mc.MaxOpenConns,
		MaxIdleConns:    mc.MaxIdleConns,
		ConnMaxLifetime: time.Duration(mc.ConnMaxLifetimeSeconds) * time.Second,
	}); err != nil {
		return fmt.Errorf("models.Init open: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("models.Init get sql.DB: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("models.Init ping: %w", err)
	}
	return nil
}

// MustInit 同 Init，失败 panic（用于进程启动期）。
func MustInit() {
	if err := Init(); err != nil {
		panic(err)
	}
}
