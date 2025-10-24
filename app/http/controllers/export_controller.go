package controllers

import (
	"encoding/json"
	"fmt"
	"h5/pkg/model"
	wechabot "h5/pkg/wechaBot"
	"h5/utils"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ExportExcel struct{}

func (*ExportExcel) Xinhua(c *gin.Context) {
	type Result struct {
		Code      string `json:"code" tag:"编码"`
		Status    string `json:"status" tag:"状态"`
		Remark    string `json:"remark" tag:"标记"`
		Mobile    string `json:"mobile" tag:"手机号"`
		Contact   string `json:"contact" tag:"联系人"`
		Organ     string `json:"organ" tag:"机构"`
		Work_num  string `json:"work_num" tag:"工号"`
		Order_no  string `json:"order_no" tag:"订单号"`
		Contact1  string `json:"contact1" tag:"收货人"`
		Mobile1   string `json:"mobile1" tag:"收货手机"`
		Province  string `json:"province" tag:"省"`
		City      string `json:"city" tag:"市"`
		Area      string `json:"area" tag:"区"`
		Address   string `json:"address" tag:"地址"`
		Ship_name string `json:"ship_name" tag:"快递公司"`
		Ship_no   string `json:"ship_no" tag:"快递单号"`
		C_time    string `json:"c_time" tag:"创建时间"`
	}

	var result []Result

	sqlQuery := "select a.code,a.status,a.remark,c.work_num,c.mobile,c.contact,c.organ,d.order_no,d.contact as contact1,d.mobile as mobile1,d.province,d.city,d.area,d.address,d.ship_name,d.ship_no,d.c_time from car_coupon a left join car_member b on a.user_id = b.id LEFT JOIN car_order_photo_agent c  on b.mobile = c.mobile and c.company = 19 LEFT JOIN car_order_photo d on a.id = d.coupon_id and d.status != -1 where a.batch_num = 'P2401291323' and a.status !=0"

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	for k, v := range result {
		status := "已激活"
		remark := "未分享"
		num, _ := strconv.Atoi(v.Status)
		if v.Remark == "1" {
			remark = "已分享"
		} else if v.Remark != "" && v.Remark != "1" {
			remark = "已领取"
		}
		if num == 2 {
			status = "已下单"
			remark = "已下单"
		}
		result[k].Status = status
		result[k].Remark = remark
		if v.C_time != "" {
			result[k].C_time = utils.FormatDateByString(v.C_time)
		}
	}

	utils.Down(result, "新华保险摆台", c)
}

func (*ExportExcel) Fjpa(c *gin.Context) {

	type Result struct {
		Name          string `json:"nume" tag:"业务员姓名"`
		Mobile        string `json:"mobile" tag:"手机号"`
		Work_num      string `json:"work_num" tag:"工号"`
		Contact       string `json:"contact" tag:"营服"`
		Organ         string `json:"organ" tag:"机构"`
		Num           int    `json:"num" tag:"权益数量"`
		Code          string `json:"code" tag:"卡券编码"`
		Active_time   string `json:"active_time" tag:"激活时间"`
		Remark        string `json:"remark" tag:"分享状态"`
		Status        string `json:"status" tag:"卡券状态"`
		Order_no      string `json:"order_no" tag:"订单号"`
		Contact1      string `json:"contact1" tag:"收货人"`
		Mobile1       string `json:"mobile1" tag:"收货手机"`
		Customer_info string `json:"customer_info" tag:"客户备注"`
		Address       string `json:"address" tag:"地址"`
		Ship_name     string `json:"ship_name" tag:"快递公司"`
		Ship_no       string `json:"ship_no" tag:"快递单号"`
		C_time        string `json:"c_time" tag:"创建时间"`
	}

	var result []Result

	sqlQuery := "select a.name,a.mobile,a.work_num,a.contact ,a.organ ,b.num ,c.code ,if(c.active_time <>0,FROM_UNIXTIME(c.active_time),'') as active_time,if(c.remark is NULL,'未分享',if(c.remark=1,'已分享','已领取')) as 'remark',case c.status when '0' then '' when 1 then '已激活' when 2 then '已下单' end as status ,d.order_no,d.contact as cus_contact,d.mobile as cus_mobile,concat(d.province,d.city,d.area,d.address) as address,d.customer_info,d.ship_no,d.ship_name,if(d.c_time<>0,FROM_UNIXTIME(d.c_time),'') as c_time from car_order_photo_agent a left join ( select mobile,sum(num) as num from car_member_bind_logs where coupon_batch = 'P2403121036' and is_del = 0 GROUP BY mobile) b on a.mobile = b.mobile LEFT JOIN car_coupon c on c.batch_num = 'P2403121036' and c.mobile = a.mobile LEFT JOIN car_order_photo d on c.id = d.coupon_id and d.`status` != -1 where a.company = 22"
	type Customer struct {
		Contact  string `json:"contact"`
		Work_num int    `json:"work_num"`
	}

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	for k, v := range result {
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
			}
		}
	}

	utils.Down(result, "福建平安摆台", c)

}

func (*ExportExcel) Ydln(c *gin.Context) {
	type Result struct {
		Code            string `json:"code" tag:"卡券编号"`
		Name            string `json:"name" tag:"卡券名称"`
		Sn              string `json:"sn" tag:"卡券编码"`
		Password        string `json:"password" tag:"兑换码"`
		Active_time     string `json:"active_time" tag:"激活时间"`
		Remark          string `json:"remark" tag:"分享状态"`
		Status          string `json:"status" tag:"卡券状态"`
		Order_no        string `json:"order_no" tag:"订单号"`
		Contact1        string `json:"contact1" tag:"收货人"`
		Mobile1         string `json:"mobile1" tag:"收货手机"`
		Customer_info   string `json:"customer_info" tag:"客户姓名"`
		Customer_mobile string `json:"customer_mobile" tag:"客户手机"`
		Address         string `json:"address" tag:"地址"`
		Ship_name       string `json:"ship_name" tag:"快递公司"`
		Ship_no         string `json:"ship_no" tag:"快递单号"`
		C_time          string `json:"c_time" tag:"创建时间"`
	}

	var result []Result

	sqlQuery := "select a.code,a.name,a.sn,a.`password`,if(a.active_time,FROM_UNIXTIME(a.active_time),'') as active_time,a.mobile,if(b.remark is NULL,'未分享',if(b.remark=1,'已分享','已领取')) as remark,case b.status when '0' then '' when 1 then '已激活' when 2 then '已下单' end as status ,c.order_no,c.contact as contact1,c.mobile as mobile1,concat(c.province,c.city,c.area,c.address) as address,c.customer_info,c.ship_no,c.ship_name,if(c.c_time<>0,FROM_UNIXTIME(c.c_time),'') as c_time from car_coupon_pkg a LEFT JOIN car_coupon b on (a.id = b.pkg_id) LEFT JOIN car_order_photo c on c.coupon_id = b.id and c.`status` <> -1 WHERE a.batch_num in ('PP2403041702','PP2403061702')"
	type Customer struct {
		Contact string `json:"contact"`
		Mobile  string `json:"mobile"`
	}

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	for k, v := range result {
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
				result[k].Customer_mobile = tom.Mobile
			}
		}
	}

	utils.Down(result, "英大辽宁摆台", c)
}

