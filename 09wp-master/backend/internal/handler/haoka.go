package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type haokaQueryProductsReq struct {
	UserID    string `json:"user_id" binding:"required"`
	Secret    string `json:"secret" binding:"required"`
	ProductID string `json:"product_id"` // 可为空：空代表返回所有上架商品
}

type haokaSyncReq struct {
	UserID string `json:"user_id" binding:"required"`
	Secret string `json:"secret" binding:"required"`
	// ProductID 可选：空代表同步全部
	ProductID string `json:"product_id"`
}

type haokaSkuReq struct {
	SkuID     uint64 `json:"sku_id"`
	SkuName   string `json:"sku_name"`
	Desc      string `json:"desc"`
}

type haokaProductUpdateReq struct {
	// 手动新增时需要填 product_id
	ProductID     uint64        `json:"product_id"`
	CategoryID    uint64        `json:"category_id"`
	Operator      string        `json:"operator"`
	ProductName   string        `json:"product_name"`
	MainPic       string        `json:"main_pic"`
	Area          string        `json:"area"`
	DisableArea   string        `json:"disable_area"`
	LittlePicture string        `json:"little_picture"`
	NetAddr       string        `json:"net_addr"`
	Flag          *bool         `json:"flag"` // 可选：不传则不更新
	NumberSel     int           `json:"number_sel"`
	BackMoneyType string        `json:"back_money_type"`
	Taocan        string        `json:"taocan"`
	Rule          string        `json:"rule"`
	Age1          int           `json:"age1"`
	Age2          int           `json:"age2"`
	PriceTime     string        `json:"price_time"`
	SKUs          []haokaSkuReq `json:"skus"` // 可选：不传则不更新 sku；传空则清空
}

func haokaEnsureCategory(op string) (model.HaokaCategory, error) {
	slug := service.OperatorCategorySlug(op)
	name := strings.TrimSpace(op)
	if name == "" {
		name = op
	}
	if slug == "unknown" {
		return model.HaokaCategory{}, errors.New("未知运营商分类")
	}

	var cat model.HaokaCategory
	err := database.DB().Where("name = ? OR slug = ?", name, slug).First(&cat).Error
	if err == nil {
		return cat, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.HaokaCategory{}, err
	}

	cat = model.HaokaCategory{
		Name:   name,
		Slug:   slug,
		Status: 1,
	}
	if err := database.DB().Create(&cat).Error; err != nil {
		return model.HaokaCategory{}, err
	}
	return cat, nil
}

func mapExternalToProduct(p service.HaokaExternalProduct, catID uint64) model.HaokaProduct {
	return model.HaokaProduct{
		CategoryID:    catID,
		ProductID:     p.ProductID,
		ProductName:   p.ProductName,
		MainPic:       p.MainPic,
		Area:          p.Area,
		DisableArea:   p.DisableArea,
		LittlePicture: p.LittlePicture,
		NetAddr:       p.NetAddr,
		Flag:          p.Flag,
		NumberSel:    p.NumberSel,
		Operator:     p.Operator,
		BackMoneyType: p.BackMoneyType,
		Taocan:       p.Taocan,
		Rule:         p.Rule,
		Age1:         p.Age1,
		Age2:         p.Age2,
		PriceTime:    p.PriceTime,
		Status:       1,
	}
}

// AdminHaokaCategories 获取分类（电信/移动/联通）
func AdminHaokaCategories(c *gin.Context) {
	// 兜底：避免首次引入代码后表未创建导致 1146
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})
	// 兜底：确保分类存在（电信/移动/联通）
	_, _ = haokaEnsureCategory("电信")
	_, _ = haokaEnsureCategory("移动")
	_, _ = haokaEnsureCategory("联通")

	var cats []model.HaokaCategory
	if err := database.DB().Where("status = 1").Order("id ASC").Find(&cats).Error; err != nil {
		response.Error(c, 500, "分类查询失败")
		return
	}
	response.OK(c, cats)
}

// AdminHaokaQueryProducts 对接“产品上架查询接口”
func AdminHaokaQueryProducts(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})
	var req haokaQueryProductsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	products, err := service.HaokaQueryProducts(req.UserID, req.Secret, req.ProductID)
	if err != nil {
		response.Error(c, -1, err.Error())
		return
	}

	// 前端展示需要稳定顺序
	//（服务层若不稳定，这里兜底）
	response.OK(c, gin.H{"list": products})
}

// AdminHaokaSync 同步外部上架商品并落库（upsert）
func AdminHaokaSync(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})
	var req haokaSyncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	saved, updated, skusSaved, total, err := service.SyncHaokaFromRemote(req.UserID, req.Secret, req.ProductID)
	if err != nil {
		response.Error(c, -1, err.Error())
		return
	}

	response.OK(c, gin.H{
		"saved":       saved,
		"updated":     updated,
		"skus_saved":  skusSaved,
		"total":       total,
	})
}

