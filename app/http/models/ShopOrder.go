package models

type ShopOrder struct {
	ID        int64   `gorm:"primaryKey"`
	OrderNo   string  `gorm:"column:order_no;uniqueIndex"`
	Status    int     `gorm:"column:status"`
	PayAmount float64 `gorm:"column:pay_amount"`
	UTime     int64   `gorm:"column:u_time"`
	Items     []ShopOrderItem `gorm:"foreignKey:OrderNo;references:OrderNo"`
}

func (c *ShopOrder) TableName() string {
	return "car_shop_order"
}


type ShopOrderItem struct {
	ID     uint `gorm:"primaryKey"`
	OrderNo string `gorm:"column:order_no;index"`
	SkuID  uint 	`gorm:"column:sku_id"`
	Amount int   `gorm:"column:amount"`
	CouponID *uint 	`gorm:"column:coupon_id"`
}

func (c *ShopOrderItem) TableName() string {
	return "car_shop_order_item"
}

type ProStandard struct {
	ID       uint `gorm:"primaryKey"`
	StockQty int
}

func (c *ProStandard) TableName() string {
	return "car_shop_pro_std"
}

type Coupon struct {
	ID     uint `gorm:"primaryKey"`
	Status int
}

func (c *Coupon) TableName() string {
	return "car_coupon"
}

type ShopCoupon struct {
	ID     uint `gorm:"primaryKey"`
	OrderNo string `gorm:"column:order_no;index"`
	Status int `gorm:"column:status"`
}

func (c *ShopCoupon) TableName() string {
	return "car_shop_coupon"
}