func (*ExportExcel) ShTp(c *gin.Context) {
	type Result struct {
		Code            string `json:"code" tag:"卡券编号"`
		Name            string `json:"name" tag:"卡券名称"`
		Sn              string `json:"sn" tag:"卡券编码"`
		Password        string `json:"password" tag:"兑换码"`
		Active_time     string `json:"active_time" tag:"激活时间"`
		Mobile          string `json:"mobile" tag:"手机号"`
		Work_num        string `json:"work_num" tag:"业务员工号"`
		Name1           string `json:"name1" tag:"业务员姓名"`
		Organ           string `json:"organ" tag:"机构名称"`
		Remark          string `json:"remark" tag:"分享状态"`
		Status          string `json:"status" tag:"卡券状态"`
		Order_no        string `json:"order_no" tag:"订单号"`
		Contact1        string `json:"contact1" tag:"收货人"`
		Mobile1         string `json:"mobile1" tag:"收货手机"`
		Customer_info   string `json:"customer_info" tag:"客户姓名"`
		Customer_mobile string `json:"customer_mobile" tag:"客户手机"`
		Address         string `json:"address" tag:"地址"`
		Ship_name       string `json:"ship_name" tag:"快递公司"`
		Ship_no         string `json:"ship_no" tag:"快递单号"`
		C_time          string `json:"c_time" tag:"创建时间"`
	}

	var result []Result

	sqlQuery := "select a.code,a.name,a.sn,a.`password`,if(a.active_time,FROM_UNIXTIME(a.active_time),'') as active_time,a.mobile,d.name as name1,d.work_num,d.organ,if(b.remark is NULL,'未分享',if(b.remark=1,'已分享','已领取')) as remark,case b.status when '0' then '未激活' when 1 then '已激活' when 2 then '已下单' end as status ,c.order_no,c.contact as contact1,c.mobile as mobile1,concat(c.province,c.city,c.area,c.address) as address,c.customer_info,c.ship_no,c.ship_name,if(c.c_time<>0,FROM_UNIXTIME(c.c_time),'') as c_time from car_coupon_pkg a LEFT JOIN car_coupon b on (a.id = b.pkg_id) LEFT JOIN car_order_photo_agent d on a.mobile = d.mobile and d.company = 24 and a.mobile <> 0 LEFT JOIN car_order_photo c on c.coupon_id = b.id and c.`status` <> -1 WHERE a.batch_num ='PP2404301621'"
	type Customer struct {
		Contact string `json:"contact"`
		Mobile  string `json:"mobile"`
	}

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	for k, v := range result {
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
				result[k].Customer_mobile = tom.Mobile
			}
		}
	}

	utils.Down(result, "上海太平个险", c)
}

func (*ExportExcel) FjTp(c *gin.Context) {
	type Result struct {
		OrderNo      string `json:"order_no" tag:"订单编号"`
		Name         string `json:"name" tag:"产品名称"`
		Num          string `json:"num" tag:"购买数量"`
		Order_amount string `json:"order_amount" tag:"订单金额"`
		PayNo        string `json:"pay_no" tag:"支付单号"`
		PayAt        string `json:"pay_at" tag:"支付时间"`
		Mobile       string `json:"mobile" tag:"手机号"`
		Work_num     string `json:"work_num" tag:"业务员工号"`
		Name1        string `json:"name1" tag:"业务员姓名"`
		Contact      string `json:"contact" tag:"中支"`
		Organ        string `json:"organ" tag:"营服"`
		Status       string `json:"status" tag:"订单状态"`
		C_time       string `json:"c_time" tag:"创建时间"`
	}

	var result []Result

	sqlQuery := "select a.order_no, '福建太平10寸照片摆台' as name,a.num,a.order_amount,a.pay_no,if(a.pay_at,FROM_UNIXTIME(a.pay_at),'') as 'pay_at',case a.status when 0 then '未付款' when 1 then '已付款' when 2 then '已完成' when -1 then '已取消' end as 'status',b.name as 'name1',b.mobile,b.contact,b.organ,b.work_num,FROM_UNIXTIME(a.c_time) as 'c_time'  from car_order_gdpa a LEFT JOIN car_order_photo_agent b on (a.uid = b.uid and b.company = 30) where a.pro_id = 'TP001' "

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "福建太平摆台购买", c)
}

func (*ExportExcel) Hngx(c *gin.Context) {
	at := c.Query("at")
	if at != "sfdjwie2ji239324" {
		slog.Error("非法访问")
		c.String(200, "非法访问")
		return
	}

	type Result struct {
		Sn            string `json:"sn" tag:"卡券编号"`
		Password      string `json:"password" tag:"兑换码"`
		Status        string `json:"status" tag:"状态"`
		Active_time   string `json:"active_time" tag:"激活时间"`
		Order_no      string `json:"order_no" tag:"订单号"`
		Contact       string `json:"contact" tag:"联系人"`
		Mobile        string `json:"mobile" tag:"手机号"`
		Province      string `json:"province" tag:"省"`
		City          string `json:"city" tag:"市"`
		Area          string `json:"area" tag:"区"`
		Address       string `json:"address" tag:"地址"`
		Organ         string `json:"organ" tag:"机构"`
		Work_num      string `json:"work_num" tag:"工号"`
		Customer_info string `json:"customer_info" tag:"客户姓名"`
		Cus_mobile    string `json:"cus_mobile" tag:"客户手机"`
		Ship_name     string `json:"ship_name" tag:"快递公司"`
		Ship_no       string `json:"ship_no" tag:"快递单号"`
	}

	var result []Result

	sqlQuery := "select a.active_time,a.status,b.sn,b.password,c.order_no,c.contact,c.mobile,c.province,c.city,c.area,c.address,c.customer_info,c.ship_name,c.ship_no,c.organ,c.work_num from car_coupon a left join  car_coupon_pkg b on a.pkg_id = b.id left join car_order_photo c on a.id = c.coupon_id where a.tp_code = 'CT000564' and a.status in(1,2) and a.active_time >1704038400"
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	for k, v := range result {
		type Customer struct {
			Contact  string `json:"contact"`
			Work_num int    `json:"work_num"`
		}

		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
			}
		}

		status := "已激活"
		num, _ := strconv.Atoi(v.Status)
		if num == 2 {
			status = "已下单"
		}
		result[k].Status = status
		if v.Active_time != "0" {
			result[k].Active_time = utils.FormatDateByString(v.Active_time)
		}
	}

	utils.Down(result, "河南阳光个险", c)
}

func (*ExportExcel) Hnkj(c *gin.Context) {
	at := c.Query("at")
	if at != "sfdjwie2ji239324" {
		c.String(200, "非法访问")
		return
	}

	type Result struct {
		Sn            string `json:"sn" tag:"卡券编号"`
		Password      string `json:"password" tag:"兑换码"`
		Status        string `json:"status" tag:"状态"`
		Active_time   string `json:"active_time" tag:"激活时间"`
		Order_no      string `json:"order_no" tag:"订单号"`
		Contact       string `json:"contact" tag:"联系人"`
		Mobile        string `json:"mobile" tag:"手机号"`
		Province      string `json:"province" tag:"省"`
		City          string `json:"city" tag:"市"`
		Area          string `json:"area" tag:"区"`
		Address       string `json:"address" tag:"地址"`
		Organ         string `json:"organ" tag:"机构"`
		Work_num      string `json:"work_num" tag:"工号"`
		Customer_info string `json:"customer_info" tag:"客户姓名"`
		Cus_mobile    string `json:"cus_mobile" tag:"客户手机"`
		Ship_name     string `json:"ship_name" tag:"快递公司"`
		Ship_no       string `json:"ship_no" tag:"快递单号"`
	}

	var result []Result

	sqlQuery := "select a.active_time,a.status,b.sn,b.password,c.order_no,c.contact,c.mobile,c.province,c.city,c.area,c.address,c.customer_info,c.ship_name,c.ship_no,c.organ,c.work_num from car_coupon a left join  car_coupon_pkg b on a.pkg_id = b.id left join car_order_photo c on a.id = c.coupon_id where a.tp_code = 'CT001089' and a.status in(1,2) "
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	type Customer struct {
		Contact  string `json:"contact"`
		Work_num int    `json:"work_num"`
	}

	for k, v := range result {
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
			}
		}

		status := "已激活"
		num, _ := strconv.Atoi(v.Status)
		if num == 2 {
			status = "已下单"
		}
		result[k].Status = status
		if v.Active_time != "0" {
			result[k].Active_time = utils.FormatDateByString(v.Active_time)
		}
	}

	utils.Down(result, "河南阳光客经", c)
}

