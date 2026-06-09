package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

type MysqlConf struct {
	DSN                    string `toml:"dsn"`
	DisfEnable             bool   `toml:"disf_enable"`
	ConnectTimeout         int    `toml:"connect_timeout"`
	MetricsEnable          bool   `toml:"metrics_enable"`
	MaxOpenConns           int    `toml:"max_open_conns"`
	MaxIdleConns           int    `toml:"max_idle_conns"`
	ConnMaxLifetimeSeconds int    `toml:"conn_max_lifetime_seconds"`
}

type AppConf struct {
	Mysql MysqlConf `toml:"mysql"`
}

var (
	cfg     AppConf
	cfgOnce sync.Once
	cfgErr  error
)

// Load 读取 conf/<env>/app.toml；env 取自参数 / 环境变量 APP_ENV，默认 dev。
func Load(env string) (*AppConf, error) {
	cfgOnce.Do(func() {
		if env == "" {
			env = os.Getenv("APP_ENV")
		}
		if env == "" {
			env = "dev"
		}
		path := resolveConfPath(env)
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			cfgErr = fmt.Errorf("load config %s failed: %w", path, err)
			return
		}
	})
	if cfgErr != nil {
		return nil, cfgErr
	}
	return &cfg, nil
}

func Get() *AppConf {
	return &cfg
}

func resolveConfPath(env string) string {
	rel := filepath.Join("conf", env, "app.toml")
	if _, err := os.Stat(rel); err == nil {
		return rel
	}
	// 兜底：根据可执行文件位置向上回退查找
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for range 5 {
			p := filepath.Join(dir, "conf", env, "app.toml")
			if _, err := os.Stat(p); err == nil {
				return p
			}
			dir = filepath.Dir(dir)
		}
	}
	return rel
}
