package controllers

import (
	"fmt"
	"h5/pkg/model"
	"h5/utils"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// OrderType 定义订单类型常量
const (
	OrderTypeYZ = "YZ"
	OrderTypeDD = "DD"
	OrderTypeDA = "DA"
	OrderTypeDS = "DS"
	OrderTypeVC = "VC"
	OrderTypeGP = "GP"
	OrderTypeZY = "ZY"
	OrderTypeCA = "CA"
)

// OrderData 存储订单基础信息
type OrderData struct {
	RowIndex int    // Excel行索引
	OrderNo  string // 订单号
	PayNo    string // 支付号（VC类型可能用到）
}

// QueryResult 查询结果
type QueryResult struct {
	OrderNo     string  `json:"order_no"`
	ProductName string  `json:"product_name,omitempty"`
	ProName     string  `json:"pro_name,omitempty"`
	Amount      int     `json:"amount"`
	OrderAmount float64 `json:"order_amount"`
	PayAmount   float64 `json:"pay_amount"`
	Num         int     `json:"num"`
	PayNo       string  `json:"pay_no"`
}

type PayOrder struct {
}

func (p *PayOrder) GetOrderProduct(c *gin.Context) {
	// 1. 打开并读取Excel
	f, err := excelize.OpenFile("order.xlsx")
	if err != nil {
		p.handleError(c, "打开Excel文件失败", err)
		return
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		p.handleError(c, "读取工作表失败", err)
		return
	}

	// 2. 分类收集订单数据
	orderGroups, err := p.collectOrderGroups(rows)
	if err != nil {
		p.handleError(c, "收集订单数据失败", err)
		return
	}

	// 3. 批量查询各类订单
	results, err := p.batchQueryOrders(orderGroups)
	if err != nil {
		p.handleError(c, "查询订单数据失败", err)
		return
	}

	// 4. 更新Excel
	if err := p.updateExcelWithResults(f, results); err != nil {
		p.handleError(c, "更新Excel失败", err)
		return
	}

	// 5. 保存文件
	if err := f.SaveAs("output.xlsx"); err != nil {
		p.handleError(c, "保存文件失败", err)
		return
	}

	c.String(http.StatusOK, "处理完成")
}

func (p *PayOrder) collectOrderGroups(rows [][]string) (map[string][]OrderData, error) {
	orderGroups := make(map[string][]OrderData)

	for i, row := range rows {
		if i == 0 || len(row) <= 10 { // 跳过标题行和无效行
			continue
		}

		orderNo := strings.TrimSpace(strings.ReplaceAll(row[10], "`", ""))
		if orderNo == "" {
			continue
		}

		// 处理特殊情况：订单号以"10"开头
		payNo := ""

		if (orderNo[:2] == "10" || orderNo[:2] == "wx") && len(row) > 11 {
			orderNo = strings.TrimSpace(strings.ReplaceAll(row[11], "`", ""))
		} else if len(row) > 11 {
			payNo = strings.TrimSpace(strings.ReplaceAll(row[11], "`", ""))
		}

		if orderNo == "" {
			continue
		}

		orderType := orderNo[:2]
		orderData := OrderData{
			RowIndex: i + 1,
			OrderNo:  orderNo,
			PayNo:    payNo,
		}

		orderGroups[orderType] = append(orderGroups[orderType], orderData)
	}

	return orderGroups, nil
}

func (p *PayOrder) batchQueryOrders(orderGroups map[string][]OrderData) (map[string]QueryResult, error) {
	results := make(map[string]QueryResult)
	db := model.RDB[model.MASTER].Db

	// 批量查询各类型订单
	for orderType, orders := range orderGroups {
		var orderNos []string
		for _, order := range orders {
			orderNos = append(orderNos, order.OrderNo)
		}

		var queryResults []QueryResult
		switch orderType {
		case OrderTypeYZ:
			if err := db.Raw(`
				SELECT order_no, pro_name as product_name, 1 as amount,
				       total_amount as order_amount, pay_amount
				FROM car_shop_yz_order_v 
				WHERE order_no IN (?)
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}

		case OrderTypeDD, OrderTypeDA:
			if err := db.Raw(`
				SELECT order_no, product_name, amount,
				       order_amount, pay_amount
				FROM car_shop_dadi_order 
				WHERE order_no IN (?)
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}
		case OrderTypeDS:
			if err := db.Raw(`
				SELECT split_no as order_no, group_concat(product_name) as product_name, sum(amount) as amount,
				       sum(order_amount) as order_amount, sum(pay_amount) as pay_amount
				FROM car_shop_dadi_order  
				WHERE split_no IN (?)
				GROUP BY split_no
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}
		case OrderTypeCA:
			if err := db.Raw(`
				SELECT order_no as order_no, '台历' as product_name, 1 as amount,
				       total_amount as order_amount, pay_amount as pay_amount
				FROM car_order_calendar  
				WHERE order_no IN (?)
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}

		case OrderTypeVC:
			// VC类型需要特殊处理，先用订单号查询
			if err := db.Raw(`
				SELECT order_no, product_name, 1 as amount,
				       amount as order_amount, pay_amount, pay_no
				FROM car_vcard_order 
				WHERE order_no IN (?)
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}

			// 收集未查到的订单，使用支付号再次查询
			var missingOrders []OrderData
			for _, order := range orders {
				found := false
				for _, result := range queryResults {
					if result.OrderNo == order.OrderNo {
						found = true
						break
					}
				}
				if !found && order.PayNo != "" {
					missingOrders = append(missingOrders, order)
				}
			}

			if len(missingOrders) > 0 {
				var payNos []string
				for _, order := range missingOrders {
					payNos = append(payNos, order.PayNo)
				}

				var additionalResults []QueryResult
				if err := db.Raw(`
					SELECT order_no, product_name, 1 as amount,
					       amount as order_amount, pay_amount, pay_no
					FROM car_vcard_order 
					WHERE pay_no IN (?) AND status <> '04'
				`, payNos).Scan(&additionalResults).Error; err != nil {
					return nil, err
				}

				queryResults = append(queryResults, additionalResults...)
			}

		case OrderTypeGP:
			if err := db.Raw(`
				SELECT o.order_no, 
				       CASE o.pro_id 
				           WHEN 'PA001' THEN '平安专版水晶相框'
				           WHEN 'PA002' THEN '平安专版私定相册'
				           WHEN 'GS001' THEN '中国人寿专版水晶相框'
				           WHEN 'GS002' THEN '中国人寿专版时光相册'
				           WHEN 'TP001' THEN '中国太平专版水晶相框'
				       END as product_name,
				       o.num as amount,
				       o.order_amount,
				       o.pay_amount
				FROM car_order_gdpa o
				WHERE order_no IN (?)
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}

		case OrderTypeZY:
			if err := db.Raw(`
				SELECT order_no, product_name, amount,
				       order_amount, pay_amount
				FROM car_shop_order_v 
				WHERE order_no IN (?)
			`, orderNos).Scan(&queryResults).Error; err != nil {
				return nil, err
			}
		}

		// 将查询结果存入map
		for _, result := range queryResults {
			results[result.OrderNo] = result
		}
	}

	return results, nil
}

func (p *PayOrder) updateExcelWithResults(f *excelize.File, results map[string]QueryResult) error {
	// 重新读取Excel以获取行信息
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}

	for i, row := range rows {
		if i == 0 || len(row) <= 10 {
			continue
		}

		orderNo := strings.TrimSpace(strings.ReplaceAll(row[10], "`", ""))
		if orderNo == "" {
			continue
		}

		if (orderNo[:2] == "10" || orderNo[:2] == "wx") && len(row) > 11 {
			orderNo = strings.TrimSpace(strings.ReplaceAll(row[11], "`", ""))
		}

		if result, ok := results[orderNo]; ok {
			rowIndex := i + 1
			productName := result.ProductName
			if productName == "" {
				productName = result.ProName
			}

			amount := result.Amount
			if amount == 0 {
				amount = result.Num
			}
			if amount == 0 {
				amount = 1
			}

			// 更新Excel单元格
			f.SetCellValue("Sheet1", fmt.Sprintf("M%d", rowIndex), productName)
			f.SetCellValue("Sheet1", fmt.Sprintf("N%d", rowIndex), amount)
			f.SetCellValue("Sheet1", fmt.Sprintf("O%d", rowIndex), result.OrderAmount)
			f.SetCellValue("Sheet1", fmt.Sprintf("P%d", rowIndex), result.PayAmount)
		}
	}

	return nil
}