func (*ExportExcel) Jxms(c *gin.Context) {

	type Result struct {
		Name          string `json:"name" tag:"代理人姓名"`
		Work_num      string `json:"work_num" tag:"工号"`
		Mobile        string `json:"mobile" tag:"手机号"`
		Organ         string `json:"organ" tag:"机构"`
		Status        string `json:"status" tag:"状态"`
		Active_time   string `json:"active_time" tag:"激活时间"`
		Order_no      string `json:"order_no" tag:"订单号"`
		Contact       string `json:"contact" tag:"联系人"`
		Amobile       string `json:"amobile" tag:"手机号"`
		Province      string `json:"province" tag:"省"`
		City          string `json:"city" tag:"市"`
		Area          string `json:"area" tag:"区"`
		Address       string `json:"address" tag:"地址"`
		Customer_info string `json:"customer_info" tag:"客户姓名"`
		Cus_mobile    int    `json:"cus_mobile" tag:"客户手机"`
		Cus_address   string `json:"cus_address" tag:"客户地址"`
		Ship_name     string `json:"ship_name" tag:"快递公司"`
		Ship_no       string `json:"ship_no" tag:"快递单号"`
	}

	var result []Result

	sqlQuery := "select a.active_time,a.status,b.name,b.mobile,b.work_num,b.organ,c.order_no,c.contact,c.mobile,c.province,c.city,c.area,c.address,c.customer_info,c.ship_name,c.ship_no from car_coupon a left join  car_order_photo_worknum b on a.mobile = b.mobile and b.company = 41 left join car_order_photo c on a.id = c.coupon_id and c.status <> -1 where a.batch_num = 'P2501271727' and a.status in(1,2) "
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	type Customer struct {
		Contact  string `json:"contact"`
		Mobile   int    `json:"mobile"`
		Address  string `json:"address"`
		Work_num int    `json:"work_num"`
	}

	for k, v := range result {
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			fmt.Println(tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
				result[k].Cus_mobile = tom.Mobile
				result[k].Cus_address = tom.Address
			}
		}

		status := "已激活"
		num, _ := strconv.Atoi(v.Status)
		if num == 2 {
			status = "已下单"
		}
		result[k].Status = status
		if v.Active_time != "0" {
			result[k].Active_time = utils.FormatDateByString(v.Active_time)
		}
	}

	utils.Down(result, "江西民生摆台", c)
}

func (*ExportExcel) Smwj(c *gin.Context) {
	at := c.Query("at")
	if at != "sfdjwie2ji239324" {
		c.String(200, "非法访问")
		return
	}

	type Result struct {
		Openid        string `json:"openid" tag:"openid"`
		Name          string `json:"name" tag:"名称"`
		Mobile        string `json:"mobile" tag:"手机号"`
		Sex           string `json:"sex" tag:"性别"`
		Question1     string `json:"question1" tag:"答题1"`
		Question2     string `json:"question2" tag:"答题2"`
		Question3     string `json:"question3" tag:"答题3"`
		Question_time string `json:"question_time" tag:"答题时间"`
		Agent_name    string `json:"agent_name" tag:"业务员姓名"`
		Agent_mobile  string `json:"agent_mobile" tag:"业务员手机"`
		Work_num      string `json:"work_num" tag:"工号"`
		Status        string `json:"status" tag:"状态"`
		C_time        string `json:"c_time" tag:"创建时间"`
	}
	sqlQuery := "select openid,name,mobile,sex,question1,question2,question3,question_time,agent_name,agent_mobile,work_num,organ,branch,agent,c_time from (select a.id, a.openid,a.work_num,a.name,a.mobile,a.sex,a.question1,a.question2,a.question3,a.question_time,a.c_time,b.mobile as agent_mobile,b.name as agent_name,b.code,c.agent,c.branch,c.organ from cs_sino_wj a ,cs_sino_cus b ,  car.car_order_photo_organ c where a.work_num = b.work_num and c.code = b.code and c.company = 21   ) as t where 1=1"

	organ, ok := c.GetQuery("organ")
	if ok {
		sqlQuery += fmt.Sprintf(" and code like '%s%%'", organ)
	}

	branch, ok := c.GetQuery("branch")
	if ok {
		sqlQuery += fmt.Sprintf(" and code like '%s%%'", branch)
	}

	agent, ok := c.GetQuery("agent")
	if ok {
		sqlQuery += fmt.Sprintf(" and code like '%s%%'", agent)
	}

	code, ok := c.GetQuery("code")
	if ok {
		sqlQuery += fmt.Sprintf(" and code like '%s%%'", code)
	}

	status, ok := c.GetQuery("status")
	if status == "1" {
		sqlQuery += " and `question_time` = 0"
	}

	if status == "2" {
		sqlQuery += " and `question_time` <> 0"
	}

	sqlQuery += " order by c_time"

	// c.String(200, sqlQuery)
	// return
	var result []Result
	db := model.RDB["db2"]
	db.Db.Raw(sqlQuery).Find(&result)

	for k, v := range result {
		t, _ := strconv.ParseInt(v.Question_time, 10, 64)
		status := "已邀约"
		if t > 0 {
			result[k].Question_time = utils.FormatDate(t)
			status = "已答题"
		}
		result[k].Status = status
		result[k].C_time = utils.FormatDateByString(v.C_time)
	}

	utils.Down(result, "问卷调查", c)
}

func (*ExportExcel) NyOrder(c *gin.Context) {
	type Result struct {
		Serial_no  string `json:"serial_no" tag:"流水号"`
		Pro_code   string `json:"pro_code" tag:"产品编码"`
		Name       string `json:"name" tag:"产品名称"`
		Thd_code   string `json:"thd_code" tag:"用户id"`
		Start_time string `json:"start_time" tag:"权益开始时间"`
		End_time   string `json:"end_time" tag:"权益结束时间"`
		Org_code   string `json:"org_code" tag:"机构代码"`
		Org_name   string `json:"org_name" tag:"机构名称"`
		Status     string `json:"status" tag:"状态"`
		C_time     string `json:"c_time" tag:"创建时间"`
	}

	var result []Result
	sqlQuery := "select a.serial_no as 'serial_no',a.pro_code as 'pro_code',b.`name` as 'name',a.thd_code as 'thd_code',a.start_time as 'start_time',a.end_time as 'end_time',org_code as 'org_code',org_name as 'org_name',case a.status when 1 then '已激活' when 2 then '已使用'when 3 then '已激活' when -1 then '已撤销' end as 'status',FROM_UNIXTIME(c_time) as 'c_time'  from car_nongyin_coupon_list a LEFT JOIN car_api_product b on a.pro_code = b.code  "

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "农业人寿客户生日礼", c)
}

