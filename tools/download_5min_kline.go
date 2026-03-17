package tools

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gitee.com/quant1x/engine/cache"
	"gitee.com/quant1x/engine/datasource/base"
	"gitee.com/quant1x/engine/market"
	"gitee.com/quant1x/exchange"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/gox/progressbar"
)

// Min5KLine 简化版5分钟K线结构体
type Min5KLine struct {
	Timestamps string  // 完整日期时间，格式: 2024-03-04 09:35:00
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	Amount     float64
}

// ConvertKLineToMin5 将原始KLine转换为简化格式
func ConvertKLineToMin5(klines []base.KLine) []Min5KLine {
	result := make([]Min5KLine, 0, len(klines))
	for _, k := range klines {
		// 提取日期时间，去掉毫秒部分
		timestamps := k.Datetime
		if idx := strings.LastIndex(timestamps, "."); idx > 0 {
			timestamps = timestamps[:idx]
		}
		// 去掉最后的3位毫秒标记
		timestamps = strings.TrimSuffix(timestamps, ".000")
		timestamps = strings.TrimSuffix(timestamps, ".004")
		timestamps = strings.TrimSuffix(timestamps, ".005")
		timestamps = strings.TrimSuffix(timestamps, ".006")
		timestamps = strings.TrimSuffix(timestamps, ".00000")
		
		result = append(result, Min5KLine{
			Timestamps: timestamps,
			Open:       k.Open,
			High:       k.High,
			Low:        k.Low,
			Close:      k.Close,
			Volume:     k.Volume,
			Amount:     k.Amount,
		})
	}
	return result
}

// SaveMin5KLineToCsv 保存简化K线数据到CSV
func SaveMin5KLineToCsv(filename string, klines []Min5KLine, appendMode bool) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	var file *os.File
	var err error
	
	if appendMode {
		// 追加模式
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		// 新建模式，先删除旧文件
		os.Remove(filename)
		file, err = os.Create(filename)
	}
	
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// 如果是新建文件或空文件，写入表头
	if !appendMode {
		if err := writer.Write([]string{"timestamps", "open", "high", "low", "close", "volume", "amount"}); err != nil {
			return err
		}
	} else {
		// 检查文件是否为空
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			if err := writer.Write([]string{"timestamps", "open", "high", "low", "close", "volume", "amount"}); err != nil {
				return err
			}
		}
	}
	
	// 写入数据
	for _, k := range klines {
		record := []string{
			k.Timestamps,
			fmt.Sprintf("%.6f", k.Open),
			fmt.Sprintf("%.6f", k.High),
			fmt.Sprintf("%.6f", k.Low),
			fmt.Sprintf("%.6f", k.Close),
			fmt.Sprintf("%.6f", k.Volume),
			fmt.Sprintf("%.6f", k.Amount),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	
	return nil
}

// GetLastTimestampFromCsv 获取CSV文件中最后一条记录的时间戳
func GetLastTimestampFromCsv(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}
	
	// 跳过表头，找最后一条数据
	if len(records) <= 1 {
		return "", nil
	}
	
	lastRecord := records[len(records)-1]
	if len(lastRecord) > 0 {
		return lastRecord[0], nil
	}
	return "", nil
}

// Download5MinKLineForCode 下载单个股票的5分钟K线数据
//
//	code: 股票代码, 如 "sz300773"
//	startDate: 开始日期，格式 "20200101"，空则使用默认值
//	updateMode: 是否为增量更新模式
//	bar: 进度条对象, 可为nil
func Download5MinKLineForCode(code string, startDate string, updateMode bool, bar *progressbar.Bar) error {
	code = exchange.CorrectSecurityCode(code)
	
	// 构建保存路径
	cacheId := cache.CacheId(code)
	length := len(cacheId)
	filename := filepath.Join(cache.GetKLinePath("5min"), cacheId[:length-3], cacheId+".csv")
	
	// 确定实际开始日期
	effectiveStartDate := startDate
	if effectiveStartDate == "" {
		effectiveStartDate = "20200101" // 默认开始日期
	}
	
	// 增量更新模式：检查现有数据
	if updateMode {
		lastTimestamp, err := GetLastTimestampFromCsv(filename)
		if err == nil && lastTimestamp != "" {
			// 解析最后一条记录的时间
			if t, err := time.Parse("2006-01-02 15:04:05", lastTimestamp); err == nil {
				// 从最后日期的下一天开始
				nextDate := t.AddDate(0, 0, 1)
				effectiveStartDate = nextDate.Format("20060102")
				logger.Infof("股票 %s 增量更新，从 %s 开始", code, effectiveStartDate)
			}
		}
	}
	
	// 从服务器获取5分钟K线数据
	klines := base.UpdateAllKLineWithStartDate(code, "5min", effectiveStartDate)
	
	if len(klines) == 0 {
		if bar != nil {
			bar.Add(1)
		}
		return nil
	}
	
	// 转换为简化格式
	min5Klines := ConvertKLineToMin5(klines)
	
	// 保存到CSV
	appendMode := updateMode
	if err := SaveMin5KLineToCsv(filename, min5Klines, appendMode); err != nil {
		logger.Errorf("保存股票 %s 5分钟K线失败: %v", code, err)
		return err
	}
	
	if bar != nil {
		bar.Add(1)
	}
	
	return nil
}

