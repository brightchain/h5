package controllers

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Sqlbak struct{}

func (*Sqlbak) Index(c *gin.Context) {
	inputFile := `D:\car_20250331_015001.sql`
	outputFile := `D:\car_coupon_pkg.sql`

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
		if strings.Contains(line, "CREATE TABLE `car_coupon_pkg`") {
			inCreateTable = true
		}

		// 检查 INSERT INTO
		if strings.Contains(line, "INSERT INTO `car_coupon_pkg`") {
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
