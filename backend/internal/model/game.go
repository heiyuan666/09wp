package model

import "time"

// GameCategory 游戏分类
type GameCategory struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Slug        string    `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Game 游戏信息
type Game struct {
	ID                uint64     `gorm:"primaryKey" json:"id"`
	CategoryID        *uint64    `gorm:"index" json:"category_id,omitempty"`
	SteamAppID        uint64     `gorm:"index;default:0" json:"steam_appid"`
	RequiredAge       int        `gorm:"default:0" json:"required_age"`
	IsFree            bool       `gorm:"default:false" json:"is_free"`
	Title             string     `gorm:"size:200;not null;index" json:"title"`
	Cover             string     `gorm:"size:500" json:"cover"`
	Banner            string     `gorm:"size:500" json:"banner"`
	VideoURL          string     `gorm:"size:500;default:''" json:"video_url"`
	ShortDescription  string     `gorm:"type:text" json:"short_description"`
	SupportedLangs    string     `gorm:"type:text" json:"supported_languages"`
	Reviews           string     `gorm:"type:longtext" json:"reviews"`
	PCRequirements    string     `gorm:"type:longtext" json:"pc_requirements"`
	MacRequirements   string     `gorm:"type:longtext" json:"mac_requirements"`
	LinuxRequirements string     `gorm:"type:longtext" json:"linux_requirements"`
	HeaderImage       string     `gorm:"size:500" json:"header_image"`
	Website           string     `gorm:"size:500" json:"website"`
	Developers        string     `gorm:"type:text" json:"developers"`
	Publishers        string     `gorm:"type:text" json:"publishers"`
	Platforms         string     `gorm:"size:120" json:"platforms"`
	Genres            string     `gorm:"type:text" json:"genres"`
	Tags              string     `gorm:"type:text" json:"tags"`
	PriceText         string     `gorm:"size:120" json:"price_text"`
	PriceCurrency     string     `gorm:"size:16" json:"price_currency"`
	PriceInitial      int        `gorm:"default:0" json:"price_initial"`
	PriceFinal        int        `gorm:"default:0" json:"price_final"`
	PriceDiscount     int        `gorm:"default:0" json:"price_discount"`
	MetacriticScore   int        `gorm:"default:0" json:"metacritic_score"`
	Description       string     `gorm:"type:longtext" json:"description"`
	ReleaseDate       *time.Time `json:"release_date,omitempty"`
	Size              string     `gorm:"size:50" json:"size"`
	Type              string     `gorm:"size:100" json:"type"`
	Developer         string     `gorm:"size:150" json:"developer"`
	Rating            float64    `gorm:"type:decimal(3,1);default:0" json:"rating"`
	SteamScore        int        `gorm:"default:0" json:"steam_score"`
	Recommendations   uint64     `gorm:"default:0" json:"recommendations_total"`
	Downloads         uint64     `gorm:"default:0" json:"downloads"`
	Likes             uint64     `gorm:"default:0" json:"likes"`
	Dislikes          uint64     `gorm:"default:0" json:"dislikes"`
	// Gallery 以 JSON 数组字符串存储，例如 ["url1","url2"]
	Gallery   string    `gorm:"type:longtext" json:"gallery"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GameResource 游戏下载资源
type GameResource struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	GameID       uint64     `gorm:"index;not null" json:"game_id"`
	Title        string     `gorm:"size:200;not null" json:"title"`
	ResourceType string     `gorm:"size:30;not null;default:'game';index" json:"resource_type"`
	Version      string     `gorm:"size:80" json:"version"`
	Size         string     `gorm:"size:50" json:"size"`
	DownloadType string     `gorm:"size:80" json:"download_type"`
	PanType      string     `gorm:"size:50;index" json:"pan_type"`
	DownloadURL  string     `gorm:"size:1000;not null" json:"download_url"`
	ExtractCode  string     `gorm:"size:50" json:"extract_code"`
	Tested       bool       `gorm:"default:false" json:"tested"`
	Author       string     `gorm:"size:100" json:"author"`
	PublishDate  *time.Time `json:"publish_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// GameReview 游戏评论（前台用户发布）
type GameReview struct {
	ID      uint64 `gorm:"primaryKey" json:"id"`
	GameID  uint64 `gorm:"index;not null" json:"game_id"`
	UserID  uint64 `gorm:"index;not null" json:"user_id"`
	Rating  int    `gorm:"default:0" json:"rating"` // 1~5，允许 0 表示仅文字
	Content string `gorm:"type:text;not null" json:"content"`
	Status  int8   `gorm:"default:1;index" json:"status"` // 1=展示 0=隐藏（预留审核/屏蔽）
	HelpfulCount   uint64 `gorm:"default:0" json:"helpful_count"`
	UnhelpfulCount uint64 `gorm:"default:0" json:"unhelpful_count"`

	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GameReviewVote 评论投票（每用户每评论唯一，用于“最有帮助”排序）
type GameReviewVote struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	ReviewID uint64 `gorm:"index;not null;uniqueIndex:uniq_review_user" json:"review_id"`
	UserID   uint64 `gorm:"index;not null;uniqueIndex:uniq_review_user" json:"user_id"`
	Vote     int8   `gorm:"default:0" json:"vote"` // 1=helpful -1=unhelpful 0=neutral
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
