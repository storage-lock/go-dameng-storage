package dameng_storage

import (
	"context"

	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
)

// DaMengSqlProvider 达梦数据库的 SQL 方言提供者
// 达梦兼容 Oracle 语法（支持 DUAL 表、SYSDATE），但不支持 MySQL 的 UNIX_TIMESTAMP(NOW())。
type DaMengSqlProvider struct {
	*sql_based_storage.Sql92Provider
}

func NewDaMengSqlProvider() *DaMengSqlProvider {
	return &DaMengSqlProvider{
		Sql92Provider: sql_based_storage.NewSql92Provider(),
	}
}

// NowTimestampSql 达梦获取当前时间的 Unix 时间戳
// 修复：原来直接继承 Sql92Provider 的 "SELECT UNIX_TIMESTAMP(NOW())"，达梦不支持这两个 MySQL
// 专有函数，执行即报语法错误，导致 GetTime 完全不可用（GetTime 是续租/抢占等核心路径的依赖）。
// 参照 go-oracle-storage 的同类修复：用 SYSDATE 与 1970-01-01 的天数差 * 86400 换算为 Unix 秒数。
// 注意：未接入真实达梦驱动验证，是按达梦兼容 Oracle 语法文档做的最佳努力实现。
func (x *DaMengSqlProvider) NowTimestampSql(ctx context.Context, tableFullName string) (string, []any) {
	return `SELECT CAST((SYSDATE - TO_DATE('1970-01-01', 'YYYY-MM-DD')) * 86400 AS BIGINT) FROM DUAL`, nil
}
