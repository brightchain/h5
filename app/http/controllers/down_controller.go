package controllers

import (
	"archive/zip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"h5/app/http/models"
	"h5/pkg/model"
	"h5/utils"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DownOrder struct {
}

func (d *DownOrder) MouseOrderDown(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := ""
	if page > 0 {
		limit = fmt.Sprintf("%d,100", (page-1)*100)
	}

	// 获取其他参数
	status, _ := strconv.Atoi(c.DefaultQuery("status", "1"))
	sdate := c.DefaultQuery("sdate", "")
	edate := c.DefaultQuery("edate", "")
	name := c.DefaultQuery("name", "")
	sign := c.DefaultQuery("sign", "")

	if len(name) <= 0 {
		c.String(200, "缺少参数！")
		return
	}
	h := md5.New()
	io.WriteString(h, name+"13687391951")
	if fmt.Sprintf("%x", h.Sum(nil)) != sign {
		c.String(200, "sign签名失败！")
		return
	}

	// 构建 WHERE 条件
	where := fmt.Sprintf("`status` = %d and company = 5", status)

	if sdate != "" && edate != "" {
		where += fmt.Sprintf(" AND `c_time` >= '%s' AND `c_time` < '%s'", sdate, edate)
	}

	type Result struct {
		Order_no      string `json:"order_no" tag:"订单编号"`
		Ship_name     string `json:"ship_name" tag:"物流公司"`
		Ship_no       string `json:"ship_no" tag:"物流单号"`
		Contact       string `json:"contact" tag:"收货人"`
		Mobile        string `json:"mobile" tag:"收货手机"`
		Customer_info string `json:"customer_info" tag:"客户姓名"`
		Province      string `json:"province" tag:"省"`
		City          string `json:"city" tag:"市"`
		Area          string `json:"area" tag:"区"`
		Img_type      string `json:"img_type" tag:"图片类型"`
		Address       string `json:"address" tag:"详细地址"`
		Remark        string `json:"remark" tag:"备注"`
		C_time        string `json:"c_time" tag:"创建时间"`
	}

	var result []Result
	db := model.RDB[model.MASTER]
	sqlQuery := fmt.Sprintf("select order_no,contact,mobile,province,city,area,address,customer_info,ship_name,ship_no,c_time,img_type from car_order_tshirt where %s limit %s", where, limit)
	fmt.Print(sqlQuery)
	db.Db.Raw(sqlQuery).Find(&result)
	if len(result) == 0 {
		c.String(200, "查询失败！")
		return
	}
	type Customer struct {
		Contact   string `json:"contact"`
		Mobile    string `json:"mobile"`
		Work_num  string `json:"work_num"`
		Work_name string `json:"work_name"`
	}
	path := "./storage/app/public"
	zipName := path + "/" + name + ".zip"
	newZipFile, err := os.Create(zipName)
	if err != nil {
		slog.Error("zip create fail", err)
		c.String(200, "zip 创建失败")
		return
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	orderDirectory := "/home/www/car/static/upload/tshirt/order"
	var orderNos []string
	for k, v := range result {
		orderNos = append(orderNos, v.Order_no)
		result[k].Province = AddChineseProvinceSuffix(v.Province)
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
				result[k].Remark = tom.Mobile
				v.Remark = fmt.Sprintf("工号：%s 客户信息：%s %s", tom.Work_num, tom.Contact, tom.Mobile)
				if tom.Work_name != "" {
					v.Remark = fmt.Sprintf("代理人：%s %s", tom.Work_name, v.Remark)
				}
				result[k].Remark = v.Remark
				result[k].Address += v.Remark
			}
		}

		zName := v.Order_no + v.Contact + ".jpg"
		timestamp, err := strconv.ParseInt(v.C_time, 10, 64)
		if err != nil {
			fmt.Println("时间戳转换错误:", err)
			return
		}
		t := time.Unix(timestamp, 0)
		date := t.Format("2006-01-02")
		result[k].C_time = t.Format("2006-01-02 15:04:05")
		orderNo := strings.ToLower(v.Order_no)
		orderFile := fmt.Sprintf("%s/%s/%s/white_front.%s", orderDirectory, date, orderNo, v.Img_type)
		err = utils.AddFileToZip(zipWriter, orderFile, zName)
		if err != nil {
			slog.Warn("压缩失败", err)
			break
		}
	}
	if err != nil {
		newZipFile.Close()
		os.Remove(zipName)
		c.String(200, "zip 压缩失败")
		return
	}

	fileName := path + "/" + name + ".xlsx"
	utils.SaveFile(result, fileName)

	err = utils.AddFileToZip(zipWriter, fileName, "")
	if err != nil {
		slog.Warn("压缩失败")
		c.String(200, "zip 压缩失败")
		return
	}
	zipWriter.Close()
	newZipFile.Close()
	if len(orderNos) > 0 {
		// 获取当前时间戳
		currentTimestamp := time.Now().Unix() // 秒级时间戳
		// 批量更新 status 和 send_time
		db.Db.Model(&models.CarOrderTshirt{}).
			Where("order_no IN ?", orderNos).
			Updates(map[string]interface{}{
				"status":    4,
				"send_time": currentTimestamp,
			})

		fmt.Printf("Updated status and send_time for orders: %v\n", orderNos)
	} else {
		fmt.Println("No orders found to update")
	}
	filename := filepath.Base(zipName)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(zipName)
	
}

// AddChineseProvinceSuffix 给省份名称添加后缀
func AddChineseProvinceSuffix(provinceName string) string {
	// 直辖市列表
	municipalities := []string{"北京", "上海", "天津", "重庆"}

	// 自治区列表
	autonomousRegions := map[string]string{
		"内蒙古": "内蒙古自治区",
		"新疆":  "新疆维吾尔自治区",
		"广西":  "广西壮族自治区",
		"宁夏":  "宁夏回族自治区",
		"西藏":  "西藏自治区",
	}

	// 特别行政区列表
	specialAdministrativeRegions := []string{"香港", "澳门"}

	// 检查是否已经包含后缀
	matched, _ := regexp.MatchString("省|市|自治区", provinceName)
	if matched {
		return provinceName
	}

	// 处理自治区
	if suffix, exists := autonomousRegions[provinceName]; exists {
		return suffix
	}

	// 处理直辖市
	for _, city := range municipalities {
		if city == provinceName {
			return provinceName + "市"
		}
	}

	// 处理特别行政区
	for _, sar := range specialAdministrativeRegions {
		if sar == provinceName {
			return provinceName + "特别行政区"
		}
	}

	// 处理普通省份
	return provinceName + "省"
}
