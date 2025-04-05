package models

type COGSReport struct {
	GeneralReport
	BeginningInventory     float64      `gorm:"column:beginning_inventory" json:"beginning_inventory"`
	Purchases              float64      `gorm:"column:purchases" json:"purchases"`
	FreightInAndOtherCost  float64      `gorm:"column:freight_in_and_other_cost" json:"freight_in_and_other_cost"`
	TotalPurchases         float64      `gorm:"column:total_purchases" json:"total_purchases"`
	PurchaseReturns        float64      `gorm:"column:purchase_returns" json:"purchase_returns"`
	PurchaseDiscounts      float64      `gorm:"column:purchase_discounts" json:"purchase_discounts"`
	TotalPurchaseDiscounts float64      `gorm:"column:total_purchase_discounts" json:"total_purchase_discounts"`
	NetPurchases           float64      `gorm:"column:net_purchases" json:"net_purchases"`
	GoodsAvailable         float64      `gorm:"column:goods_available" json:"goods_available"`
	EndingInventory        float64      `gorm:"column:ending_inventory" json:"ending_inventory"`
	COGS                   float64      `gorm:"column:cogs" json:"cogs"`
	InventoryAccount       AccountModel `gorm:"-" json:"inventory_account"`
}
