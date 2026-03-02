package models

import "time"

type Import struct {
	Id         int32     `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	OrderNo    string    `gorm:"column:order_no;default:;NOT NULL;comment:'订单编号'"`
	ShipName   string    `gorm:"column:ship_name;default:;NOT NULL;comment:'快递公司'"`
	ShipNo     string    `gorm:"column:ship_no;default:;NOT NULL;comment:'快递单号'"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:'创建时间'"`
	UpdatedAt  time.Time `gorm:"column:updated_at;comment:'更新时间'"`
}

func (i *Import) TableName() string {
	return "import"
}
