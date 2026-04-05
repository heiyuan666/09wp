package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/model"

	"github.com/xuri/excelize/v2"
)

type ResourceImportRow struct {
	ID uint64

	Title       string
	Link        string
	CategoryID  uint64
	Source      string
	ExternalID  string
	Description string
	ExtractCode string
	Cover       string
	Tags        string

	LinkValid     bool
	LinkCheckMsg  string
	LinkCheckedAt *time.Time

	TransferStatus     string
	TransferMsg        string
	TransferRetryCount int
	TransferLastAt     *time.Time

	ViewCount uint64
	SortOrder int
	Status    int8

	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func ParseResourceSpreadsheet(filename string, data []byte) ([]ResourceImportRow, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "csv":
		return parseResourceCSV(data)
	case "xlsx":
		fallthrough
	default:
		if ext == "xls" || ext == "xlsx" || ext == "csv" {
			// handled above for csv; xls is not supported without extra libs
		}
	}
	// XLSX is the default supported Excel format
	return parseResourceXLSX(data)
}

func parseResourceCSV(data []byte) ([]ResourceImportRow, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	// 尽量自动探测分隔符，避免“逗号 vs 其他分隔”导致列错位
	headerLine := string(bytes.SplitN(data, []byte{'\n'}, 2)[0])
	delims := []rune{',', ';', '\t', '|'}
	bestComma := ','
	bestScore := -1
	for _, d := range delims {
		score := strings.Count(headerLine, string(d))
		if score > bestScore {
			bestScore = score
			bestComma = d
		}
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = bestComma
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	records := make([][]string, 0, 256)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, row)
	}
	if len(records) <= 1 {
		return nil, fmt.Errorf("CSV 为空或表头缺失")
	}
	return parseResourceRowsFromTable(records)
}

func parseResourceXLSX(data []byte) ([]ResourceImportRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("XLSX 无工作表")
	}
	sheet := sheets[0]
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, fmt.Errorf("XLSX 为空或表头缺失")
	}
	return parseResourceRowsFromTable(rows)
}

