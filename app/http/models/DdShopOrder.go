package models

type DdShopOrder struct {
	ID        int64   `gorm:"primaryKey"`
	OrderNo   string  `gorm:"column:order_no;uniqueIndex"`	
	CouponID  *uint   `gorm:"column:coupon_id"`
	SkuID	 uint    `gorm:"column:sku_id"`
	SpuID	 uint    `gorm:"column:spu_id"`
	Amount    uint    `gorm:"column:amount"`
	Status    int     `gorm:"column:status"`
	UTime     int64   `gorm:"column:u_time"`
}

func (c *DdShopOrder) TableName() string {
	return "car_shop_dadi_order"
}

type DdShopProduct struct {
	ID        int64   `gorm:"primaryKey"`
	StockQty	 uint    `gorm:"column:stock_qty"`
}