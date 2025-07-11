package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
	"example.com/m/v2/internal/config"
	"example.com/m/v2/pkg/models"
)

// ExcelStorage Excel存储实现
type ExcelStorage struct {
	config    config.StorageConfig
	outputDir string
	file      *excelize.File
	filePath  string
	mu        sync.Mutex
	rowIndex  int
	sheetName string
}

// NewExcelStorage 创建Excel存储实例
func NewExcelStorage(cfg config.StorageConfig) (*ExcelStorage, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建Excel文件
	f := excelize.NewFile()
	
	// 设置工作表名称
	sheetName := "CrawlData"
	f.SetSheetName("Sheet1", sheetName)

	// 创建输出文件路径
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("crawl_data_%s.xlsx", timestamp)
	filePath := filepath.Join(cfg.OutputDir, filename)

	storage := &ExcelStorage{
		config:    cfg,
		outputDir: cfg.OutputDir,
		file:      f,
		filePath:  filePath,
		rowIndex:  1,
		sheetName: sheetName,
	}

	// 设置表头
	if err := storage.setupHeaders(); err != nil {
		return nil, fmt.Errorf("设置表头失败: %w", err)
	}

	return storage, nil
}

// setupHeaders 设置表头
func (es *ExcelStorage) setupHeaders() error {
	headers := []string{
		"URL", "标题", "内容", "描述", "作者", "来源", 
		"发布日期", "抓取时间", "关键词", "标签", "链接", "图片",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", getExcelColumn(i), es.rowIndex)
		if err := es.file.SetCellValue(es.sheetName, cell, header); err != nil {
			return err
		}
	}

	// 设置表头样式
	if err := es.setHeaderStyle(); err != nil {
		return err
	}

	es.rowIndex++
	return nil
}

// setHeaderStyle 设置表头样式
func (es *ExcelStorage) setHeaderStyle() error {
	// 创建样式
	style, err := es.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6F3FF"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 应用样式到表头行
	for i := 0; i < 12; i++ {
		cell := fmt.Sprintf("%s%d", getExcelColumn(i), 1)
		if err := es.file.SetCellStyle(es.sheetName, cell, cell, style); err != nil {
			return err
		}
	}

	return nil
}

// Save 保存单个数据项
func (es *ExcelStorage) Save(item *models.Item) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// 数据行
	data := []interface{}{
		item.URL,
		item.Title,
		item.Content,
		item.Description,
		item.Author,
		item.Source,
		item.PublishDate.Format("2006-01-02 15:04:05"),
		item.Timestamp.Format("2006-01-02 15:04:05"),
		strings.Join(item.Keywords, "; "),
		strings.Join(item.Tags, "; "),
		strings.Join(item.Links, "; "),
		strings.Join(item.Images, "; "),
	}

	// 写入数据
	for i, value := range data {
		cell := fmt.Sprintf("%s%d", getExcelColumn(i), es.rowIndex)
		if err := es.file.SetCellValue(es.sheetName, cell, value); err != nil {
			return fmt.Errorf("写入单元格失败: %w", err)
		}
	}

	// 设置行高
	if err := es.file.SetRowHeight(es.sheetName, es.rowIndex, 25); err != nil {
		return fmt.Errorf("设置行高失败: %w", err)
	}

	es.rowIndex++
	return nil
}

// SaveBatch 批量保存数据项
func (es *ExcelStorage) SaveBatch(items []*models.Item) error {
	for _, item := range items {
		if err := es.Save(item); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭Excel存储
func (es *ExcelStorage) Close() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.file == nil {
		return nil
	}

	// 设置列宽
	if err := es.adjustColumnWidths(); err != nil {
		return fmt.Errorf("调整列宽失败: %w", err)
	}

	// 添加筛选器
	if err := es.addAutoFilter(); err != nil {
		return fmt.Errorf("添加筛选器失败: %w", err)
	}

	// 保存文件
	if err := es.file.SaveAs(es.filePath); err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}

	// 关闭文件
	if err := es.file.Close(); err != nil {
		return fmt.Errorf("关闭Excel文件失败: %w", err)
	}

	es.file = nil
	return nil
}