func parseResourceRowsFromTable(rows [][]string) ([]ResourceImportRow, error) {
	// rows: 第一行是表头
	header := normalizeHeaderRow(rows[0])
	colIndex := make(map[string]int, len(header))
	for i, key := range header {
		if key == "" {
			continue
		}
		// 同一列重复映射时，保留第一个
		if _, ok := colIndex[key]; !ok {
			colIndex[key] = i
		}
	}

	get := func(r []string, key string) string {
		i, ok := colIndex[key]
		if !ok {
			return ""
		}
		if i < 0 || i >= len(r) {
			return ""
		}
		return strings.TrimSpace(r[i])
	}

	out := make([]ResourceImportRow, 0, len(rows)-1)
	for idx := 1; idx < len(rows); idx++ {
		r := rows[idx]
		// 跳过全空行
		if isAllBlank(r) {
			continue
		}

		title := get(r, "title")
		link := get(r, "link")
		categoryID := parseUint64Flexible(get(r, "category_id"))

		rowID := parseUint64Flexible(get(r, "id"))

		sortOrder := parseIntFlexible(get(r, "sort_order"), 0)
		status := parseStatusFlexible(get(r, "status"), 1)

		linkValid := parseBoolFlexible(get(r, "link_valid"), true)
		linkCheckMsg := get(r, "link_check_msg")
		linkCheckedAt, _ := parseTimeFlexible(get(r, "link_checked_at"))

		transferStatus := get(r, "transfer_status")
		transferMsg := get(r, "transfer_msg")
		transferRetryCount := parseIntFlexible(get(r, "transfer_retry_count"), 0)
		transferLastAt, _ := parseTimeFlexible(get(r, "transfer_last_at"))

		viewCount := parseUint64Flexible(get(r, "view_count"))

		createdAt, _ := parseTimeFlexible(get(r, "created_at"))
		updatedAt, _ := parseTimeFlexible(get(r, "updated_at"))

		row := ResourceImportRow{
			ID: rowID,

			Title:       title,
			Link:        link,
			CategoryID:  categoryID,
			Source:      get(r, "source"),
			ExternalID:  get(r, "external_id"),
			Description: get(r, "description"),
			ExtractCode: get(r, "extract_code"),
			Cover:       get(r, "cover"),
			Tags:        get(r, "tags"),

			LinkValid:     linkValid,
			LinkCheckMsg:  linkCheckMsg,
			LinkCheckedAt: linkCheckedAt,

			TransferStatus:     transferStatus,
			TransferMsg:        transferMsg,
			TransferRetryCount: transferRetryCount,
			TransferLastAt:     transferLastAt,

			ViewCount: viewCount,
			SortOrder: sortOrder,
			Status:    status,

			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		out = append(out, row)
	}

	return out, nil
}

func normalizeHeaderRow(headerRow []string) []string {
	out := make([]string, 0, len(headerRow))
	for _, h := range headerRow {
		out = append(out, canonicalHeaderKey(h))
	}
	return out
}

func canonicalHeaderKey(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "_", "")

	// 兼容导出文件中的字段名（下划线会被移除，所以用 normalize 后的 key）
	switch s {
	// 必填/核心
	case "id":
		return "id"
	case "title", "资源标题", "标题":
		return "title"
	case "link", "链接", "url":
		return "link"
	case "categoryid", "分类id", "分类":
		return "category_id"
	case "source":
		return "source"
	case "externalid", "external":
		return "external_id"
	case "description", "描述", "简介":
		return "description"
	case "extractcode", "提取码":
		return "extract_code"
	case "cover", "封面":
		return "cover"
	case "tags", "标签":
		return "tags"

	// 状态类
	case "linkvalid", "linkvalidb", "是否有效":
		return "link_valid"
	case "linkcheckmsg", "linkcheckmsges":
		return "link_check_msg"
	case "linkcheckedat", "linkchecktime":
		return "link_checked_at"
	case "transferstatus":
		return "transfer_status"
	case "transfermsg":
		return "transfer_msg"
	case "transferretrycount":
		return "transfer_retry_count"
	case "transferlastat":
		return "transfer_last_at"

	// 数值/排序
	case "viewcount":
		return "view_count"
	case "sortorder":
		return "sort_order"
	case "status", "显示状态", "是否显示":
		return "status"

	// 时间
	case "createdat":
		return "created_at"
	case "updatedat":
		return "updated_at"
	default:
		// 未识别字段忽略
		return ""
	}
}

func parseUint64Flexible(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Excel 里数字有时会落成 1.0
	if strings.Contains(s, ".") {
		f, err := strconv.ParseFloat(s, 64)
		if err == nil && f > 0 {
			return uint64(f)
		}
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseIntFlexible(s string, defaultValue int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultValue
	}
	if strings.Contains(s, ".") {
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return int(f)
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return v
}

func parseStatusFlexible(s string, defaultValue int8) int8 {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultValue
	}
	switch strings.ToLower(s) {
	case "1", "true", "显示":
		return 1
	case "0", "false", "隐藏":
		return 0
	default:
		v, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return defaultValue
		}
		return int8(v)
	}
}

func parseBoolFlexible(s string, defaultValue bool) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultValue
	}
	switch strings.ToLower(s) {
	case "1", "true", "yes", "是", "有效":
		return true
	case "0", "false", "no", "否", "无效", "失效":
		return false
	default:
		// 兼容 0/1 之外的数字字符串
		v, err := strconv.Atoi(s)
		if err == nil {
			return v != 0
		}
		return defaultValue
	}
}

func parseTimeFlexible(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}

	// 尝试把 Excel serial date 当作“天数”
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		// Excel: 0 对应 1899-12-30（有 1900 闰年历史 bug，这里按常见做法）
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.Local)
		// days may be fractional (time part)
		d := time.Duration(f*24*float64(time.Hour) + 0.5)
		t := base.Add(d)
		return &t, nil
	}

	return nil, fmt.Errorf("无法解析时间: %s", s)
}

func isAllBlank(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

const timeLayout = "2006-01-02 15:04:05"

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeLayout)
}

