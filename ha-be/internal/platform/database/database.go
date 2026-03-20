package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(url string) (*gorm.DB, error) {
	if strings.TrimSpace(url) == "" {
		return nil, nil
	}

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func WithTx(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

type RawQueryResult[T any] struct {
	Items []T
}

func RawQuery[T any](db *gorm.DB, sql string, args ...any) ([]T, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}
	var result []T
	if err := db.Raw(sql, args...).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