func (*ExportExcel) HntbOrder(c *gin.Context) {
	type Result struct {
		Order_no     string `json:"order_no" tag:"订单号"`
		Name         string `json:"name" tag:"产品名称"`
		Num          string `json:"num" tag:"产品数量"`
		Price        string `json:"price" tag:"产品价格"`
		Organ        string `json:"organ" tag:"机构名称"`
		Work_num     string `json:"work_num" tag:"工号"`
		Contact      string `json:"contact" tag:"收货人"`
		Mobile       string `json:"mobile" tag:"收货手机"`
		Address      string `json:"address" tag:"收货地址"`
		Order_amount string `json:"order_amount" tag:"订单金额"`
		Pay_amount   string `json:"pay_amount" tag:"支付金额"`
		Pay_at       string `json:"pay_at" tag:"支付时间"`
		Status       string `json:"status" tag:"订单状态"`
		C_time       string `json:"c_time" tag:"创建时间"`
	}

	var result []Result
	sqlQuery := "select a.order_no,case c.pro_id when 1 then '日进斗巾厨房湿巾' when 2 then '有两把刷子（天然竹制锅刷）' when 3 then '富得流油厨房清洁套装' when 4 then '一锤定赢养生锤' when 5 then '照片摆台' when 6 then '艾护全身 福到万家 灸贴套装' when 7 then '聚宝罐五件套' when 8 then '法兰绒时尚午睡毯' when 9 then '金龙鱼伴手礼盒' end 'name',c.num 'num',c.price,organ ,work_num ,a.contact ,a.mobile ,concat(a.province,a.city,a.area,a.address) 'address',total_amount 'order_amount' ,c.total_amount 'pay_amount' ,if(pay_at>0,FROM_UNIXTIME(pay_at,'%Y-%m-%d %h:%i:%s'),'') 'pay_at',case status when 0 then '未支付' when 1 then '已付款' end 'status',FROM_UNIXTIME(a.c_time,'%Y-%m-%d %h:%i:%s') 'c_time'  from cs_mall_hntb_order a LEFT JOIN cs_mall_hntb_agent b on a.uid = b.id LEFT JOIN cs_mall_hntb_order_item c on a.order_no = c.order_no WHERE a.status not in(-1,0)"

	db := model.RDB["db3"]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "湖南太保礼品增订", c)
}

func (*ExportExcel) PhotoCancal(c *gin.Context) {
	type Result struct {
		Order_no    string `json:"order_no" tag:"订单号"`
		Contact     string `json:"contact" tag:"收货人"`
		Mobile      string `json:"mobile" tag:"收货手机"`
		Province    string `json:"province" tag:"省"`
		City        string `json:"city" tag:"市"`
		Area        string `json:"area" tag:"区"`
		Address     string `json:"address" tag:"收货地址"`
		Cus_contact string `json:"cus_contact" tag:"客户联系人"`
		Cus_mobile  string `json:"cus_mobile" tag:"客户手机号"`
		Organ       string `json:"organ" tag:"机构名称"`
		Work_num    string `json:"work_num" tag:"工号"`
		Style       string `json:"style" tag:"模板名称"`
		Name        string `json:"name" tag:"卡券名称"`
	}
	now := time.Now()
	loc := now.Location()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	timestamp := startOfDay.Unix()
	weekDay := now.Weekday()
	startOfTime := timestamp - 86400
	if weekDay == 1 {
		startOfTime = timestamp - 259200
	}
	sqlQuery := fmt.Sprintf(`select order_no , contact ,a.mobile ,province,city,area,address ,SUBSTRING_INDEX(REPLACE (customer_info,CONCAT(SUBSTRING_INDEX(customer_info, '"contact":', 1),'"contact":"'),''),'"', 1) as cus_contact,SUBSTRING_INDEX(REPLACE (customer_info,CONCAT(SUBSTRING_INDEX(customer_info, '"mobile":', 1),'"mobile":"'),''),'"', 1) as cus_mobile,organ ,work_num ,style ,c.name  from  car.car_order_photo a LEFT JOIN car.car_coupon b on a.coupon_id = b.id LEFT JOIN car.car_coupon_type c on b.tp_code = c.code  where a.status = -1 and b.status = 1  and a.u_time > %d and a.u_time <= %d`, startOfTime, timestamp)

	db := model.RDB[model.MASTER]
	var result []Result
	err := db.Db.Raw(sqlQuery).Find(&result).Error
	if err != nil {
		fmt.Printf("%v", err)
		c.String(200, "查询失败！")
		return
	}
	path := "./storage/app/public"
	name := fmt.Sprintf("照片摆台工厂取消订单_%s", time.Now().Format("20060102"))
	fileName := path + "/" + name + ".xlsx"
	utils.SaveFile(result, fileName)
	weBot := wechabot.NewWechaBot("bot1")

	res, err := weBot.Upload(fileName)
	if err != nil {
		fmt.Printf("%v", err)
		c.String(200, "文件上传失败！")
		return
	}

	err = weBot.SendFile(res.MediaID)
	if err != nil {
		fmt.Printf("%v", err)
		c.String(200, "消息发送失败！")
		return
	}

	c.String(200, "发送成功！")
	return

}

func (*ExportExcel) GdpaOrder(c *gin.Context) {
	type Result struct {
		Order_no   string `json:"order_no" tag:"订单号"`
		Agt_mobile string `json:"agt_mobile" tag:"业务员手机"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Province   string `json:"province" tag:"省"`
		City       string `json:"city" tag:"市"`
		Area       string `json:"area" tag:"区"`
		Address    string `json:"address" tag:"收货地址"`
		Ship_name  string `json:"ship_name" tag:"快递公司"`
		Ship_no    string `json:"ship_no" tag:"快递单号"`
		Ship_time  string `json:"ship_time" tag:"发货时间"`
		Status     string `json:"status" tag:"状态"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}

	sqlQuery := `
	SELECT 
		a.order_no,
		b.mobile AS agt_mobile,
		a.contact,
		a.mobile,
		a.province,
		a.city,
		a.area,
		a.address,
		a.ship_name,
		a.ship_no,
		IF(a.ship_time > 0, DATE_FORMAT(FROM_UNIXTIME(a.ship_time), '%Y-%m-%d %H:%i:%s'), '') AS ship_time,
		DATE_FORMAT(FROM_UNIXTIME(a.c_time), '%Y-%m-%d %H:%i:%s') AS c_time
	FROM car_order_tshirt a
	JOIN car_coupon b 
	WHERE a.coupon_id = b.id 
	AND a.status <> -1 
	AND b.tp_code = 'CT001604'
`

	db := model.RDB[model.MASTER]
	var result []Result
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "广东平安马克杯订单", c)

}

