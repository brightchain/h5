package handler

import (
	"errors"
	"log"
	"regexp"
	
)

func HandleOrderCancel(orderNo string) error {
	if orderNo == "" {
		return errors.New("缺少订单号")
	}
	re := regexp.MustCompile(`^[A-Za-z]+`)
	prefix := re.FindString(orderNo)

	switch prefix {
	case "ZY":
		return cancelZY(orderNo)
	case "DD", "DA":
		return cancelDD(orderNo)
	default:
		return errors.New("unsupported order prefix")
	}
}

func cancelZY(orderNo string) error {
	log.Println("Cancel ZY order:", orderNo)
	// TODO:
	// 1. 查订单
	// 2. 判断状态
	// 3. 更新状态
	// 4. 回滚库存
	// 5. 退优惠券
	return nil
}

func cancelDD(orderNo string) error {
	log.Println("Cancel DD order:", orderNo)
	return nil
}