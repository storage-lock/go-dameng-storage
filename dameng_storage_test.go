package dameng_storage

import (
	"context"
	"os"
	"testing"

	storage_test_helper "github.com/storage-lock/go-storage-test-helper"
	"github.com/stretchr/testify/assert"
)

func TestDaMengStorage(t *testing.T) {
	dsn := os.Getenv("STORAGE_LOCK_DAMENG_DSN")
	if dsn == "" {
		t.Skip("STORAGE_LOCK_DAMENG_DSN not set, skipping DaMeng test")
	}

	ctx := context.Background()
	cm := NewDaMengConnectionManagerFromDSN(dsn)
	s, err := NewDaMengStorage(ctx, &DaMengStorageOptions{
		ConnectionManager: cm,
		TableName:         storage_test_helper.TestTableName,
	})
	assert.Nil(t, err)
	storage_test_helper.TestStorage(t, s)
}