func (*ExportExcel) GdpaOrderZj(c *gin.Context) {
	type Result struct {
		Code       string `json:"code" tag:"优惠券包编号"`
		Name       string `json:"name" tag:"名称"`
		Sn         string `json:"sn" tag:"序列号"`
		Password   string `json:"password" tag:"兑换码"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		Phone      string `json:"phone" tag:"业务员手机"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
	}
	var result []Result
	sqlQuery := "select b.code,b.name,b.sn,b.`password`,if(b.status =0,'未激活','已激活') status,if(b.active_time,FROM_UNIXTIME(b.active_time, '%Y-%m-%d %H:%i:%s'),'') active_time,b.mobile as phone,d.order_no,d.contact,d.mobile,concat(d.province,d.city,d.area,d.address) address,d.ship_name,d.ship_no from tmp_gdpa a LEFT JOIN car_coupon_pkg b on a.`password` = b.`password` LEFT JOIN car_coupon c on b.id = c.pkg_id LEFT JOIN car_order_tshirt d on c.id = d.coupon_id WHERE a.type = 1"

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "湛江马克杯数据", c)
}

func (*ExportExcel) GdpaOrderZs(c *gin.Context) {
	type Result struct {
		Code       string `json:"code" tag:"优惠券包编号"`
		Name       string `json:"name" tag:"名称"`
		Sn         string `json:"sn" tag:"序列号"`
		Password   string `json:"password" tag:"兑换码"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		Phone      string `json:"phone" tag:"业务员手机"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
	}
	var result []Result
	sqlQuery := "select b.code,b.name,b.sn,b.`password`,if(b.status =0,'未激活','已激活') status,if(b.active_time,FROM_UNIXTIME(b.active_time, '%Y-%m-%d %H:%i:%s'),'') active_time,b.mobile as phone,d.order_no,d.contact,d.mobile,concat(d.province,d.city,d.area,d.address) address,d.ship_name,d.ship_no from tmp_gdpa a LEFT JOIN car_coupon_pkg b on a.`password` = b.`password` LEFT JOIN car_coupon c on b.id = c.pkg_id LEFT JOIN car_order_tshirt d on c.id = d.coupon_id WHERE a.type = 2"

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "中山马克杯数据", c)
}

func (*ExportExcel) GdpaImport(c *gin.Context) {
	type Result struct {
		Organ      string `json:"organ" tag:"机构名称"`
		Name       string `json:"name" tag:"姓名"`
		Work_num   string `json:"work_num" tag:"工号"`
		Mobile     string `json:"mobile" tag:"手机号"`
		Num        int    `json:"num" tag:"权益数量"`
		Active_num int    `json:"active_num" tag:"激活数量"`
		Order_num  int    `json:"order_num" tag:"订单数量"`
		Ship_num   int    `json:"ship_num" tag:"发货数量"`
	}
	var result []Result
	sqlQuery := "SELECT a.organ, a.name, a.work_num, a.mobile, a.num, (SELECT COUNT(id) FROM car_coupon b WHERE b.mobile = a.mobile AND b.batch_num = 'D2502131732') AS active_num, (SELECT COUNT(c.id) FROM car_coupon b LEFT JOIN car_order_tshirt c ON b.id = c.coupon_id AND c.status <> -1 WHERE b.mobile = a.mobile AND b.batch_num = 'D2502131732') AS order_num, (SELECT SUM(CASE WHEN c.ship_no IS NOT NULL AND c.ship_no <> '' THEN 1 ELSE 0 END) FROM car_coupon b LEFT JOIN car_order_tshirt c ON b.id = c.coupon_id AND c.status <> -1 WHERE b.mobile = a.mobile AND b.batch_num = 'D2502131732') AS ship_num FROM tmp_gdpa a WHERE a.type = 3"

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "广东平安手机权益绑定数据", c)
}

func (*ExportExcel) GdpaOrders(c *gin.Context) {
	typeVal := c.Query("type")
	if typeVal == "" {
		c.String(200, "缺少参数！")
		return
	}
	type Result struct {
		Code       string `json:"code" tag:"优惠券包编号"`
		Name       string `json:"name" tag:"名称"`
		Sn         string `json:"sn" tag:"序列号"`
		Password   string `json:"password" tag:"兑换码"`
		Organ      string `json:"organ" tag:"机构名称"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		Phone      string `json:"phone" tag:"业务员手机"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
	}
	var result []Result
	sqlQuery := "select b.code,b.name,b.sn,b.`password`,a.organ,if(b.status =0,'未激活','已激活') status,if(b.active_time,FROM_UNIXTIME(b.active_time, '%Y-%m-%d %H:%i:%s'),'') active_time,b.mobile as phone,d.order_no,d.contact,d.mobile,concat(d.province,d.city,d.area,d.address) address,d.ship_name,d.ship_no from tmp_gdpa a LEFT JOIN car_coupon_pkg b on a.`password` = b.`password` LEFT JOIN car_coupon c on b.id = c.pkg_id LEFT JOIN car_order_tshirt d on c.id = d.coupon_id WHERE a.type = ?"

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery, typeVal).Find(&result)
	name := "广东平安券码征订数据"
	if typeVal == "5" {
		name = "广东平安主管支持数据"
	}
	utils.Down(result, name, c)
}

func (*ExportExcel) YggxOrder(c *gin.Context) {
	type Result struct {
		Phone      string `json:"phone" tag:"业务员手机"`
		Num        string `json:"num" tag:"匹配数量"`
		Sn         string `json:"sn" tag:"序列号"`
		Id         string `json:"id" tag:"卡券ID"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
	}
	var result []Result
	sqlQuery := "select a.mobile phone,a.num,b.sn,b.`id`,case b.`status` when 0 then '未激活' when 1 then '已激活' when 2 then '已下单' when 3 then '已过期' end status,if(b.active_time,FROM_UNIXTIME(b.active_time, '%Y-%m-%d %H:%i:%s'),'') active_time,d.order_no,d.contact,d.mobile,concat(d.province,d.city,d.area,d.address) address,d.ship_name,d.ship_no from car_member_bind_logs a LEFT JOIN car_coupon b on a.`mobile` = b.`mobile` and a.coupon_batch = b.batch_num LEFT JOIN car_order_tshirt d on b.id = d.coupon_id and d.status <> -1 WHERE a.coupon_batch = 'D2503171756'"

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	name := "马克杯下单数据"

	utils.Down(result, name, c)
}

func (*ExportExcel) Hnms(c *gin.Context) {
	type Result struct {
		Mobile     string `json:"mobile" tag:"代理人手机号"`
		Num        string `json:"num" tag:"匹配数量"`
		Name       string `json:"name" tag:"代理人姓名"`
		Organ      string `json:"organ" tag:"机构名称"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		Status     string `json:"status" tag:"状态"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Contact    string `json:"contact" tag:"收货人"`
		Phone      string `json:"phone" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
	}
	var result []Result
	sqlQuery := "select a.mobile,a.num num,d.name,d.organ,if(b.active_time,FROM_UNIXTIME(b.active_time),'') 'active_time',case b.status when '' then '未激活' when 1 then '已激活' when 2 then '已下单' end 'status',c.order_no,c.contact,c.mobile phone,concat(c.province,c.city,c.area,c.address) address,if(c.c_time,FROM_UNIXTIME(c.c_time),'') order_time,c.ship_name,c.ship_no from (SELECT mobile,sum(num) num,coupon_batch from car_member_bind_logs where coupon_batch = 'P2504051322' GROUP BY mobile) a LEFT JOIN car_coupon b on a.mobile = b.mobile and b.batch_num = 'P2504051322' LEFT JOIN car_order_photo c on b.id = c.coupon_id and c.batch_num = 'P2504051322' and c.status != -1 LEFT JOIN car_order_photo_worknum d on a.mobile = d.mobile and d.company = 43 where a.coupon_batch = 'P2504051322' "

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "河南民生摆台订单", c)
}

