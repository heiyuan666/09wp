package service

import (
	"errors"
	"log"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"gorm.io/gorm"
)

// StartHaokaSyncWorker 启动号卡定时同步任务
func StartHaokaSyncWorker() {
	go func() {
		var lastRun time.Time
		for {
			var cfg model.SystemConfig
			if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
				time.Sleep(30 * time.Second)
				continue
			}

			if !cfg.HaokaSyncEnabled {
				time.Sleep(30 * time.Second)
				continue
			}
			if strings.TrimSpace(cfg.HaokaUserID) == "" || strings.TrimSpace(cfg.HaokaSecret) == "" {
				time.Sleep(30 * time.Second)
				continue
			}

			interval := cfg.HaokaSyncInterval
			if interval < 300 {
				interval = 300
			}
			if !lastRun.IsZero() && time.Since(lastRun) < time.Duration(interval)*time.Second {
				time.Sleep(30 * time.Second)
				continue
			}

			saved, updated, skus, total, err := SyncHaokaFromRemote(cfg.HaokaUserID, cfg.HaokaSecret, "")
			lastRun = time.Now()
			if err != nil {
				log.Printf("haoka auto sync failed: %v", err)
			} else {
				log.Printf("haoka auto sync done: total=%d saved=%d updated=%d skus=%d", total, saved, updated, skus)
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

func ensureHaokaCategory(op string) (model.HaokaCategory, error) {
	slug := OperatorCategorySlug(op)
	name := strings.TrimSpace(op)
	if slug == "unknown" || name == "" {
		return model.HaokaCategory{}, errors.New("unknown operator")
	}
	var cat model.HaokaCategory
	err := database.DB().Where("name = ? OR slug = ?", name, slug).First(&cat).Error
	if err == nil {
		return cat, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.HaokaCategory{}, err
	}
	cat = model.HaokaCategory{Name: name, Slug: slug, Status: 1}
	if err := database.DB().Create(&cat).Error; err != nil {
		return model.HaokaCategory{}, err
	}
	return cat, nil
}

// SyncHaokaFromRemote 从远端接口拉取并落库（供手动/定时复用）
func SyncHaokaFromRemote(userID, secret, productID string) (saved, updated, skusSaved, total int, err error) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})

	products, err := HaokaQueryProducts(userID, secret, productID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	total = len(products)

	for _, p := range products {
		cat, cErr := ensureHaokaCategory(p.Operator)
		if cErr != nil {
			continue
		}

		var exist model.HaokaProduct
		qErr := database.DB().Where("product_id = ?", p.ProductID).First(&exist).Error
		if qErr != nil && !errors.Is(qErr, gorm.ErrRecordNotFound) {
			continue
		}

		if errors.Is(qErr, gorm.ErrRecordNotFound) {
			np := model.HaokaProduct{
				CategoryID:    cat.ID,
				ProductID:     p.ProductID,
				ProductName:   p.ProductName,
				MainPic:       p.MainPic,
				Area:          p.Area,
				DisableArea:   p.DisableArea,
				LittlePicture: p.LittlePicture,
				NetAddr:       p.NetAddr,
				Flag:          p.Flag,
				NumberSel:     p.NumberSel,
				Operator:      p.Operator,
				BackMoneyType: p.BackMoneyType,
				Taocan:        p.Taocan,
				Rule:          p.Rule,
				Age1:          p.Age1,
				Age2:          p.Age2,
				PriceTime:     p.PriceTime,
				Status:        1,
			}
			if err := database.DB().Create(&np).Error; err == nil {
				saved++
			}
		} else {
			if err := database.DB().Model(&model.HaokaProduct{}).Where("id = ?", exist.ID).Updates(map[string]any{
				"category_id":     cat.ID,
				"product_name":    p.ProductName,
				"main_pic":        p.MainPic,
				"area":            p.Area,
				"disable_area":    p.DisableArea,
				"little_picture":  p.LittlePicture,
				"net_addr":        p.NetAddr,
				"flag":            p.Flag,
				"number_sel":      p.NumberSel,
				"operator":        p.Operator,
				"back_money_type": p.BackMoneyType,
				"taocan":          p.Taocan,
				"rule":            p.Rule,
				"age1":            p.Age1,
				"age2":            p.Age2,
				"price_time":      p.PriceTime,
				"status":          1,
				"updated_at":      time.Now(),
			}).Error; err == nil {
				updated++
			}
		}

		_ = database.DB().Where("product_id = ?", p.ProductID).Delete(&model.HaokaSku{}).Error
		for _, s := range p.Skus {
			ns := model.HaokaSku{
				ProductID: p.ProductID,
				SkuID:     s.SkuID,
				SkuName:   s.SkuName,
				Desc:      s.Desc,
			}
			if err := database.DB().Create(&ns).Error; err == nil {
				skusSaved++
			}
		}
	}
	return
}

