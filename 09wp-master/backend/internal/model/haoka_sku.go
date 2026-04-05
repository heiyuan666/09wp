package model

import "time"

// HaokaSku 外部商品套餐的办理区域/sku 信息
type HaokaSku struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	ProductID uint64 `gorm:"index;not null" json:"product_id"`
	SkuID     uint64 `gorm:"not null;index" json:"sku_id"`
	SkuName   string `gorm:"size:120;default:''" json:"sku_name"`
	// MySQL 不允许 TEXT/BLOB 字段设置 DEFAULT 值；否则 AutoMigrate 会建表失败。
	Desc      string `gorm:"type:text" json:"desc"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (HaokaSku) TableName() string {
	return "haoka_skus"
}