func (*ExportExcel) Gdtk(c *gin.Context) {
	type Result struct {
		OrderNo      string `json:"order_no" tag:"订单编号"`
		Product      string `json:"product" tag:"产品名称"`
		Num          string `json:"num" tag:"购买数量"`
		Order_amount string `json:"order_amount" tag:"订单金额"`
		PayNo        string `json:"pay_no" tag:"支付单号"`
		PayAt        string `json:"pay_at" tag:"支付时间"`
		Mobile       string `json:"mobile" tag:"手机号"`
		Work_num     string `json:"work_num" tag:"业务员工号"`
		Organ        string `json:"organ" tag:"中支"`
		Name         string `json:"name" tag:"营业部"`
		Status       string `json:"status" tag:"订单状态"`
		C_time       string `json:"c_time" tag:"创建时间"`
	}

	var result []Result

	sqlQuery := "select a.order_no, case a.pro_id when 'TK001' then '元气参耳鲜炖宝' when 'TK002' then '景点旅游门票年卡' end product,a.num,a.order_amount,a.pay_no,if(a.pay_at,FROM_UNIXTIME(a.pay_at),'') as 'pay_at',case a.status when 0 then '未付款' when 1 then '已付款' when 2 then '已完成' when -1 then '已取消' end as 'status',b.name as 'name',b.mobile,b.contact,b.organ,b.work_num,FROM_UNIXTIME(a.c_time,'%Y-%m-%d %H:%i:%s') as 'c_time'  from car_order_gdpa a LEFT JOIN car_order_photo_agent b on (a.uid = b.uid and b.company = 44) where a.company = 4 "

	db := model.RDB[model.MASTER]
	err := db.Db.Raw(sqlQuery).Find(&result).Error
	if err != nil {
		c.String(200, "暂无订单数据！")
		return
	}
	if len(result) == 0 {
		c.String(200, "暂无订单数据！")
	}

	utils.Down(result, "广东泰康客养礼采购订单", c)
}