func (p *PayOrder) handleError(c *gin.Context, message string, err error) {
	slog.Error(message, err)
	c.String(http.StatusInternalServerError, message)
}

func (p *PayOrder) ExcelFix(c *gin.Context) {
	f, err := excelize.OpenFile("1.xlsx")
	if err != nil {
		p.handleError(c, "打开Excel文件失败", err)
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		fmt.Errorf("no sheets found in excel file")
		return
	}
	sheetName := sheets[0]

	colNum := 8
	// 获取列名
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Errorf("failed to get rows: %v", err)
		return
	}

	colName := "I"

	for rowIdx, row := range rows {
		if colNum >= len(row) {
			continue
		}

		original := row[colNum]
		cleaned := cleanHTML(original)

		if rowIdx == 1 {
			fmt.Printf("Row %d - Original: %s\n", rowIdx+1, original)
			fmt.Printf("Row %d - Cleaned: %s\n", rowIdx+1, cleaned)
		}

		// 设置单元格值
		cellRef := fmt.Sprintf("%s%d", colName, rowIdx+1)
		if err := f.SetCellValue(sheetName, cellRef, cleaned); err != nil {
			fmt.Errorf("failed to set cell %s: %v", cellRef, err)
			return
		}
	}

	// Save the modified Excel file
	if err := f.SaveAs("2.xlsx"); err != nil {
		fmt.Errorf("failed to save modified excel: %v", err)
		return
	}

	c.String(http.StatusOK, "处理完成")
}