// adjustColumnWidths 调整列宽
func (es *ExcelStorage) adjustColumnWidths() error {
	// 定义各列的宽度
	columnWidths := map[string]float64{
		"A": 50,  // URL
		"B": 30,  // 标题
		"C": 50,  // 内容
		"D": 30,  // 描述
		"E": 15,  // 作者
		"F": 15,  // 来源
		"G": 20,  // 发布日期
		"H": 20,  // 抓取时间
		"I": 20,  // 关键词
		"J": 20,  // 标签
		"K": 30,  // 链接
		"L": 30,  // 图片
	}

	for col, width := range columnWidths {
		if err := es.file.SetColWidth(es.sheetName, col, col, width); err != nil {
			return err
		}
	}

	return nil
}

// addAutoFilter 添加自动筛选器
func (es *ExcelStorage) addAutoFilter() error {
	if es.rowIndex <= 1 {
		return nil // 没有数据行，不需要筛选器
	}

	// 设置筛选器范围
	filterRange := fmt.Sprintf("A1:L%d", es.rowIndex-1)
	return es.file.AutoFilter(es.sheetName, filterRange, "")
}

// GetOutputPath 获取输出文件路径
func (es *ExcelStorage) GetOutputPath() string {
	return es.filePath
}

// AddChart 添加图表
func (es *ExcelStorage) AddChart() error {
	if es.rowIndex <= 2 {
		return nil // 数据不足，无法创建图表
	}

	// 创建饼图显示来源分布
	chart, err := es.file.AddChart(es.sheetName, fmt.Sprintf("N2"), &excelize.Chart{
		Type: excelize.Pie,
		Series: []excelize.ChartSeries{
			{
				Name:       "来源分布",
				Categories: fmt.Sprintf("%s!$F$2:$F$%d", es.sheetName, es.rowIndex-1),
				Values:     fmt.Sprintf("%s!$F$2:$F$%d", es.sheetName, es.rowIndex-1),
			},
		},
		Title: []excelize.RichTextRun{
			{
				Text: "数据来源分布",
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowPercent: true,
		},
		ShowBlanksAs: "gap",
		XAxis: excelize.ChartAxis{
			Font: excelize.Font{Size: 12},
		},
		YAxis: excelize.ChartAxis{
			Font: excelize.Font{Size: 12},
		},
		Dimension: excelize.ChartDimension{
			Width:  480,
			Height: 290,
		},
	})

	return chart
}

// AddSummary 添加摘要信息
func (es *ExcelStorage) AddSummary() error {
	// 在单独的工作表中添加摘要
	summarySheet := "Summary"
	es.file.NewSheet(summarySheet)

	// 添加摘要数据
	summaryData := [][]interface{}{
		{"统计项目", "数值"},
		{"总记录数", es.rowIndex - 1},
		{"抓取时间", time.Now().Format("2006-01-02 15:04:05")},
		{"存储位置", es.filePath},
	}

	for i, row := range summaryData {
		for j, value := range row {
			cell := fmt.Sprintf("%s%d", getExcelColumn(j), i+1)
			if err := es.file.SetCellValue(summarySheet, cell, value); err != nil {
				return err
			}
		}
	}

	// 设置摘要表样式
	if err := es.setSummaryStyle(summarySheet); err != nil {
		return err
	}

	return nil
}

// setSummaryStyle 设置摘要表样式
func (es *ExcelStorage) setSummaryStyle(sheetName string) error {
	// 创建表头样式
	headerStyle, err := es.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	// 应用样式
	if err := es.file.SetCellStyle(sheetName, "A1", "B1", headerStyle); err != nil {
		return err
	}

	// 设置列宽
	if err := es.file.SetColWidth(sheetName, "A", "A", 20); err != nil {
		return err
	}
	if err := es.file.SetColWidth(sheetName, "B", "B", 30); err != nil {
		return err
	}

	return nil
}

// getExcelColumn 获取Excel列名
func getExcelColumn(index int) string {
	if index < 26 {
		return string(rune('A' + index))
	}
	
	// 处理超过26列的情况
	first := index / 26
	second := index % 26
	return string(rune('A'+first-1)) + string(rune('A'+second))
} 