func (*ExportExcel) TkdgOrder(c *gin.Context) {
	type Result struct {
		Code       string `json:"code" tag:"优惠券包编号"`
		Name       string `json:"name" tag:"名称"`
		Sn         string `json:"sn" tag:"序列号"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		Organ      string `json:"organ" tag:"机构"`
		Work_num   string `json:"work_num" tag:"工号"`
		Phone      string `json:"phone" tag:"业务员手机"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		Remark     string `json:"remark" tag:"备注"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}
	var result []Result
	sqlQuery := `
	select b.code,b.name,b.sn,b.password,if(d.order_no<>'','已下单',if(c.remark<>'','已分享',if(b.status =0,'未激活','已激活'))) status,if(b.active_time,FROM_UNIXTIME(b.active_time, '%Y-%m-%d %H:%i:%s'),'') active_time,e.organ,e.work_num,b.mobile as phone,d.order_no,d.contact,d.mobile,concat(d.province,d.city,d.area,d.address) address,SUBSTRING_INDEX(REPLACE (d.customer_info,CONCAT(SUBSTRING_INDEX(d.customer_info, '"contact":', 1),'"contact":"'),''),'"', 1) as remark,d.ship_name,d.ship_no,if(d.c_time,FROM_UNIXTIME(d.c_time, '%Y-%m-%d %H:%i:%s'),'') c_time from car_coupon_pkg b  LEFT JOIN car_coupon c on b.id = c.pkg_id LEFT JOIN car_order_photo d on c.id = d.coupon_id and d.status <> -1 LEFT JOIN (select * from car_order_photo_agent where company = 2) e on(c.user_id = e.uid)  WHERE b.batch_num = 'PB250429469'
	`

	db := model.RDB[model.MASTER]
	err := db.Db.Raw(sqlQuery).Find(&result).Error
	if err != nil {
		c.String(200, "暂无订单数据！")
		return
	}
	if len(result) == 0 {
		c.String(200, "暂无订单数据！")
	}

	utils.Down(result, "泰康定格美好摆台订单", c)
}

func (*ExportExcel) FjrbsOrder(c *gin.Context) {
	type Result struct {
		Name       string `json:"name" tag:"业务员姓名"`
		Phone      string `json:"phone" tag:"业务员手机"`
		Work_num   string `json:"work_num" tag:"业务员工号"`
		Company    string `json:"company" tag:"机构名称"`
		Organ      string `json:"organ" tag:"四级机构"`
		Num        string `json:"num" tag:"匹配数量"`
		Sn         string `json:"sn" tag:"序列号"`
		Id         string `json:"id" tag:"卡券ID"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Pro_name   string `json:"pro_name" tag:"产品名称"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}
	var result []Result
	sqlQuery := `SELECT 
 c.name,
    c.work_num,
    c.organ,
    c.contact company,
    a.mobile phone,
		a.num as num11,
    CASE 
				when a.status = 0 then a.num
        WHEN b.id IS NOT NULL AND b.id <> '' THEN 1
				
        ELSE a.num - (
            select count(*) 
            from car_coupon bc 
            where bc.mobile = a.mobile 
            and bc.batch_num = 'P2505281009' 
        )
    END AS num,
    b.sn,
    b.id,
    CASE b.status 
        WHEN 0 THEN '未激活' 
        WHEN 1 THEN '已激活' 
        WHEN 2 THEN '已下单' 
        WHEN 3 THEN '已过期' 
    END status,
    IF(b.active_time,FROM_UNIXTIME(b.active_time, '%Y-%m-%d %H:%i:%s'),'') active_time ,
		 d.order_no,
    d.pro_name,
    d.contact,
    d.mobile,
    CONCAT(d.province,d.city,d.area,d.address) address,
    d.ship_name,
    d.ship_no,
    IF(d.c_time,FROM_UNIXTIME(d.c_time, '%Y-%m-%d %H:%i:%s'),'') c_time
 
FROM 
    (
        SELECT 
            mobile, 
            SUM(num) AS num, 
						uid,
            coupon_batch, 
            MAX(status) as status
        FROM car_member_bind_logs
        WHERE coupon_batch = 'P2505281009'
        GROUP BY mobile,status
    ) a
    LEFT JOIN car_coupon b ON a.mobile = b.mobile AND a.coupon_batch = b.batch_num and a.status =1 LEFT JOIN car_order_photo_agent c ON a.mobile = c.mobile AND c.company = 45 LEFT JOIN
		
		(
        SELECT pro_name,c_time,status,coupon_id,order_no,contact,mobile,province,city,area,address,ship_name,ship_no FROM car_order_photo WHERE batch_num = 'P2505281009'
        UNION ALL
        SELECT 
            CASE  
                WHEN pro_id IN(3,4,62,63,64,66)  THEN '定制马克杯' 
                WHEN pro_id = 61 THEN '定制手机壳' 
                WHEN pro_id = 26  THEN '定制帆布袋' 
            END pro_name,
            c_time,status,coupon_id,order_no,contact,mobile,province,city,area,address,ship_name,ship_no 
            FROM car_order_tshirt WHERE batch_num = 'P2505281009'
        UNION ALL
        SELECT '定制相册' pro_name,c_time,status,coupon_id,order_no,contact,mobile,province,city,area,address,ship_name,ship_no 
            FROM car_order_album WHERE batch_num = 'P2505281009'
    ) d ON b.id = d.coupon_id AND d.status <> -1
		
		 WHERE a.coupon_batch = 'P2505281009' 
	`

	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	name := "福建人保寿幸福御定数据"

	utils.Down(result, name, c)
}

func (*ExportExcel) Gsqy(c *gin.Context) {

	type Result struct {
		Name       string `json:"name" tag:"业务员姓名"`
		Phone      string `json:"phone" tag:"业务员手机"`
		Work_num   string `json:"work_num" tag:"业务员工号"`
		Company    string `json:"company" tag:"机构名称"`
		Organ      string `json:"organ" tag:"营业区"`
		Num        string `json:"num" tag:"匹配数量"`
		Sn         string `json:"sn" tag:"序列号"`
		Id         string `json:"id" tag:"卡券ID"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Pro_name   string `json:"pro_name" tag:"产品名称"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}
	var result []Result
	sqlQuery := `SELECT c.name, c.work_num, c.organ, c.contact company, a.mobile phone, a.num AS num11, CASE WHEN a.STATUS = 0 THEN a.num WHEN b.id IS NOT NULL AND b.id <> '' THEN 1 ELSE a.num - ( SELECT count(*) FROM car_coupon bc WHERE bc.mobile = a.mobile AND bc.batch_num = 'P2507041746' ) END AS num, b.sn, b.id, CASE b.STATUS WHEN 0 THEN '未激活' WHEN 1 THEN '已激活' WHEN 2 THEN '已下单' WHEN 3 THEN '已过期' END STATUS, IF ( b.active_time, FROM_UNIXTIME( b.active_time, '%Y-%m-%d %H:%i:%s' ), '' ) active_time, d.order_no, d.pro_name, d.contact, d.mobile, CONCAT( d.province, d.city, d.area, d.address ) address, d.ship_name, d.ship_no, IF ( d.c_time, FROM_UNIXTIME( d.c_time, '%Y-%m-%d %H:%i:%s' ), '' ) c_time FROM ( SELECT mobile, SUM( num ) AS num, uid, coupon_batch, MAX( STATUS ) AS STATUS FROM car_member_bind_logs WHERE coupon_batch = 'P2507041746' GROUP BY mobile, STATUS ) a LEFT JOIN car_coupon b ON a.mobile = b.mobile AND a.coupon_batch = b.batch_num AND a.STATUS = 1 LEFT JOIN car_order_photo_agent c ON a.mobile = c.mobile AND c.company = 48 LEFT JOIN ( SELECT pro_name, c_time, STATUS, coupon_id, order_no, contact, mobile, province, city, area, address, ship_name, ship_no FROM car_order_photo WHERE batch_num = 'P2507041746' ) d ON b.id = d.coupon_id AND d.STATUS <> - 1 WHERE a.coupon_batch = 'P2507041746'
	`
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "国寿清远摆台", c)

}

func (*ExportExcel) GsqyTotal(c *gin.Context) {

	type Result struct {
		Name      string `json:"name" tag:"业务员姓名"`
		Mobile    string `json:"mobile" tag:"业务员手机"`
		Work_num  string `json:"work_num" tag:"业务员工号"`
		Contact   string `json:"contact" tag:"机构名称"`
		Organ     string `json:"organ" tag:"营业区"`
		Activity  string `json:"activity" tag:"激活状态"`
		Total_num string `json:"total_num" tag:"匹配数量"`
		Order_num string `json:"order_num" tag:"订单数量"`
		Last_num  string `json:"last_num" tag:"剩余数量"`
	}
	var result []Result
	sqlQuery := `SELECT a.name, a.mobile, a.work_num, a.contact, a.organ, b.total_num AS total_num, CASE WHEN c.id IS NOT NULL THEN '已激活' ELSE '未激活' END AS activity, IFNULL(c.order_count, 0) AS order_num, GREATEST(IFNULL(b.total_num, 0) - IFNULL(c.order_count, 0), 0) AS last_num FROM car_order_photo_agent a LEFT JOIN ( SELECT mobile, SUM(num) AS total_num FROM car_member_bind_logs WHERE coupon_batch = 'P2507041746' GROUP BY mobile ) b ON a.mobile = b.mobile LEFT JOIN ( SELECT mobile, MAX(id) AS id, SUM(status = 2) AS order_count FROM car_coupon WHERE batch_num = 'P2507041746' GROUP BY mobile ) c ON a.mobile = c.mobile WHERE a.company = 48 GROUP BY a.mobile, a.name, a.organ, a.contact, b.total_num, c.id
	`
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)

	utils.Down(result, "国寿清远摆台代理人统计", c)

}

func (*ExportExcel) Xcgs(c *gin.Context) {
	typeVal := c.Query("type")
	type Result struct {
		Name       string `json:"name" tag:"业务员姓名"`
		Phone      string `json:"phone" tag:"业务员手机"`
		Work_num   string `json:"work_num" tag:"业务员工号"`
		Company    string `json:"company" tag:"机构名称"`
		Organ      string `json:"organ" tag:"营业区"`
		Num        string `json:"num" tag:"匹配数量"`
		Sn         string `json:"sn" tag:"序列号"`
		Id         string `json:"id" tag:"卡券ID"`
		Status     string `json:"status" tag:"状态"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Pro_name   string `json:"pro_name" tag:"产品名称"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}
	batchNum := "P2507240918"
	company := 50
	execName := "许昌国寿摆台"
	if typeVal == "1" {
		batchNum = "P2507151734"
		company = 51
		execName = "许昌长葛国寿摆台"
	}
	var result []Result

	sqlQuery := `SELECT c.name, c.work_num, c.organ, c.contact company, a.mobile phone, a.num AS num11, CASE WHEN a.STATUS = 0 THEN a.num WHEN b.id IS NOT NULL AND b.id <> '' THEN 1 ELSE a.num - ( SELECT count(*) FROM car_coupon bc WHERE bc.mobile = a.mobile AND bc.batch_num = ? ) END AS num, b.sn, b.id, CASE b.STATUS WHEN 0 THEN '未激活' WHEN 1 THEN '已激活' WHEN 2 THEN '已下单' WHEN 3 THEN '已过期' END STATUS, IF ( b.active_time, FROM_UNIXTIME( b.active_time, '%Y-%m-%d %H:%i:%s' ), '' ) active_time, d.order_no, d.pro_name, d.contact, d.mobile, CONCAT( d.province, d.city, d.area, d.address ) address, d.ship_name, d.ship_no, IF ( d.c_time, FROM_UNIXTIME( d.c_time, '%Y-%m-%d %H:%i:%s' ), '' ) c_time FROM ( SELECT mobile, SUM( num ) AS num, uid, coupon_batch, MAX( STATUS ) AS STATUS FROM car_member_bind_logs WHERE coupon_batch = ? GROUP BY mobile, STATUS ) a LEFT JOIN car_coupon b ON a.mobile = b.mobile AND a.coupon_batch = b.batch_num AND a.STATUS = 1 LEFT JOIN car_order_photo_agent c ON a.mobile = c.mobile AND c.company = ? LEFT JOIN ( SELECT pro_name, c_time, STATUS, coupon_id, order_no, contact, mobile, province, city, area, address, ship_name, ship_no FROM car_order_photo WHERE batch_num = ? ) d ON b.id = d.coupon_id AND d.STATUS <> - 1 WHERE a.coupon_batch = ?
	`
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery, batchNum, batchNum, company, batchNum, batchNum).Find(&result)

	utils.Down(result, execName, c)

}

func (e *ExportExcel) Whgs(c *gin.Context) {
	f, err := excelize.OpenFile("weihai.xlsx")
	if err != nil {
		e.handleError(c, "打开Excel文件失败", err)
		return
	}

	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		fmt.Errorf("no sheets found in excel file")
		return
	}
	sheetName := sheets[0]

	// 获取列名
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Errorf("failed to get rows: %v", err)
		return
	}
	var passwd []string
	var organs = make(map[string]string)
	for rowIdx, row := range rows {
		if rowIdx == 0 {
			continue
		}
		if len(row) > 4 && row[4] != "" {
        passwd = append(passwd, row[4])
        organs[row[4]] = ""
        if len(row) > 5 && row[5] != "" {
            organs[row[4]] = row[5]
        }
    }
	}
	type Result struct {
		Sn         string `json:"sn" tag:"卡号"`
		Password   string `json:"password" tag:"兑换码"`
		Phone      string `json:"phone" tag:"业务员手机"`
		Organ      string `json:"organ" tag:"机构"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Pro_name   string `json:"pro_name" tag:"产品名称"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
		Status     string `json:"status" tab:"订单状态"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}

	var result []Result

	sql := `select a.sn,a.password,b.mobile as phone,IF ( b.active_time, FROM_UNIXTIME( b.active_time, '%Y-%m-%d %H:%i:%s' ), '' ) active_time,c.order_no,c.organ,c.contact,c.pro_name,c.mobile,concat(c.province,c.city,c.area,c.address) as address,c.ship_name,c.ship_no, CASE b.STATUS WHEN 0 THEN '未激活' WHEN 1 THEN '已激活' WHEN 2 THEN '已下单' WHEN 3 THEN '已过期' END status, IF ( c.c_time, FROM_UNIXTIME( c.c_time, '%Y-%m-%d %H:%i:%s' ), '' ) c_time from car_coupon_pkg a LEFT JOIN car_coupon b on a.id = b.pkg_id and b.batch_num ='B250910688' LEFT JOIN car_order_photo c on b.id = c.coupon_id and c.status != -1 where a.password in(?)`
	db := model.RDB[model.MASTER]
	db.Db.Raw(sql, passwd).Find(&result)
	for k, v := range result {
		org := organs[v.Password]
		if org != "" {
			result[k].Organ = org
		}
	}

	utils.Down(result, "威海国寿", c)

}

