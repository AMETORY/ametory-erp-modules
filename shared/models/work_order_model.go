package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type WorkOrder struct {
	shared.BaseModel
	Code                string              `gorm:"unique;not null" json:"code,omitempty"` // contoh: WO-PRD-202504
	ProductID           string              `gorm:"not null" json:"product_id,omitempty"`  // barang jadi yang akan diproduksi
	Product             *ProductModel       `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	QuantityPlanned     int                 `gorm:"not null" json:"quantity_planned,omitempty"`
	QuantityDone        int                 `json:"quantity_done,omitempty"`
	BomID               string              `gorm:"not null" json:"bom_id,omitempty"`
	BOM                 BillOfMaterial      `gorm:"foreignKey:BomID;constraint:OnDelete:CASCADE"`
	Status              string              `gorm:"default:DRAFT" json:"status,omitempty"` // DRAFT, RELEASED, IN_PROGRESS, DONE, CANCELLED
	ScheduledDate       time.Time           `json:"scheduled_date,omitempty"`
	StartedAt           *time.Time          `json:"started_at,omitempty"`
	FinishedAt          *time.Time          `json:"finished_at,omitempty"`
	Notes               string              `json:"notes,omitempty"`
	ProductionProcesses []ProductionProcess `gorm:"foreignKey:WorkOrderID" json:"production_processes,omitempty"`
}

type ProductionProcess struct {
	shared.BaseModel
	WorkOrderID      string                     `gorm:"not null" json:"work_order_id,omitempty"`
	WorkOrder        WorkOrder                  `json:"work_order,omitempty"`
	WorkCenterID     uint                       `gorm:"not null" json:"work_center_id,omitempty"`
	WorkCenter       WorkCenter                 `json:"work_center,omitempty"`
	BOM              BillOfMaterial             `json:"bom,omitempty"`
	ProcessName      string                     `json:"process_name,omitempty"`               // misal: Cutting, Welding, Packing
	Sequence         int                        `json:"sequence,omitempty" gorm:"default:1"`  // urutan dalam proses produksi
	ProductID        string                     `gorm:"not null" json:"product_id,omitempty"` // produk utama
	Product          ProductModel               `json:"product,omitempty"`
	QuantityPlanned  float64                    `gorm:"not null" json:"quantity_planned,omitempty"` // jumlah yang harus diproduksi di tahap ini
	QuantityDone     float64                    `json:"quantity_done,omitempty"`                    // hasil produksi aktual
	QuantityScrap    float64                    `json:"quantity_scrap,omitempty"`                   // hasil yang rusak/tidak layak
	UnitOfMeasure    string                     `json:"unit_of_measure,omitempty"`                  // misal: "pcs", "kg"
	StartedAt        *time.Time                 `json:"started_at,omitempty"`
	FinishedAt       *time.Time                 `json:"finished_at,omitempty"`
	AssignedToUserID *string                    `json:"assigned_to_user_id,omitempty"` // user/operator yang bertanggung jawab
	AssignedTo       *UserModel                 `json:"assigned_to,omitempty"`
	Status           string                     `json:"status,omitempty" gorm:"default:DRAFT"`                            // DRAFT, WAITING, IN_PROGRESS, DONE, FAILED, CANCELLED
	MaterialCost     float64                    `json:"material_cost,omitempty"`                                          // otomatis dari BOM (per unit * qty)
	OtherCost        float64                    `json:"other_cost,omitempty"`                                             // dari additional cost
	TotalCost        float64                    `json:"total_cost,omitempty"`                                             // MaterialCost + OtherCost
	Notes            string                     `json:"notes,omitempty"`                                                  // catatan umum
	QCNotes          string                     `json:"qc_notes,omitempty"`                                               // catatan quality control
	IsRework         bool                       `json:"is_rework,omitempty"`                                              // jika proses ini adalah hasil dari rework
	AdditionalCosts  []ProductionAdditionalCost `gorm:"foreignKey:ProductionProcessID" json:"additional_costs,omitempty"` // Relasi ke Additional Costs
	Outputs          []ProductionOutput         `gorm:"foreignKey:ProductionProcessID" json:"outputs,omitempty"`          // Relasi ke Output (jika multi-output atau output tambahan)
}

type ProductionAdditionalCost struct {
	shared.BaseModel
	ProductionProcessID string            `json:"production_process_id"`
	ProductionProcess   ProductionProcess `gorm:"foreignKey:ProductionProcessID" json:"production_process"`
	Type                string            `json:"type"` // contoh: "Labor", "Electricity", "Maintenance"
	Description         string            `json:"description"`
	Amount              float64           `json:"amount"`
}

type ProductionOutput struct {
	shared.BaseModel
	ProductionProcessID string            `json:"production_process_id"`
	ProductionProcess   ProductionProcess `gorm:"foreignKey:ProductionProcessID" json:"production_process"`
	ProductID           string            `json:"product_id"`
	Product             ProductModel      `gorm:"foreignKey:ProductID" json:"product"`
	Quantity            float64           `json:"quantity"`
	UnitID              string            `json:"unit_id"`
	Unit                UnitModel         `gorm:"foreignKey:UnitID;references:ID" json:"unit"`
	IsPrimaryOutput     bool              `json:"is_primary_output"`
	Notes               string            `json:"notes"`
}

type WorkCenter struct {
	shared.BaseModel
	Code        string  `gorm:"unique;not null" json:"code"` // contoh: "WC-MESIN-01"
	Name        string  `gorm:"not null" json:"name"`        // contoh: "Mesin Milling 01"
	Description string  `json:"description,omitempty"`
	Location    string  `json:"location,omitempty"` // misal: "Lantai 1 - Workshop A"
	Capacity    float64 `json:"capacity,omitempty"` // jumlah unit yang bisa diproses sekaligus
}
