package dameng_storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/storage-lock/go-storage"
)

// DaMengConnectionManager 达梦数据库连接管理器
type DaMengConnectionManager struct {
	dsn      string
	Host     string
	Port     int
	Database string
	Username string
	Password string

	mu         sync.Mutex
	db         *sql.DB
	initialized bool
}

var _ storage.ConnectionManager[*sql.DB] = &DaMengConnectionManager{}

// NewDaMengConnectionManagerFromDSN 从数据源名称创建连接管理器
func NewDaMengConnectionManagerFromDSN(dsn string) *DaMengConnectionManager {
	return &DaMengConnectionManager{dsn: dsn}
}

// NewDaMengConnectionManagerFromConfig 从配置参数创建连接管理器
func NewDaMengConnectionManagerFromConfig(host string, port int, database, username, password string) *DaMengConnectionManager {
	return &DaMengConnectionManager{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
	}
}

func (x *DaMengConnectionManager) Name() string {
	return "dameng-connection-manager"
}

func (x *DaMengConnectionManager) Take(ctx context.Context) (*sql.DB, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.initialized {
		return x.db, nil
	}

	// 达梦 DSN 格式: dm://username:password@host:port/database
	var dsn string
	if x.dsn != "" {
		dsn = x.dsn
	} else {
		dsn = fmt.Sprintf("dm://%s:%s@%s:%d/%s",
			x.Username, x.Password, x.Host, x.Port, x.Database)
	}

	db, err := sql.Open("dm", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	x.db = db
	x.initialized = true
	return x.db, nil
}

func (x *DaMengConnectionManager) Return(ctx context.Context, conn *sql.DB) error {
	return nil
}

func (x *DaMengConnectionManager) Shutdown(ctx context.Context) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.db != nil {
		return x.db.Close()
	}
	return nil
}