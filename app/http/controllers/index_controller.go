package controllers

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type Index struct{}

func (*Index) Index(c *gin.Context) {
	c.String(http.StatusOK, "测试页面")
}

func (*Index) TxtToExcel(c *gin.Context) {
	reader, err := os.Open("2.txt")
	if err != nil {
		c.String(http.StatusBadRequest, "打开1.txt失败: %v", err)
		return
	}
	defer reader.Close()

	cities, err := readUniqueCities(reader)
	if err != nil {
		c.String(http.StatusInternalServerError, "读取1.txt失败: %v", err)
		return
	}
	if len(cities) == 0 {
		c.String(http.StatusBadRequest, "1.txt没有有效城市数据")
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)
	streamWriter, err := f.NewStreamWriter(sheetName)
	if err != nil {
		c.String(http.StatusInternalServerError, "创建Excel失败: %v", err)
		return
	}

	if err := streamWriter.SetRow("A1", []interface{}{"开通城市"}); err != nil {
		c.String(http.StatusInternalServerError, "写入Excel表头失败: %v", err)
		return
	}
	for i, city := range cities {
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			c.String(http.StatusInternalServerError, "生成Excel单元格失败: %v", err)
			return
		}
		if err := streamWriter.SetRow(cell, []interface{}{city}); err != nil {
			c.String(http.StatusInternalServerError, "写入Excel失败: %v", err)
			return
		}
	}
	if err := streamWriter.Flush(); err != nil {
		c.String(http.StatusInternalServerError, "保存Excel数据失败: %v", err)
		return
	}

	filename := fmt.Sprintf("开通城市-%s.xlsx", time.Now().Format("20060102150405"))
	disposition := fmt.Sprintf("attachment; filename=open_cities.xlsx; filename*=UTF-8''%s", url.QueryEscape(filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", disposition)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if err := f.Write(c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "输出Excel失败: %v", err)
	}
}

func readUniqueCities(reader io.Reader) ([]string, error) {
	bufReader := bufio.NewReaderSize(reader, 1024*1024)
	seen := make(map[string]struct{})
	cities := make([]string, 0)
	var token strings.Builder

	addCity := func() {
		text := cleanCity(token.String())
		token.Reset()

		for _, city := range splitCityNames(text) {
			if city == "" || city == "开通城市" {
				continue
			}
			if _, ok := seen[city]; ok {
				continue
			}

			seen[city] = struct{}{}
			cities = append(cities, city)
		}
	}

	for {
		r, _, err := bufReader.ReadRune()
		if err == nil {
			if isCityDelimiter(r) {
				addCity()
			} else {
				token.WriteRune(r)
			}
			continue
		}

		if err == io.EOF {
			addCity()
			break
		}
		return nil, err
	}

	return cities, nil
}

func cleanCity(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '\ufeff' {
			return -1
		}
		return r
	}, value)
}

func isCityDelimiter(r rune) bool {
	return r == ',' || r == '，' || r == '\n' || r == '\r'
}

func splitCityNames(value string) []string {
	if value == "" {
		return nil
	}

	suffixes := []string{"特别行政区", "自治州", "地区", "林区", "盟", "市"}
	cities := make([]string, 0)
	start := 0

	for pos := range value {
		if pos < start {
			continue
		}
		for _, suffix := range suffixes {
			if strings.HasPrefix(value[pos:], suffix) {
				end := pos + len(suffix)
				cities = append(cities, value[start:end])
				start = end
				break
			}
		}
	}

	if start < len(value) {
		cities = append(cities, value[start:])
	}

	return cities
}
