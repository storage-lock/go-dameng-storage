package dameng_storage

import (
	"context"
	"database/sql"
	"strings"
	"time"

	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

// DaMengStorage 基于达梦数据库的存储引擎实现
// 达梦数据库兼容 SQL 标准，通过主键唯一约束和 WHERE 条件更新支持 CAS
type DaMengStorage struct {
	*sql_based_storage.SqlBasedStorage
	options *DaMengStorageOptions
}

var _ storage.Storage = &DaMengStorage{}

func NewDaMengStorage(ctx context.Context, options *DaMengStorageOptions) (*DaMengStorage, error) {
	if err := options.Check(); err != nil {
		return nil, err
	}

	// 达梦数据库兼容 SQL92 标准，但 NowTimestampSql 需要用达梦自己的方言（见 DaMengSqlProvider）
	baseStorageOption := sql_based_storage.NewSqlBasedStorageOptions().
		SetConnectionManager(options.ConnectionManager).
		SetSqlProvider(NewDaMengSqlProvider()).
		SetTableFullName(options.TableName)
	baseStorage, _ := sql_based_storage.NewSqlBasedStorage(baseStorageOption)

	s := &DaMengStorage{
		SqlBasedStorage: baseStorage,
		options:         options,
	}

	if err := s.Init(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

const StorageName = "dameng-storage"

func (x *DaMengStorage) GetName() string {
	return StorageName
}

func (x *DaMengStorage) Capabilities() []storage.StorageCapability {
	return []storage.StorageCapability{
		storage.CapabilityCAS,
		storage.CapabilityReliableTime,
		storage.CapabilityWatch,
	}
}

func (x *DaMengStorage) Watch(ctx context.Context, lockId string) (<-chan *storage.LockInformation, error) {
	return storage.NewPollingLockWatcher(x, 500*time.Millisecond).Watch(ctx, lockId)
}

// CreateWithVersion 达梦唯一约束冲突时返回 ErrLockAlreadyExists
func (x *DaMengStorage) CreateWithVersion(ctx context.Context, lockId string, version storage.Version, lockInformation *storage.LockInformation) error {
	err := x.SqlBasedStorage.CreateWithVersion(ctx, lockId, version, lockInformation)
	if err != nil {
		// 达梦唯一约束冲突错误: 违反唯一约束条件
		msg := err.Error()
		if strings.Contains(msg, "违反唯一约束") || strings.Contains(msg, "unique constraint") {
			return storage_lock.ErrLockAlreadyExists
		}
	}
	return err
}

// DaMengStorageOptions 达梦存储选项
type DaMengStorageOptions struct {
	ConnectionManager storage.ConnectionManager[*sql.DB]
	TableName         string
}

func NewDaMengStorageOptions() *DaMengStorageOptions {
	return &DaMengStorageOptions{
		TableName: storage.DefaultStorageTableName,
	}
}

func (x *DaMengStorageOptions) SetConnectionManager(cm storage.ConnectionManager[*sql.DB]) *DaMengStorageOptions {
	x.ConnectionManager = cm
	return x
}

func (x *DaMengStorageOptions) SetTableName(name string) *DaMengStorageOptions {
	x.TableName = name
	return x
}

func (x *DaMengStorageOptions) Check() error {
	if x.ConnectionManager == nil {
		return ErrConnectionManagerNil
	}
	if x.TableName == "" {
		return ErrTableNameEmpty
	}
	return nil
}
