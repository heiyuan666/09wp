package model

import "time"

// HaokaCategory 号卡运营商分类（电信/移动/联通）
type HaokaCategory struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Slug      string    `gorm:"size:80;not null;uniqueIndex" json:"slug"`
	Status    int8      `gorm:"default:1;index" json:"status"` // 1=启用 0=禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (HaokaCategory) TableName() string {
	return "haoka_categories"
}