// Download5MinKLineForAll 下载所有股票的5分钟K线数据
func Download5MinKLineForAll(startDate string, updateMode bool) {
	logger.Info("开始下载所有股票的5分钟K线数据...")
	if startDate == "" {
		logger.Info("使用默认开始日期: 2020-01-01")
	} else {
		logger.Infof("使用开始日期: %s", startDate)
	}
	
	// 获取所有股票代码（包括指数）
	allCodes := market.GetCodeList()
	total := len(allCodes)
	logger.Infof("发现 %d 只标的需要下载", total)
	
	// 创建主进度条
	barIndex := 1
	bar := progressbar.NewBar(barIndex, "执行[下载5分钟K线-全部]", total)
	
	// 使用并发下载, 控制并发数量
	maxWorkers := runtime.NumCPU()
	if maxWorkers > 8 {
		maxWorkers = 8
	}
	
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)
	
	for _, code := range allCodes {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		
		go func(stockCode string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			
			_ = Download5MinKLineForCode(stockCode, startDate, updateMode, bar)
		}(code)
	}
	
	wg.Wait()
	bar.Wait()
	
	logger.Info("所有股票5分钟K线下载完成")
}

// Download5MinKLineForSector 下载指定板块的5分钟K线
//
//	sector: 板块前缀, 如 "sh600", "sz300", "sh688" 等
func Download5MinKLineForSector(sector string, startDate string, updateMode bool) {
	// 获取该板块的所有股票
	pattern := filepath.Join(cache.GetDayPath(), sector, "*.csv")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		logger.Errorf("未找到板块 %s 的股票", sector)
		return
	}
	
	total := len(matches)
	logger.Infof("开始下载板块 %s 的5分钟K线, 共 %d 只股票", sector, total)
	if startDate != "" {
		logger.Infof("开始日期: %s", startDate)
	}
	
	// 创建进度条
	barIndex := 1
	bar := progressbar.NewBar(barIndex, "执行[下载5分钟K线-"+sector+"]", total)
	
	// 使用并发下载
	maxWorkers := runtime.NumCPU()
	if maxWorkers > 8 {
		maxWorkers = 8
	}
	
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)
	
	for _, match := range matches {
		filename := filepath.Base(match)
		code := filename[:len(filename)-4] // 去掉.csv后缀
		
		wg.Add(1)
		semaphore <- struct{}{}
		
		go func(stockCode string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			
			_ = Download5MinKLineForCode(stockCode, startDate, updateMode, bar)
		}(code)
	}
	
	wg.Wait()
	bar.Wait()
	
	logger.Infof("板块 %s 5分钟K线下载完成", sector)
}

// Show5MinKLinePath 显示5分钟K线数据存储路径
func Show5MinKLinePath() {
	path := cache.GetKLinePath("5min")
	fmt.Printf("5分钟K线数据存储路径: %s\n", path)
	
	// 检查路径是否存在, 如果不存在则创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("路径不存在, 将会在首次下载时创建\n")
	} else {
		fmt.Printf("路径已存在\n")
	}
}

// Get5MinKLineStats 获取5分钟K线统计信息
func Get5MinKLineStats(code string) (int, string, string) {
	code = exchange.CorrectSecurityCode(code)
	cacheId := cache.CacheId(code)
	length := len(cacheId)
	filename := filepath.Join(cache.GetKLinePath("5min"), cacheId[:length-3], cacheId+".csv")
	
	file, err := os.Open(filename)
	if err != nil {
		return 0, "", ""
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil || len(records) <= 1 {
		return 0, "", ""
	}
	
	// 跳过表头
	count := len(records) - 1
	firstTimestamp := ""
	lastTimestamp := ""
	
	if count > 0 {
		firstTimestamp = records[1][0]
		lastTimestamp = records[len(records)-1][0]
	}
	
	return count, firstTimestamp, lastTimestamp
}

// ValidateDateFormat 验证日期格式 (YYYYMMDD)
func ValidateDateFormat(date string) bool {
	if len(date) != 8 {
		return false
	}
	_, err := time.Parse("20060102", date)
	return err == nil
}

// Get5MinKLineInfo 获取5分钟K线信息
func Get5MinKLineInfo() {
	path := cache.GetKLinePath("5min")
	
	fmt.Println("=== 5分钟K线数据信息 ===")
	fmt.Printf("存储路径: %s\n", path)
	fmt.Printf("默认开始日期: 2020-01-01\n")
	fmt.Printf("日期格式: YYYYMMDD\n")
	fmt.Println()
	fmt.Println("CSV格式:")
	fmt.Println("  timestamps,open,high,low,close,volume,amount")
	fmt.Println("  示例: 2024-03-04 09:35:00,13.41,13.53,13.31,13.33,18332092,-15900905.99")
	fmt.Println()
	fmt.Println("使用说明:")
	fmt.Println("  1. 下载单个股票: ./engine 5min --code=sz300773")
	fmt.Println("  2. 下载指定日期: ./engine 5min --code=sz300773 --start=20230101")
	fmt.Println("  3. 增量更新: ./engine 5min --code=sz300773 --update")
	fmt.Println("  4. 下载某个板块: ./engine 5min --sector=sh600 --start=20240101")
	fmt.Println("  5. 下载所有股票: ./engine 5min --all")
	fmt.Println("  6. 增量更新全部: ./engine 5min --all --update")
	fmt.Println("  7. 显示存储路径: ./engine 5min --path")
	fmt.Println("  8. 查看统计信息: ./engine 5min --code=sz300773 --stats")
}

func init() {
	// 设置最大CPU核数
	runtime.GOMAXPROCS(runtime.NumCPU())
}