func cleanHTML(input string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned := re.ReplaceAllString(input, "")

	// Remove extra whitespace and newlines
	cleaned = strings.TrimSpace(cleaned)
	// Replace multiple newlines with single newline
	reNewline := regexp.MustCompile(`\n\s*\n+`)
	cleaned = reNewline.ReplaceAllString(cleaned, "\n\n")

	return cleaned
}

func (p *PayOrder) Ddkf(c *gin.Context) {
	f, err := excelize.OpenFile("失败清单.xlsx")
	if err != nil {
		p.handleError(c, "打开Excel文件失败", err)
		return
	}
	defer f.Close()
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		p.handleError(c, "读取工作表失败", err)
		return
	}
	var orderNo []string
	var orderNoStd []string
	for i, row := range rows {
		if i == 0 || len(row) < 2 {
			continue
		}
		if row[1][:5] == "wh_std" {
			orderNoStd = append(orderNoStd, row[1])
		} else {
			orderNo = append(orderNo, row[1])
		}

	}
	if len(orderNo) <= 0 {
		p.handleError(c, "订单号不存在！", nil)
	}

	db := model.RDB[model.MASTER].Db
	type list struct {
		Serial    string `json:"serial"`
		Flag      string `json:"flag"`
		Coupon_id int    `json:"coupon_id"`
		Status    int    `json:"status"`
		Name      string `json:"name"`
		BackTime  string `json:"back_time"`
	}

	var lists []list
	var liststds []list

	err = db.Raw("select a.serial,a.flag,a.coupon_id,b.status,FROM_UNIXTIME(a.back_time,'%Y-%m-%d %H:%i:%s') back_time,c.name from car_ddkf_coupon_list a left join car_coupon b on a.coupon_id = b.id left join car_api_product c on a.pro_code = c.code where a.serial in (?)", orderNo).Scan(&lists).Error
	if err != nil {
		p.handleError(c, "订单查询失败", err)
	}
	if len(lists) <= 0 {
		p.handleError(c, "订单号不存在！", nil)
	}
	results := make(map[string]list)
	for _, v := range lists {
		results[v.Serial] = v
	}

	err = db.Raw("select a.serial,a.flag,a.coupon_id,b.status,FROM_UNIXTIME(a.back_time,'%Y-%m-%d %H:%i:%s') back_time from car_ddkf_coupon_info a left join car_coupon b on a.coupon_id = b.id left join car_api_product c on a.pro_code = c.code where a.serial in (?)", orderNo).Scan(&liststds).Error
	if err != nil {
		p.handleError(c, "订单查询失败", err)
	}
	if len(liststds) > 0 {
		for _, v := range liststds {
			results[v.Serial] = v
		}
	}
	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}
		if result, ok := results[row[1]]; ok {
			rowIndex := i + 1

			// 更新Excel单元格
			f.SetCellValue("Sheet1", fmt.Sprintf("F%d", rowIndex), result.Coupon_id)
			f.SetCellValue("Sheet1", fmt.Sprintf("G%d", rowIndex), result.Name)
			f.SetCellValue("Sheet1", fmt.Sprintf("H%d", rowIndex), result.Status)
			f.SetCellValue("Sheet1", fmt.Sprintf("I%d", rowIndex), result.Flag)
			f.SetCellValue("Sheet1", fmt.Sprintf("J%d", rowIndex), result.BackTime)
			fmt.Printf("Row %d - Serial: %s, Flag: %s, Status: %d, BackTime: %s\n", rowIndex, result.Serial, result.Flag, result.Status, result.BackTime)
		}
	}

	// 5. 保存文件
	if err := f.SaveAs("output.xlsx"); err != nil {
		p.handleError(c, "保存文件失败", err)
		return
	}

	c.String(http.StatusOK, "处理完成")

}

