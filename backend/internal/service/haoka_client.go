package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// keep-alive reference to avoid gopls unusedfunc warning (used by optional flows / future extensions)
var _ = sortByProductIDStable

const haokaGetProductsURL = "https://haokaopenapi.lot-ml.com/api/order/GetProductsV2"

type HaokaExternalSku struct {
	SkuID   uint64 `json:"SkuID"`
	SkuName string `json:"SkuName"`
	Desc    string `json:"Desc"`
}

type HaokaExternalProduct struct {
	ProductID     uint64             `json:"productID"`
	ProductName   string             `json:"productName"`
	MainPic       string             `json:"mainPic"`
	Area          string             `json:"area"`
	DisableArea   string             `json:"disableArea"`
	LittlePicture string             `json:"littlepicture"`
	NetAddr       string             `json:"netAddr"`
	Flag          bool               `json:"flag"`
	NumberSel     int                `json:"numberSel"`
	Operator      string             `json:"operator"`
	BackMoneyType string             `json:"BackMoneyType"`
	Taocan        string             `json:"Taocan"`
	Rule          string             `json:"Rule"`
	Age1          int                `json:"Age1"`
	Age2          int                `json:"Age2"`
	PriceTime     string             `json:"PriceTime"`
	Skus          []HaokaExternalSku `json:"Skus"`
}

type haokaGetProductsV2Resp struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    []HaokaExternalProduct `json:"data"`
	Errs    any                    `json:"errs"`
}

func md5LowerHex(input string) string {
	sum := md5.Sum([]byte(input))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}

// haokaSign 按用户给的规则构造：
// Md5("ProductID=" + ProductID + "&Timestamp=" + Timestamp + "&user_id=" + user_id + secret)
func haokaSign(userID, productID, timestamp, secret string) string {
	base := fmt.Sprintf("ProductID=%s&Timestamp=%s&user_id=%s", productID, timestamp, userID)
	return md5LowerHex(base + secret)
}

func haokaTimestampNow10() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// HaokaQueryProducts 调外部接口，返回外部 products 列表
// external API：Post + form-data
func HaokaQueryProducts(userID, secret, productID string) ([]HaokaExternalProduct, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("user_id 或 secret 不能为空")
	}
	timestamp := haokaTimestampNow10()
	userSign := haokaSign(userID, productID, timestamp, secret)

	// form-data 构造（无文件，仅字段）
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("user_id", userID)
	_ = writer.WriteField("Timestamp", timestamp)
	_ = writer.WriteField("ProductID", productID)
	_ = writer.WriteField("user_sign", userSign)
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, haokaGetProductsURL, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed haokaGetProductsV2Resp
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, err
	}

	if parsed.Code != 0 {
		if strings.TrimSpace(parsed.Message) == "" {
			return nil, fmt.Errorf("haoka query failed: code=%d", parsed.Code)
		}
		return nil, fmt.Errorf("haoka query failed: %s (code=%d)", parsed.Message, parsed.Code)
	}
	return parsed.Data, nil
}

// OperatorCategorySlug 将运营商映射为分类 slug
// 注意：slug 用于 URL/唯一性，仅作内部使用。
func OperatorCategorySlug(op string) string {
	switch strings.TrimSpace(op) {
	case "电信":
		return "dianxin"
	case "移动":
		return "yidong"
	case "联通":
		return "liantong"
	default:
		return "unknown"
	}
}

// sortByProductIDStable 用于排序（保证前端显示稳定）
func sortByProductIDStable(ps []HaokaExternalProduct) {
	// 外部数据顺序不保证每次一致，这里做稳定排序
	sort.SliceStable(ps, func(i, j int) bool {
		return ps[i].ProductID < ps[j].ProductID
	})
}
