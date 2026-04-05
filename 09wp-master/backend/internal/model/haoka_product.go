package model

import "time"

// HaokaProduct 号卡商品套餐（来源：外部上架查询接口）
type HaokaProduct struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	CategoryID uint64 `gorm:"index;not null" json:"category_id"`

	// 外部 productID：只需要唯一索引即可，避免重复定义导致 (`product_id`,`product_id`) 这种非法索引
	ProductID   uint64 `gorm:"uniqueIndex;not null" json:"product_id"`
	ProductName string `gorm:"size:200;not null" json:"product_name"`
	MainPic     string `gorm:"size:500;default:''" json:"main_pic"`
	Area        string `gorm:"size:120;default:''" json:"area"`
	// MySQL 不允许 TEXT/BLOB 字段设置 DEFAULT 值；否则 AutoMigrate 会建表失败。
	DisableArea string `gorm:"type:text" json:"disable_area"`
	LittlePicture string `gorm:"size:500;default:''" json:"little_picture"`
	NetAddr     string `gorm:"size:500;default:''" json:"net_addr"`

	Flag        bool   `gorm:"default:true;index" json:"flag"`
	NumberSel   int    `gorm:"default:0" json:"number_sel"`
	Operator    string `gorm:"size:30;default:'';index" json:"operator"`
	BackMoneyType string `gorm:"size:30;default:''" json:"back_money_type"`
	Taocan      string `gorm:"type:text" json:"taocan"`
	Rule        string `gorm:"type:text" json:"rule"`
	Age1        int    `gorm:"default:0" json:"age1"`
	Age2        int    `gorm:"default:0" json:"age2"`
	PriceTime   string `gorm:"size:60;default:''" json:"price_time"`

	Status      int8   `gorm:"default:1;index" json:"status"` // 1=启用 0=停用（本地显示控制）

	// SKUs 数组单独建表

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (HaokaProduct) TableName() string {
	return "haoka_products"
}