func (p *PayOrder) Zking(c *gin.Context) {
	f, err := excelize.OpenFile("紫金订单.xlsx")
	if err != nil {
		p.handleError(c, "打开Excel文件失败", err)
		return
	}
	defer f.Close()
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		p.handleError(c, "读取工作表失败", err)
		return
	}
	var orderNo []string
	for i, row := range rows {
		if i == 0 || len(row) < 2 {
			continue
		}
		if row[1] == "#N/A" {
			orderNo = append(orderNo, row[0])
		}

	}

	if len(orderNo) <= 0 {
		p.handleError(c, "订单号不存在！", nil)
	}

	db := model.RDB[model.MASTER].Db
	type list struct {
		Serial_no string `json:"serial_no"`
		Flag      string `json:"flag"`
		Coupon_id int    `json:"coupon_id"`
		Status    int    `json:"status"`
		Name      string `json:"name"`
		BackTime  string `json:"back_time"`
	}

	var lists []list

	err = db.Raw("select a.serial_no,a.flag,a.coupon_id,b.status,FROM_UNIXTIME(a.back_time,'%Y-%m-%d %H:%i:%s') back_time,c.name from car_zking_coupon_list a left join car_coupon b on a.coupon_id = b.id left join car_api_product c on a.pro_code = c.code where a.serial_no in (?)", orderNo).Scan(&lists).Error
	if err != nil {
		p.handleError(c, "订单查询失败", err)
	}
	if len(lists) <= 0 {
		p.handleError(c, "订单号不存在！", nil)
	}
	results := make(map[string]list)
	for _, v := range lists {
		results[v.Serial_no] = v
	}

	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}
		if result, ok := results[row[0]]; ok {
			rowIndex := i + 1

			// 更新Excel单元格
			f.SetCellValue("Sheet1", fmt.Sprintf("M%d", rowIndex), result.Coupon_id)
			f.SetCellValue("Sheet1", fmt.Sprintf("N%d", rowIndex), result.Name)
			f.SetCellValue("Sheet1", fmt.Sprintf("O%d", rowIndex), result.Status)
			f.SetCellValue("Sheet1", fmt.Sprintf("P%d", rowIndex), result.Flag)
			f.SetCellValue("Sheet1", fmt.Sprintf("Q%d", rowIndex), result.BackTime)
			fmt.Printf("Row %d - Serial: %s, Flag: %s, Status: %d, BackTime: %s\n", rowIndex, result.Serial_no, result.Flag, result.Status, result.BackTime)
		}
	}

	// 5. 保存文件
	if err := f.SaveAs("output.xlsx"); err != nil {
		p.handleError(c, "保存文件失败", err)
		return
	}

	c.String(http.StatusOK, "处理完成")
}

func (p *PayOrder) ProductNum(c *gin.Context) {
	now := time.Now()
	loc := now.Location()

	// 计算本周六 0 点
	weekday := int(now.Weekday())
	if weekday == 0 { // Go 里 Sunday = 0
		weekday = 7
	}
	// 还有几天到周六
	daysUntilSat := 5 - weekday
	thisSaturday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).
		AddDate(0, 0, daysUntilSat)

	// 上周六 0 点 = 本周六往前推 7 天
	lastSaturday := thisSaturday.AddDate(0, 0, -7)

	startTime := lastSaturday.Unix()
	endTime := thisSaturday.Unix()

	sql := `SELECT a.product_name,sum(a.num) as num from (select product_name,sum(amount) as num from car_shop_dadi_order where status in(1,2,3,9) and c_time > ? and c_time < ? GROUP BY spu_id
			union all
			SELECT  product_name,sum(amount) as num from car_shop_order_v where status in(1,2,3,9)  and c_time > ? and c_time < ? GROUP BY product_id
		) a GROUP BY a.product_name`

	db := model.RDB[model.MASTER]
	type Result struct {
		Product_name string `json:"product_name" tag:"产品名称"`
		Num          string `json:"num" tag:"下单数量"`
	}
	var result []Result
	db.Db.Raw(sql, startTime, endTime, startTime, endTime).Find(&result)

	utils.Down(result, "产品下单数量", c)

}