func formatTime(t time.Time) string {
	return t.Format(timeLayout)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ExportResourcesToXLSX 生成 XLSX 文件（写入到 outPath）
func ExportResourcesToXLSX(resources []model.Resource, outPath string) error {
	f := excelize.NewFile()
	sheet := "resources"
	f.SetSheetName("Sheet1", sheet)

	// 固定列：用于“导出 -> 原样导入资源管理”
	// 注意：这些字段名会被前端/导入解析做大小写/下划线兼容处理。
	header := []string{
		"id",
		"title",
		"link",
		"category_id",
		"source",
		"external_id",
		"description",
		"extract_code",
		"cover",
		"tags",
		"link_valid",
		"link_check_msg",
		"link_checked_at",
		"transfer_status",
		"transfer_msg",
		"transfer_retry_count",
		"transfer_last_at",
		"view_count",
		"sort_order",
		"status",
		"created_at",
		"updated_at",
	}
	for col, h := range header {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for i, r := range resources {
		rowNum := i + 2
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), r.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), r.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), r.Link)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), r.CategoryID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), r.Source)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), r.ExternalID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), r.Description)
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), r.ExtractCode)
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", rowNum), r.Cover)
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", rowNum), r.Tags)
		_ = f.SetCellValue(sheet, fmt.Sprintf("K%d", rowNum), boolToInt(r.LinkValid))
		_ = f.SetCellValue(sheet, fmt.Sprintf("L%d", rowNum), r.LinkCheckMsg)
		_ = f.SetCellValue(sheet, fmt.Sprintf("M%d", rowNum), formatTimePtr(r.LinkCheckedAt))
		_ = f.SetCellValue(sheet, fmt.Sprintf("N%d", rowNum), r.TransferStatus)
		_ = f.SetCellValue(sheet, fmt.Sprintf("O%d", rowNum), r.TransferMsg)
		_ = f.SetCellValue(sheet, fmt.Sprintf("P%d", rowNum), r.TransferRetryCount)
		_ = f.SetCellValue(sheet, fmt.Sprintf("Q%d", rowNum), formatTimePtr(r.TransferLastAt))
		_ = f.SetCellValue(sheet, fmt.Sprintf("R%d", rowNum), r.ViewCount)
		_ = f.SetCellValue(sheet, fmt.Sprintf("S%d", rowNum), r.SortOrder)
		_ = f.SetCellValue(sheet, fmt.Sprintf("T%d", rowNum), r.Status)
		_ = f.SetCellValue(sheet, fmt.Sprintf("U%d", rowNum), formatTime(r.CreatedAt))
		_ = f.SetCellValue(sheet, fmt.Sprintf("V%d", rowNum), formatTime(r.UpdatedAt))
	}

	return f.SaveAs(outPath)
}

// ExportResourcesToCSV 生成 CSV 文件（写入到 outPath）
func ExportResourcesToCSV(resources []model.Resource, outPath string) error {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// 固定列：与 XLSX 同步，便于“导出 -> 导入”
	header := []string{
		"id",
		"title",
		"link",
		"category_id",
		"source",
		"external_id",
		"description",
		"extract_code",
		"cover",
		"tags",
		"link_valid",
		"link_check_msg",
		"link_checked_at",
		"transfer_status",
		"transfer_msg",
		"transfer_retry_count",
		"transfer_last_at",
		"view_count",
		"sort_order",
		"status",
		"created_at",
		"updated_at",
	}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, r := range resources {
		line := []string{
			strconv.FormatUint(r.ID, 10),
			r.Title,
			r.Link,
			strconv.FormatUint(r.CategoryID, 10),
			r.Source,
			r.ExternalID,
			r.Description,
			r.ExtractCode,
			r.Cover,
			r.Tags,
			strconv.Itoa(boolToInt(r.LinkValid)),
			r.LinkCheckMsg,
			formatTimePtr(r.LinkCheckedAt),
			r.TransferStatus,
			r.TransferMsg,
			strconv.Itoa(r.TransferRetryCount),
			formatTimePtr(r.TransferLastAt),
			strconv.FormatUint(r.ViewCount, 10),
			strconv.Itoa(r.SortOrder),
			strconv.Itoa(int(r.Status)),
			formatTime(r.CreatedAt),
			formatTime(r.UpdatedAt),
		}
		if err := w.Write(line); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return os.WriteFile(outPath, buf.Bytes(), 0644)
}
