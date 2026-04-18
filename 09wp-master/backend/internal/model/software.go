package model

import "time"

// SoftwareCategory 软件分类
type SoftwareCategory struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Slug      string    `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	SortOrder int       `gorm:"default:0;index" json:"sort_order"`
	Status    int8      `gorm:"default:1;index" json:"status"` // 1=启用 0=禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Software 软件主体
type Software struct {
	ID                uint64         `gorm:"primaryKey" json:"id"`
	CategoryID        uint64         `gorm:"index;not null" json:"category_id"`
	Name              string         `gorm:"size:200;not null;index" json:"name"`
	Summary           string         `gorm:"type:text" json:"summary"`
	Version           string         `gorm:"size:80;default:''" json:"version"`
	Cover             string         `gorm:"size:500;default:''" json:"cover"`
	CoverThumb        string         `gorm:"size:500;default:''" json:"cover_thumb"`
	Icon              string         `gorm:"size:500;default:''" json:"icon"`        // 软件小图标（原图 URL）
	IconThumb         string         `gorm:"size:500;default:''" json:"icon_thumb"` // 列表用缩略图
	Screenshots       JSONStringList `gorm:"type:text;column:screenshots" json:"screenshots"`
	Size              string         `gorm:"size:50;default:''" json:"size"`
	Platforms         string         `gorm:"size:255;default:''" json:"platforms"` // 逗号分隔
	Website           string         `gorm:"size:500;default:''" json:"website"`
	DownloadDirect    JSONStringList `gorm:"type:text;column:download_direct" json:"download_direct"`
	DownloadPan       JSONStringList `gorm:"type:text;column:download_pan" json:"download_pan"`
	DownloadExtract   string         `gorm:"size:100;default:''" json:"download_extract"`
	PublishedAt       *time.Time     `json:"published_at,omitempty"`
	UpdatedAtOverride *time.Time     `json:"updated_at_override,omitempty"` // 业务更新时间
	Status            int8           `gorm:"default:1;index" json:"status"` // 1=上架 0=下架
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// SoftwareVersion 软件版本
type SoftwareVersion struct {
	ID              uint64         `gorm:"primaryKey" json:"id"`
	SoftwareID      uint64         `gorm:"index;not null" json:"software_id"`
	Version         string         `gorm:"size:80;not null;index" json:"version"`
	ReleaseNotes    string         `gorm:"type:text" json:"release_notes"`
	PublishedAt     *time.Time     `json:"published_at,omitempty"`
	DownloadDirect  JSONStringList `gorm:"type:text;column:download_direct" json:"download_direct"`
	DownloadPan     JSONStringList `gorm:"type:text;column:download_pan" json:"download_pan"`
	DownloadExtract string         `gorm:"size:100;default:''" json:"download_extract"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
