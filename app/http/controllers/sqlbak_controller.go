package controllers

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type Sqlbak struct{}

func (*Sqlbak) Index(c *gin.Context) {
	inputFile := `E:\code\sharelive\src\car_20250616_015001.sql`
	outputFile := `D:\car_shop_dadi_order.sql`

	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("打开输入文件失败: %v\n", err)
		return
	}
	defer file.Close()

	// 创建输出文件
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer out.Close()

	// 创建 Scanner
	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// 设置缓冲区大小优化大文件读取（10GB）
	const maxBufferSize = 1024 * 1024 // 1MB
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxBufferSize)

	// 标记提取状态
	inCreateTable := false
	inInsertData := false

	// 逐行读取文件
	for scanner.Scan() {
		line := scanner.Text()

		// 检查 CREATE TABLE
		if strings.Contains(line, "CREATE TABLE `car_shop_dadi_order`") {
			inCreateTable = true
		}

		// 检查 INSERT INTO
		if strings.Contains(line, "INSERT INTO `car_shop_dadi_order`") {
			inInsertData = true
		}

		// 写入匹配的行
		if inCreateTable || inInsertData {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				fmt.Printf("写入失败: %v\n", err)
				return
			}
		}

		// CREATE TABLE 结束
		if inCreateTable && strings.HasSuffix(strings.TrimSpace(line), ");") {
			inCreateTable = false
		}

		// INSERT INTO 结束
		if inInsertData && (strings.HasPrefix(line, "--") || strings.TrimSpace(line) == "" || (strings.Contains(line, "INSERT INTO") && !strings.Contains(line, "`car_coupon_pkg`"))) {
			inInsertData = false
		}
	}

	// 检查读取过程中的错误
	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	fmt.Println("提取完成，结果已保存到", outputFile)
}

type OrdersData struct {
	OrderNo     string
	TrackingNo  string
	ExpressComp string
}

// 优化后的处理函数
func (*Sqlbak) Excel(c *gin.Context) {
	startTime := time.Now()

	sqlFile := `D:\car_shop_dadi_order.sql`
	outputFile := `D:\car_shop_dadi_order.xlsx`
	excelFile := `D:\1.xlsx`

	// 检查文件是否存在
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		c.String(200, "SQL文件不存在")
		return
	}

	// 使用流式处理解析SQL文件
	orders, err := parseOrdersFromLargeFile(sqlFile)
	if err != nil {
		c.String(200, fmt.Sprintf("解析SQL文件失败: %v", err))
		return
	}

	log.Printf("解析完成，共找到 %d 条订单数据", len(orders))

	// 处理Excel文件
	err = processExcelFile(excelFile, outputFile, orders)
	if err != nil {
		c.String(200, fmt.Sprintf("处理Excel文件失败: %v", err))
		return
	}

	duration := time.Since(startTime)
	c.String(200, fmt.Sprintf("处理完成！耗时: %v，结果已保存到: %s", duration, outputFile))
}

// 流式处理大文件
func parseOrdersFromLargeFile(filename string) (map[string]OrdersData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	orders := make(map[string]OrdersData)
	scanner := bufio.NewScanner(file)

	// 设置更大的缓冲区以处理长行
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	// 编译正则表达式
	insertRe := regexp.MustCompile(`INSERT\s+INTO\s+.*?\s+VALUES\s*`)
	valueRe := regexp.MustCompile(`\(([^)]+)\)`)

	lineCount := 0
	matchCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// 进度显示
		if lineCount%10000 == 0 {
			log.Printf("已处理 %d 行", lineCount)
		}

		// 检查是否是INSERT语句
		if !insertRe.MatchString(line) {
			continue
		}

		// 提取所有VALUES中的内容
		matches := valueRe.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				values := match[1]
				order, err := parseOrderValues(values)
				if err != nil {
					log.Printf("解析订单数据失败: %v", err)
					continue
				}
				if order.OrderNo != "" {
					orders[order.OrderNo] = order
					matchCount++
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件时出错: %v", err)
	}

	log.Printf("文件解析完成，共处理 %d 行，匹配 %d 条订单", lineCount, matchCount)
	return orders, nil
}

