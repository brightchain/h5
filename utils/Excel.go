package utils

import (
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func getHeaders(v *reflect.Value) []string {

	columns := make([]string, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		//取tag
		column := field.Tag.Get("tag")
		isExp := field.Tag.Get("exp")
		if isExp == "1" {
			continue
		}
		columns = append(columns, column)
	}
	return columns
}

func Down[T any](data []T, filename string, c *gin.Context) {
	DownExcel(data, filename, "Sheet1", c)
}

func DownExcel[T any](data []T, filename string, sheetName string, c *gin.Context) {
	f := excelize.NewFile()
	f.SetSheetName(sheetName, sheetName)
	v0 := reflect.ValueOf(data[0])
	if v0.Kind() == reflect.Ptr {
		v0 = v0.Elem()
	}
	if v0.Kind() != reflect.Struct {
		return
	}
	header := getHeaders(&v0)
	_ = f.SetSheetRow(sheetName, "A1", &header)

	rowNum := 1 //数据开始行数
	for _, value := range data {
		row := make([]interface{}, 0)
		vI := reflect.ValueOf(value)
		if vI.Kind() == reflect.Ptr {
			vI = vI.Elem()
		}
		for i := 0; i < vI.NumField(); i++ {
			isExp := vI.Type().Field(i).Tag.Get("exp")
			if isExp == "1" {
				continue
			}
			row = append(row, fmt.Sprintf("%v", vI.Field(i)))
		}
		rowNum++
		f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &row)
	}

	disposition := fmt.Sprintf("attachment; filename=%s-%s.xlsx", url.QueryEscape(filename), time.Now().Format("2006-01-02"))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Disposition", disposition)
	c.Writer.Header().Set("Content-Transfer-Encoding", "binary")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	_ = f.Write(c.Writer)
}

func SaveFile[T any](data []T, filename string) {
	sheetName := "Sheet1"
	f := excelize.NewFile()
	f.SetSheetName(sheetName, sheetName)
	v0 := reflect.ValueOf(data[0])
	if v0.Kind() == reflect.Ptr {
		v0 = v0.Elem()
	}
	if v0.Kind() != reflect.Struct {
		return
	}
	header := getHeaders(&v0)
	_ = f.SetSheetRow(sheetName, "A1", &header)

	rowNum := 1 //数据开始行数
	for _, value := range data {
		row := make([]interface{}, 0)
		vI := reflect.ValueOf(value)
		if vI.Kind() == reflect.Ptr {
			vI = vI.Elem()
		}
		for i := 0; i < vI.NumField(); i++ {
			isExp := vI.Type().Field(i).Tag.Get("exp")
			if isExp == "1" {
				continue
			}
			row = append(row, fmt.Sprintf("%v", vI.Field(i)))
		}
		rowNum++
		f.SetSheetRow("Sheet1", fmt.Sprintf("A%d", rowNum), &row)
	}
	if err := f.SaveAs(filename); err != nil {
		slog.Error("excel error", err)
	}
}