// AdminHaokaProductList 获取本地落库的号卡产品
func AdminHaokaProductList(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})
	// 支持可选 operator/category_id/flag + 分页
	categoryID := strings.TrimSpace(c.Query("category_id"))
	operator := strings.TrimSpace(c.Query("operator"))
	flag := strings.TrimSpace(c.Query("flag")) // "true"/"false"

	query := database.DB().Model(&model.HaokaProduct{}).
		Where("status = 1")
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if operator != "" {
		query = query.Where("operator = ?", operator)
	}
	if flag != "" {
		switch flag {
		case "true":
			query = query.Where("flag = ?", true)
		case "false":
			query = query.Where("flag = ?", false)
		}
	}

	// 分页
	page := 1
	pageSize := 20
	if v := strings.TrimSpace(c.DefaultQuery("page", "1")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.DefaultQuery("page_size", "20")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 100 {
				n = 100
			}
			pageSize = n
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, 500, "列表查询失败")
		return
	}

	var list []model.HaokaProduct
	if err := query.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		response.Error(c, 500, "列表查询失败")
		return
	}

	// 带分类名
	type item struct {
		model.HaokaProduct
		CategoryName string `json:"category_name"`
	}
	items := make([]item, 0, len(list))
	for _, p := range list {
		var cat model.HaokaCategory
		_ = database.DB().Where("id = ?", p.CategoryID).First(&cat).Error
		items = append(items, item{
			HaokaProduct:  p,
			CategoryName: cat.Name,
		})
	}

	response.OK(c, gin.H{
		"list":      items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdminHaokaProductDetail 号卡商品详情（含 skus）
func AdminHaokaProductDetail(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})

	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		response.Error(c, 400, "id 不能为空")
		return
	}

	var p model.HaokaProduct
	if err := database.DB().Where("id = ? AND status = 1", id).First(&p).Error; err != nil {
		response.Error(c, 404, "号卡不存在")
		return
	}

	var cat model.HaokaCategory
	_ = database.DB().Where("id = ?", p.CategoryID).First(&cat).Error

	var skus []model.HaokaSku
	_ = database.DB().Where("product_id = ?", p.ProductID).Order("id ASC").Find(&skus).Error

	response.OK(c, gin.H{
		"product":       p,
		"category_name": cat.Name,
		"skus":          skus,
	})
}

// AdminHaokaProductUpdate 编辑商品/字段/上架状态，并可更新 sku 列表
func AdminHaokaProductUpdate(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})

	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		response.Error(c, 400, "id 不能为空")
		return
	}

	var req haokaProductUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	// 基础校验
	if req.CategoryID == 0 {
		response.Error(c, 400, "category_id 不能为空")
		return
	}
	if strings.TrimSpace(req.ProductName) == "" {
		response.Error(c, 400, "product_name 不能为空")
		return
	}

	// 更新 product
	updateMap := map[string]any{
		"category_id":     req.CategoryID,
		"operator":        req.Operator,
		"product_name":    req.ProductName,
		"main_pic":        req.MainPic,
		"area":            req.Area,
		"disable_area":    req.DisableArea,
		"little_picture":  req.LittlePicture,
		"net_addr":        req.NetAddr,
		"number_sel":      req.NumberSel,
		"back_money_type": req.BackMoneyType,
		"taocan":          req.Taocan,
		"rule":            req.Rule,
		"age1":            req.Age1,
		"age2":            req.Age2,
		"price_time":      req.PriceTime,
		"updated_at":      time.Now(),
	}
	if req.Flag != nil {
		updateMap["flag"] = *req.Flag
	}

	var p model.HaokaProduct
	if err := database.DB().Where("id = ? AND status = 1", id).First(&p).Error; err != nil {
		response.Error(c, 404, "号卡不存在")
		return
	}

	if err := database.DB().Model(&model.HaokaProduct{}).Where("id = ?", p.ID).Updates(updateMap).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	// 更新 skus（当 req.SKUs 非 nil 时才做替换；这样前端不传 SKUs 时也不会清空）
	if req.SKUs != nil {
		if err := database.DB().Where("product_id = ?", p.ProductID).Delete(&model.HaokaSku{}).Error; err != nil {
			response.Error(c, 500, "更新 sku 失败")
			return
		}
		for _, s := range req.SKUs {
			if s.SkuID == 0 {
				continue
			}
			ns := model.HaokaSku{
				ProductID: p.ProductID,
				SkuID:     s.SkuID,
				SkuName:   s.SkuName,
				Desc:      s.Desc,
			}
			_ = database.DB().Create(&ns).Error
		}
	}

	response.OK(c, gin.H{"ok": true})
}