func (e *ExportExcel) handleError(c *gin.Context, message string, err error) {
	slog.Error(message, err)
	c.String(http.StatusInternalServerError, message)
}


func (*ExportExcel) Xzgs(c *gin.Context) {
	at := c.Query("at")
	if at != "sfdjwie2ji239324" {
		c.String(200, "非法访问")
		return
	}

	type Result struct {
		Sn            string `json:"sn" tag:"卡券编号"`
		Password      string `json:"password" tag:"兑换码"`
		Status        string `json:"status" tag:"状态"`
		Active_time   string `json:"active_time" tag:"激活时间"`
		Order_no      string `json:"order_no" tag:"订单号"`
		Contact       string `json:"contact" tag:"联系人"`
		Mobile        string `json:"mobile" tag:"手机号"`
		Province      string `json:"province" tag:"省"`
		City          string `json:"city" tag:"市"`
		Area          string `json:"area" tag:"区"`
		Address       string `json:"address" tag:"地址"`
		Customer_info string `json:"customer_info" tag:"客户姓名"`
		Cus_mobile    string `json:"cus_mobile" tag:"客户手机"`
		Ship_name     string `json:"ship_name" tag:"快递公司"`
		Ship_no       string `json:"ship_no" tag:"快递单号"`
		C_time       string `json:"c_time" tag:"下单时间"`
		
	}

	var result []Result

	sqlQuery := "select IF ( b.active_time, FROM_UNIXTIME( b.active_time, '%Y-%m-%d %H:%i:%s' ), '' ) active_time,a.status,b.sn,b.password,c.order_no,c.contact,c.mobile,c.province,c.city,c.area,c.address,c.customer_info,c.ship_name,c.ship_no,c.organ,c.work_num,IF ( c.c_time, FROM_UNIXTIME( c.c_time, '%Y-%m-%d %H:%i:%s' ), '' ) c_time from car_coupon_pkg b left join  car_coupon a on a.pkg_id = b.id left join car_order_photo c on a.id = c.coupon_id and c.status != -1 where b.batch_num = 'PB2509291824'"
	db := model.RDB[model.MASTER]
	db.Db.Raw(sqlQuery).Find(&result)
	type Customer struct {
		Contact  string `json:"contact"`
		Work_num int    `json:"work_num"`
	}

	for k, v := range result {
		if v.Customer_info != "" {
			var tom Customer
			err := json.Unmarshal([]byte(v.Customer_info), &tom)
			if err == nil {
				result[k].Customer_info = tom.Contact
			}
		}

		status := "未激活"
		num, _ := strconv.Atoi(v.Status)
		if num == 1 {
			status = "已激活"
		}else if num == 2 {
			status = "已下单"
		}
		result[k].Status = status
		
	}

	utils.Down(result, "滁州国寿10寸摆台", c)
}


func (e *ExportExcel) Whgss(c *gin.Context) {
	
	type Result struct {
		Sn         string `json:"sn" tag:"卡号"`
		Password   string `json:"password" tag:"兑换码"`
		Phone      string `json:"phone" tag:"业务员手机"`
		Work_num      string `json:"work_num" tag:"业务员工号"`
		Agt_name      string `json:"agt_name" tag:"业务员姓名"`
		Organ      string `json:"organ" tag:"机构"`
		ActiveTime string `json:"active_time" tag:"激活时间"`
		OrderNo    string `json:"order_no" tag:"订单号"`
		Pro_name   string `json:"pro_name" tag:"产品名称"`
		Contact    string `json:"contact" tag:"收货人"`
		Mobile     string `json:"mobile" tag:"收货手机"`
		Address    string `json:"address" tag:"收货地址"`
		ShipName   string `json:"ship_name" tag:"快递公司"`
		ShipNo     string `json:"ship_no" tag:"快递单号"`
		Status     string `json:"status" tab:"订单状态"`
		C_time     string `json:"c_time" tag:"下单时间"`
	}

	var result []Result

	sql := `select a.sn,a.password,b.mobile as phone,IF ( b.active_time, FROM_UNIXTIME( b.active_time, '%Y-%m-%d %H:%i:%s' ), '' ) active_time,c.order_no,c.organ,c.contact,c.pro_name,SUBSTRING_INDEX(REPLACE (customer_info,CONCAT(SUBSTRING_INDEX(customer_info, '"contact":', 1),'"contact":"'),''),'"', 1) agt_name,c.work_num,c.mobile,concat(c.province,c.city,c.area,c.address) as address,c.ship_name,c.ship_no, CASE b.STATUS WHEN 0 THEN '未激活' WHEN 1 THEN '已激活' WHEN 2 THEN '已下单' WHEN 3 THEN '已过期' END status, IF ( c.c_time, FROM_UNIXTIME( c.c_time, '%Y-%m-%d %H:%i:%s' ), '' ) c_time from car_coupon_pkg a LEFT JOIN car_coupon b on a.id = b.pkg_id and b.batch_num ='P2510110950' LEFT JOIN car_order_photo c on b.id = c.coupon_id and c.status != -1 where a.batch_num = 'PP2510110950'`
	db := model.RDB[model.MASTER]
	db.Db.Raw(sql).Find(&result)
	

	utils.Down(result, "威海国寿", c)

}
