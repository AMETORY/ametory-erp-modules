package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type DeductionSettingModel struct {
	shared.BaseModel
	Name                       string       `gorm:"size:50;not null" json:"name,omitempty"`
	Description                string       `gorm:"size:255" json:"description,omitempty"`
	CompanyID                  string       `json:"company_id" gorm:"not null"`
	Company                    CompanyModel `gorm:"foreignKey:CompanyID" json:"company"`
	DeductionType              string       `gorm:"size:50;not null" json:"deduction_type,omitempty"`            // Jenis potongan (e.g., "tax", "bpjs", "loan", "late", "not_presence")
	CalculationMethod          string       `gorm:"size:20;not null" json:"calculation_method,omitempty"`        // Metode perhitungan (e.g., "percentage", "fixed", "progressive")
	Amount                     float64      `gorm:"type:decimal(10,2)" json:"amount,omitempty"`                  // Nilai potongan (bisa persentase atau nominal)
	IsActive                   bool         `gorm:"default:true" json:"is_active,omitempty"`                     // Apakah setting ini aktif?
	IsDefault                  bool         `gorm:"default:false" json:"is_default,omitempty"`                   // Apakah setting ini default?
	Priority                   int          `json:"priority,omitempty"`                                          // Prioritas urutan perhitungan
	EffectiveDate              time.Time    `json:"effective_date,omitempty"`                                    // Tanggal mulai berlaku
	EndDate                    *time.Time   `json:"end_date,omitempty"`                                          // Tanggal berakhir (jika ada)
	ExcludePositionalAllowance bool         `gorm:"default:false" json:"exclude_positional_allowance,omitempty"` // ExcludePositionalAllowance: apakah potongan ini mengecualikan allowance posisi (jabatan)?
	ExcludeTransportAllowance  bool         `gorm:"default:false" json:"exclude_transport_allowance,omitempty"`  // ExcludeTransportAllowance: apakah potongan ini mengecualikan allowance transportasi?
	ExcludeMealAllowance       bool         `gorm:"default:false" json:"exclude_meal_allowance,omitempty"`       // ExcludeMealAllowance: apakah potongan ini mengecualikan allowance makan?
	TotalWorkingDays           float64      `gorm:"type:decimal(10,2)" json:"total_working_days,omitempty"`      // Total hari kerja dalam bulan
}