// AdminHaokaProductSetFlag 单独切换上架状态（flag）
func AdminHaokaProductSetFlag(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})

	id := c.Param("id")
	var req struct {
		Flag bool `json:"flag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	if err := database.DB().Model(&model.HaokaProduct{}).
		Where("id = ? AND status = 1", id).
		Update("flag", req.Flag).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	response.OK(c, gin.H{"ok": true})
}

// AdminHaokaProductCreate 手动新增号卡（不依赖外部查询接口）
func AdminHaokaProductCreate(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})

	var req haokaProductUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if req.ProductID == 0 {
		response.Error(c, 400, "product_id 不能为空")
		return
	}
	if req.CategoryID == 0 {
		response.Error(c, 400, "category_id 不能为空")
		return
	}
	if strings.TrimSpace(req.ProductName) == "" {
		response.Error(c, 400, "product_name 不能为空")
		return
	}

	// 创建 product
	p := model.HaokaProduct{
		CategoryID:    req.CategoryID,
		ProductID:     req.ProductID,
		ProductName:   req.ProductName,
		MainPic:       req.MainPic,
		Area:          req.Area,
		DisableArea:   req.DisableArea,
		LittlePicture: req.LittlePicture,
		NetAddr:       req.NetAddr,
		NumberSel:     req.NumberSel,
		Operator:      req.Operator,
		BackMoneyType: req.BackMoneyType,
		Taocan:        req.Taocan,
		Rule:          req.Rule,
		Age1:          req.Age1,
		Age2:          req.Age2,
		PriceTime:     req.PriceTime,
		Status:        1,
	}
	if req.Flag != nil {
		p.Flag = *req.Flag
	} else {
		p.Flag = true
	}

	if err := database.DB().Create(&p).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}

	// 写入 skus
	for _, s := range req.SKUs {
		if s.SkuID == 0 {
			continue
		}
		ns := model.HaokaSku{
			ProductID: p.ProductID,
			SkuID:     s.SkuID,
			SkuName:   s.SkuName,
			Desc:      s.Desc,
		}
		_ = database.DB().Create(&ns).Error
	}

	response.OK(c, gin.H{"id": p.ID})
}

// AdminHaokaProductUpsertFromExternal 保存/更新一个外部 product（包含 sku）
// 前端建议直接把 query-products 返回的 product 对象原样传过来。
func AdminHaokaProductUpsertFromExternal(c *gin.Context) {
	_ = database.DB().AutoMigrate(&model.HaokaCategory{}, &model.HaokaProduct{}, &model.HaokaSku{})
	// 为了让接口可用且简单：接收 product 对象（不要求包含 skus 完整结构）
	var p service.HaokaExternalProduct
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	if p.ProductID == 0 {
		response.Error(c, -1, "productID 不能为空")
		return
	}

	cat, err := haokaEnsureCategory(p.Operator)
	if err != nil {
		response.Error(c, -1, err.Error())
		return
	}

	// upsert product
	var exist model.HaokaProduct
	err = database.DB().Where("product_id = ?", p.ProductID).First(&exist).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 500, "保存失败")
		return
	}

	np := mapExternalToProduct(p, cat.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := database.DB().Create(&np).Error; err != nil {
			response.Error(c, 500, "保存失败")
			return
		}
	} else {
		if err := database.DB().Model(&model.HaokaProduct{}).Where("id = ?", exist.ID).Updates(map[string]any{
			"category_id":     np.CategoryID,
			"product_name":    np.ProductName,
			"main_pic":        np.MainPic,
			"area":            np.Area,
			"disable_area":   np.DisableArea,
			"little_picture": np.LittlePicture,
			"net_addr":       np.NetAddr,
			"flag":            np.Flag,
			"number_sel":      np.NumberSel,
			"operator":        np.Operator,
			"back_money_type": np.BackMoneyType,
			"taocan":          np.Taocan,
			"rule":            np.Rule,
			"age1":            np.Age1,
			"age2":            np.Age2,
			"price_time":     np.PriceTime,
			"status":          np.Status,
			"updated_at":      time.Now(),
		}).Error; err != nil {
			response.Error(c, 500, "保存失败")
			return
		}
	}

	// 覆盖 skus
	_ = database.DB().Where("product_id = ?", p.ProductID).Delete(&model.HaokaSku{}).Error
	for _, sku := range p.Skus {
		ns := model.HaokaSku{
			ProductID: p.ProductID,
			SkuID:     sku.SkuID,
			SkuName:   sku.SkuName,
			Desc:      sku.Desc,
		}
		_ = database.DB().Create(&ns).Error
	}

	response.OK(c, gin.H{"ok": true})
}

// Health?（预留）
func HaokaHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

