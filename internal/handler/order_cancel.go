package handler

import (
	"errors"
	"log"
	"regexp"
	"time"

	"h5/app/http/models"
	"h5/pkg/logger"
	"h5/pkg/model"

	"gorm.io/gorm"
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
	
	if orderNo == "" {
		logger.LogError("取消自营订单订单: "+orderNo, nil)
		return errors.New("订单号不能为空")
	}

	db := model.RDB[model.MASTER]
	if db == nil || db.Db == nil {
		log.Println("database not initialized")
		return errors.New("database not initialized")
	}

	return db.Db.Transaction(func(tx *gorm.DB) error {
		var order models.ShopOrder

		// 一定要 Preload Items
		if err := tx.
			Preload("Items").
			Where("order_no = ?", orderNo).
			First(&order).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.LogError("订单不存在: "+orderNo, nil)
				return nil
			}
			return err
		}

		// 幂等
		if order.Status != 0 {
			logger.LogError("订单状态不允许取消: "+orderNo, nil)
			return nil
		}

		// 更新订单状态
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status": 7,
			"u_time": time.Now().Unix(),
		}).Error; err != nil {
			return err
		}

		// 恢复库存
		for _, item := range order.Items {
			if err := tx.Exec(`
				UPDATE car_shop_pro_std
				SET stock_qty = stock_qty + ?
				WHERE id = ?
			`, item.Amount, item.SkuID).Error; err != nil {
				return err
			}
			logger.LogError("恢复库存: "+orderNo, nil)
		}

		// 恢复商品优惠券
		for _, item := range order.Items {
			if item.CouponID != nil && *item.CouponID > 0  {
				if err := tx.Model(&models.Coupon{}).
					Where("id = ? AND status = 2", *item.CouponID).
					Update("status", 1).Error; err != nil {
					return err
				}
			}
		}

		// 恢复订单优惠券
		if order.OrderNo != "" {
			if err := tx.Model(&models.ShopCoupon{}).
				Where("order_no = ? AND status = 2", order.OrderNo).
				Update("status", 1).Error; err != nil {
				logger.LogError("恢复商城优惠券 order_no="+order.OrderNo, nil)
			}
		}

		return nil // 自动 commit
	})
}


func cancelDD(orderNo string) error {
	if orderNo == "" {
		logger.LogError("取消大地订单: "+orderNo, nil)
		return errors.New("订单号不能为空")
	}

	db := model.RDB[model.MASTER]
	if db == nil || db.Db == nil {
		log.Println("database not initialized")
		return errors.New("database not initialized")
	}
	return db.Db.Transaction(func(tx *gorm.DB) error {
		var order models.DdShopOrder

		// 一定要 Preload Items
		if err := tx.Where("order_no = ?", orderNo).
			First(&order).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.LogError("订单不存在: "+orderNo, nil)
				return nil
			}
			return err
		}

		// 幂等
		if order.Status != 0 {
			logger.LogError("订单状态不允许取消: "+orderNo, nil)
			return nil
		}

		// 更新订单状态
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status": 7,
			"u_time": time.Now().Unix(),
		}).Error; err != nil {
			return err
		}

		// 恢复库存
		if err := tx.Exec(`
				UPDATE car_shop_dianli_goods
				SET stock_qty = stock_qty + ?
				WHERE id = ?
			`, order.Amount, order.SpuID).Error; err != nil {
				return err
			}
			logger.LogError("恢复库存: "+orderNo, nil)

		// 恢复商品优惠券
		if order.CouponID != nil && *order.CouponID > 0  {
			if err := tx.Model(&models.Coupon{}).
				Where("id = ? AND status = 2", *order.CouponID).
				Update("status", 1).Error; err != nil {
				return err
			}
		}
		
		return nil // 自动 commit
	})
}