// 分批处理大文件的替代方案
func parseOrdersFromLargeFileChunked(filename string) (map[string]OrdersData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	orders := make(map[string]OrdersData)
	var mu sync.Mutex

	const chunkSize = 64 * 1024 * 1024 // 64MB chunks
	buf := make([]byte, chunkSize)
	leftover := ""

	insertRe := regexp.MustCompile(`INSERT\s+INTO\s+.*?\s+VALUES\s*`)
	valueRe := regexp.MustCompile(`\(([^)]+)\)`)

	chunkCount := 0
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}

		chunkCount++
		log.Printf("处理第 %d 个数据块", chunkCount)

		chunk := leftover + string(buf[:n])

		// 找到最后一个完整行
		lastNewline := strings.LastIndex(chunk, "\n")
		if lastNewline != -1 {
			processChunk := chunk[:lastNewline]
			leftover = chunk[lastNewline+1:]

			// 并发处理数据块
			go func(data string) {
				localOrders := make(map[string]OrdersData)
				lines := strings.Split(data, "\n")

				for _, line := range lines {
					if !insertRe.MatchString(line) {
						continue
					}

					matches := valueRe.FindAllStringSubmatch(line, -1)
					for _, match := range matches {
						if len(match) > 1 {
							values := match[1]
							order, parseErr := parseOrderValues(values)
							if parseErr != nil {
								continue
							}
							if order.OrderNo != "" {
								localOrders[order.OrderNo] = order
							}
						}
					}
				}

				// 合并结果
				mu.Lock()
				for k, v := range localOrders {
					orders[k] = v
				}
				mu.Unlock()
			}(processChunk)
		} else {
			leftover = chunk
		}

		if err != nil {
			break
		}
	}

	// 处理剩余内容
	if leftover != "" {
		lines := strings.Split(leftover, "\n")
		for _, line := range lines {
			if !insertRe.MatchString(line) {
				continue
			}

			matches := valueRe.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					values := match[1]
					order, parseErr := parseOrderValues(values)
					if parseErr != nil {
						continue
					}
					if order.OrderNo != "" {
						orders[order.OrderNo] = order
					}
				}
			}
		}
	}

	return orders, nil
}

// 处理Excel文件
func processExcelFile(inputFile, outputFile string, orders map[string]OrdersData) error {
	f, err := excelize.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表失败: %v", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("Excel文件为空")
	}

	// 检查和添加表头
	headers := rows[0]
	trackingColIndex, expressColIndex := findOrAddColumns(f, sheetName, headers)

	// 处理数据行
	matchCount := 0
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) == 0 {
			continue
		}

		searchKey := strings.TrimSpace(rows[i][0])
		if order, found := orders[searchKey]; found {
			if order.TrackingNo != "" {
				cellName := fmt.Sprintf("%s%d", getColumnName(trackingColIndex), i+1)
				f.SetCellValue(sheetName, cellName, order.TrackingNo)
			}

			if order.ExpressComp != "" {
				cellName := fmt.Sprintf("%s%d", getColumnName(expressColIndex), i+1)
				f.SetCellValue(sheetName, cellName, order.ExpressComp)
			}

			matchCount++
		}
	}

	log.Printf("匹配了 %d 条记录", matchCount)

	if err := f.SaveAs(outputFile); err != nil {
		return fmt.Errorf("保存Excel文件失败: %v", err)
	}

	return nil
}

// 查找或添加列
func findOrAddColumns(f *excelize.File, sheetName string, headers []string) (int, int) {
	trackingColIndex := -1
	expressColIndex := -1

	for i, header := range headers {
		if strings.Contains(header, "快递单号") || strings.Contains(header, "tracking") {
			trackingColIndex = i
		}
		if strings.Contains(header, "快递公司") || strings.Contains(header, "express") {
			expressColIndex = i
		}
	}

	if trackingColIndex == -1 {
		trackingColIndex = len(headers)
		f.SetCellValue(sheetName, fmt.Sprintf("%s1", getColumnName(trackingColIndex)), "快递单号")
	}
	if expressColIndex == -1 {
		expressColIndex = len(headers)
		if trackingColIndex == len(headers) {
			expressColIndex = len(headers) + 1
		}
		f.SetCellValue(sheetName, fmt.Sprintf("%s1", getColumnName(expressColIndex)), "快递公司")
	}

	return trackingColIndex, expressColIndex
}

// 优化后的字段解析函数
func parseOrderValues(values string) (OrdersData, error) {
	var order OrdersData

	// 使用更高效的字段分割
	fields := splitFieldsOptimized(values)

	if len(fields) < 35 {
		return order, fmt.Errorf("字段数量不足: %d", len(fields))
	}

	order.OrderNo = cleanValue(fields[1])
	order.TrackingNo = cleanValue(fields[29])
	order.ExpressComp = cleanValue(fields[30])

	return order, nil
}

// 优化的字段分割函数
func splitFieldsOptimized(values string) []string {
	fields := make([]string, 0, 40) // 预分配容量
	var current strings.Builder
	current.Grow(256) // 预分配字符串构建器容量

	inQuotes := false
	escaped := false

	for _, char := range values {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}

		switch char {
		case '\\':
			escaped = true
			current.WriteRune(char)
		case '\'':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				fields = append(fields, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}

func cleanValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		value = value[1 : len(value)-1]
	}
	return value
}

func getColumnName(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}
