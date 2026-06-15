package controllers

import (
	"fmt"
	"h5/pkg/model"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

const diyAgentCompany = 88

type diyAgentMonthlyStat struct {
	Month    string `gorm:"column:month"`
	WorkNum  string `gorm:"column:work_num"`
	Name     string `gorm:"column:name"`
	Mobile   string `gorm:"column:mobile"`
	OrgName  string `gorm:"column:org_name"`
	TotalNum int64  `gorm:"column:total_num"`
	OrderNum int64  `gorm:"column:order_num"`
	AgreeNum int64  `gorm:"column:agree_num"`
}

func (*ExportExcel) AgentStatistics(c *gin.Context) {
	startTime, endTime, err := parseMonthRange(c.Query("start_month"), c.Query("end_month"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	whereTime := ""
	args := []interface{}{diyAgentCompany}
	if !startTime.IsZero() {
		whereTime = " AND c.c_time >= ? AND c.c_time < ?"
		args = append(args, startTime.Unix(), endTime.Unix())
	}
	args = append(args, diyAgentCompany)

	sqlQuery := `
		SELECT
			s.month,
			s.agt_work_num AS work_num,
			COALESCE(a.contact, '') AS name,
			COALESCE(a.mobile, '') AS mobile,
			COALESCE(a.org_name, '') AS org_name,
			s.total_num,
			s.order_num,
			s.agree_num
		FROM (
			SELECT
				DATE_FORMAT(FROM_UNIXTIME(c.c_time), '%Y-%m') AS month,
				c.agt_work_num,
				COUNT(*) AS total_num,
				SUM(c.order_time > 0) AS order_num,
				SUM(c.agree_time > 0) AS agree_num
			FROM cs_diy_cus c
			WHERE c.company = ? AND c.status <> -1 AND c.c_time > 0 AND c.agt_work_num <> ''` + whereTime + `
			GROUP BY DATE_FORMAT(FROM_UNIXTIME(c.c_time), '%Y-%m'), c.agt_work_num
		) s
		LEFT JOIN cs_diy_agt a ON a.id = (
			SELECT MAX(a2.id)
			FROM cs_diy_agt a2
			WHERE a2.company = ? AND a2.work_num = s.agt_work_num
		)
		ORDER BY s.month, a.org_name, s.agt_work_num`

	var result []diyAgentMonthlyStat
	db := model.RDB["db1"]
	if err := db.Db.Raw(sqlQuery, args...).Scan(&result).Error; err != nil {
		c.String(http.StatusInternalServerError, "查询代理人统计数据失败: %v", err)
		return
	}
	if len(result) == 0 {
		c.String(http.StatusOK, "暂无代理人统计数据")
		return
	}

	if err := writeAgentStatisticsWorkbook(c, result); err != nil {
		c.String(http.StatusInternalServerError, "生成代理人统计表失败: %v", err)
	}
}

func parseMonthRange(startMonth, endMonth string) (time.Time, time.Time, error) {
	if startMonth == "" && endMonth == "" {
		return time.Time{}, time.Time{}, nil
	}
	if startMonth == "" {
		startMonth = endMonth
	}
	if endMonth == "" {
		endMonth = startMonth
	}

	location := time.Local
	start, err := time.ParseInLocation("2006-01", startMonth, location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("start_month 格式必须为 YYYY-MM")
	}
	end, err := time.ParseInLocation("2006-01", endMonth, location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("end_month 格式必须为 YYYY-MM")
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_month 不能早于 start_month")
	}

	return start, end.AddDate(0, 1, 0), nil
}

func writeAgentStatisticsWorkbook(c *gin.Context, data []diyAgentMonthlyStat) error {
	f := excelize.NewFile()
	defer f.Close()

	headers := []interface{}{
		"月份", "代理人工号", "代理人姓名", "代理人手机号", "机构名称", "客户数", "订单数", "上传图片数",
	}

	currentMonth := ""
	rowNum := 1
	for _, item := range data {
		if item.Month != currentMonth {
			currentMonth = item.Month
			rowNum = 1
			if len(f.GetSheetList()) == 1 && f.GetSheetList()[0] == "Sheet1" {
				if err := f.SetSheetName("Sheet1", currentMonth); err != nil {
					return err
				}
			} else if _, err := f.NewSheet(currentMonth); err != nil {
				return err
			}
			if err := f.SetSheetRow(currentMonth, "A1", &headers); err != nil {
				return err
			}
			if err := f.SetPanes(currentMonth, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
				return err
			}
			if err := f.SetColWidth(currentMonth, "A", "H", 16); err != nil {
				return err
			}
		}

		rowNum++
		row := []interface{}{
			item.Month, item.WorkNum, item.Name, item.Mobile, item.OrgName, item.TotalNum, item.OrderNum, item.AgreeNum,
		}
		cell := fmt.Sprintf("A%d", rowNum)
		if err := f.SetSheetRow(currentMonth, cell, &row); err != nil {
			return err
		}
	}

	filename := fmt.Sprintf("代理人数据统计-%s.xlsx", time.Now().Format("2006-01-02"))
	disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", disposition)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	return f.Write(c.Writer)
}
