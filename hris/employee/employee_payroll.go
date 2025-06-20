package employee

import "github.com/AMETORY/ametory-erp-modules/shared/models"

type NonTaxableIncomeLevel struct {
	Code                         string  `json:"code"`
	Description                  string  `json:"description"`
	Amount                       float64 `json:"amount"`
	EffectiveRateAverageCategory string  `json:"effective_rate_average_category"`
}

func NewNonTaxableIncomeLevel() []NonTaxableIncomeLevel {
	return []NonTaxableIncomeLevel{
		{"-", "Non Pajak", 0, ""},
		{"TK/0", "Tidak Kawin Tanpa Tanggungan", 54000000, "A"},
		{"TK/1", "Tidak Kawin 1 Orang Tanggungan", 58500000, "A"},
		{"TK/2", "Tidak Kawin 2 Orang Tanggungan", 63000000, "B"},
		{"TK/3", "Tidak Kawin 3 Orang Tanggungan", 67500000, "B"},
		{"K/0", "Kawin Tanpa Tanggungan", 58500000, "A"},
		{"K/1", "Kawin 1 Orang Tanggungan", 63000000, "B"},
		{"K/2", "Kawin 2 Orang Tanggungan", 67500000, "B"},
		{"K/3", "Kawin 3 Orang Tanggungan", 72000000, "C"},
		{"K/1/0", "Kawin Penghasilan Istri Digabung Dengan Suami Tanpa Tanggungan", 112500000, "A"},
		{"K/1/1", "Kawin Penghasilan Istri Digabung Dengan Suami 1 Orang Tanggungan", 117000000, "A"},
		{"K/1/2", "Kawin Penghasilan Istri Digabung Dengan Suami 2 Orang Tanggungan", 121500000, "B"},
		{"K/1/3", "Kawin Penghasilan Istri Digabung Dengan Suami 3 Orang Tanggungan", 126000000, "C"},
	}
}

func (s EmployeeService) GetNonTaxableIncomeLevelAmount(m *models.EmployeeModel) float64 {
	for _, v := range NewNonTaxableIncomeLevel() {
		if v.Code == m.NonTaxableIncomeLevelCode {
			return v.Amount
		}
	}
	return 0
}
func (s EmployeeService) GetNonTaxableIncomeLevelCategory(m *models.EmployeeModel) string {
	for _, v := range NewNonTaxableIncomeLevel() {
		if v.Code == m.NonTaxableIncomeLevelCode {
			return v.EffectiveRateAverageCategory
		}
	}
	return ""
}

func (s EmployeeService) EffectiveRateAverageCategoryA(taxable float64) float64 {
	if taxable < 5400000 {
		return float64(0) / 100
	} else if taxable < 5650000 {
		return float64(0.25) / 100
	} else if taxable < 5950000 {
		return float64(0.5) / 100
	} else if taxable < 6300000 {
		return float64(0.75) / 100
	} else if taxable < 6750000 {
		return float64(1) / 100
	} else if taxable < 7500000 {
		return float64(1.25) / 100
	} else if taxable < 8550000 {
		return float64(1.5) / 100
	} else if taxable < 9650000 {
		return float64(1.75) / 100
	} else if taxable < 10050000 {
		return float64(2) / 100
	} else if taxable < 10350000 {
		return float64(2.25) / 100
	} else if taxable < 10700000 {
		return float64(2.5) / 100
	} else if taxable < 11050000 {
		return float64(3) / 100
	} else if taxable < 11600000 {
		return float64(3.5) / 100
	} else if taxable < 12500000 {
		return float64(4) / 100
	} else if taxable < 13750000 {
		return float64(5) / 100
	} else if taxable < 15100000 {
		return float64(6) / 100
	} else if taxable < 16950000 {
		return float64(7) / 100
	} else if taxable < 19750000 {
		return float64(8) / 100
	} else if taxable < 24150000 {
		return float64(9) / 100
	} else if taxable < 26450000 {
		return float64(10) / 100
	} else if taxable < 28000000 {
		return float64(11) / 100
	} else if taxable < 30050000 {
		return float64(12) / 100
	} else if taxable < 32400000 {
		return float64(13) / 100
	} else if taxable < 35400000 {
		return float64(14) / 100
	} else if taxable < 39100000 {
		return float64(15) / 100
	} else if taxable < 43850000 {
		return float64(16) / 100
	} else if taxable < 47800000 {
		return float64(17) / 100
	} else if taxable < 51400000 {
		return float64(18) / 100
	} else if taxable < 56300000 {
		return float64(19) / 100
	} else if taxable < 62200000 {
		return float64(20) / 100
	} else if taxable < 68600000 {
		return float64(21) / 100
	} else if taxable < 77500000 {
		return float64(22) / 100
	} else if taxable < 89000000 {
		return float64(23) / 100
	} else if taxable < 103000000 {
		return float64(24) / 100
	} else if taxable < 125000000 {
		return float64(25) / 100
	} else if taxable < 157000000 {
		return float64(26) / 100
	} else if taxable < 206000000 {
		return float64(27) / 100
	} else if taxable < 337000000 {
		return float64(28) / 100
	} else if taxable < 454000000 {
		return float64(29) / 100
	} else if taxable < 550000000 {
		return float64(30) / 100
	} else if taxable < 695000000 {
		return float64(31) / 100
	} else if taxable < 910000000 {
		return float64(32) / 100
	} else if taxable < 1400000000 {
		return float64(33) / 100
	} else {
		return float64(34) / 100
	}
}

func (s EmployeeService) EffectiveRateAverageCategoryB(taxable float64) float64 {
	if taxable < 6200000 {
		return float64(0) / 100
	} else if taxable < 6500000 {
		return float64(0.25) / 100
	} else if taxable < 6850000 {
		return float64(0.5) / 100
	} else if taxable < 7300000 {
		return float64(0.75) / 100
	} else if taxable < 9200000 {
		return float64(1) / 100
	} else if taxable < 10750000 {
		return float64(1.5) / 100
	} else if taxable < 11250000 {
		return float64(2) / 100
	} else if taxable < 11600000 {
		return float64(2.5) / 100
	} else if taxable < 12600000 {
		return float64(3) / 100
	} else if taxable < 13600000 {
		return float64(4) / 100
	} else if taxable < 14950000 {
		return float64(5) / 100
	} else if taxable < 16400000 {
		return float64(6) / 100
	} else if taxable < 18450000 {
		return float64(7) / 100
	} else if taxable < 21850000 {
		return float64(8) / 100
	} else if taxable < 26000000 {
		return float64(9) / 100
	} else if taxable < 27700000 {
		return float64(10) / 100
	} else if taxable < 29350000 {
		return float64(11) / 100
	} else if taxable < 31450000 {
		return float64(12) / 100
	} else if taxable < 33950000 {
		return float64(13) / 100
	} else if taxable < 37100000 {
		return float64(14) / 100
	} else if taxable < 41100000 {
		return float64(15) / 100
	} else if taxable < 45800000 {
		return float64(16) / 100
	} else if taxable < 49500000 {
		return float64(17) / 100
	} else if taxable < 53800000 {
		return float64(18) / 100
	} else if taxable < 58500000 {
		return float64(19) / 100
	} else if taxable < 64000000 {
		return float64(20) / 100
	} else if taxable < 71000000 {
		return float64(21) / 100
	} else if taxable < 80000000 {
		return float64(22) / 100
	} else if taxable < 93000000 {
		return float64(23) / 100
	} else if taxable < 109000000 {
		return float64(24) / 100
	} else if taxable < 129000000 {
		return float64(25) / 100
	} else if taxable < 163000000 {
		return float64(26) / 100
	} else if taxable < 211000000 {
		return float64(27) / 100
	} else if taxable < 374000000 {
		return float64(28) / 100
	} else if taxable < 459000000 {
		return float64(29) / 100
	} else if taxable < 555000000 {
		return float64(30) / 100
	} else if taxable < 704000000 {
		return float64(31) / 100
	} else if taxable < 957000000 {
		return float64(32) / 100
	} else if taxable < 1405000000 {
		return float64(33) / 100
	} else {
		return float64(34) / 100
	}
}

func (s EmployeeService) EffectiveRateAverageCategoryC(taxable float64) float64 {
	if taxable < 6600000 {
		return float64(0) / 100
	} else if taxable < 6950000 {
		return float64(0.25) / 100
	} else if taxable < 7350000 {
		return float64(0.5) / 100
	} else if taxable < 7800000 {
		return float64(0.75) / 100
	} else if taxable < 8850000 {
		return float64(1) / 100
	} else if taxable < 9800000 {
		return float64(1.25) / 100
	} else if taxable < 10950000 {
		return float64(1.5) / 100
	} else if taxable < 11200000 {
		return float64(1.75) / 100
	} else if taxable < 12050000 {
		return float64(2) / 100
	} else if taxable < 12950000 {
		return float64(3) / 100
	} else if taxable < 14150000 {
		return float64(4) / 100
	} else if taxable < 15550000 {
		return float64(5) / 100
	} else if taxable < 17050000 {
		return float64(6) / 100
	} else if taxable < 19500000 {
		return float64(7) / 100
	} else if taxable < 22700000 {
		return float64(8) / 100
	} else if taxable < 26600000 {
		return float64(9) / 100
	} else if taxable < 28100000 {
		return float64(10) / 100
	} else if taxable < 30100000 {
		return float64(11) / 100
	} else if taxable < 32600000 {
		return float64(12) / 100
	} else if taxable < 35400000 {
		return float64(13) / 100
	} else if taxable < 38900000 {
		return float64(14) / 100
	} else if taxable < 43000000 {
		return float64(15) / 100
	} else if taxable < 47400000 {
		return float64(16) / 100
	} else if taxable < 51200000 {
		return float64(17) / 100
	} else if taxable < 55800000 {
		return float64(18) / 100
	} else if taxable < 60400000 {
		return float64(19) / 100
	} else if taxable < 66700000 {
		return float64(20) / 100
	} else if taxable < 74500000 {
		return float64(21) / 100
	} else if taxable < 83200000 {
		return float64(22) / 100
	} else if taxable < 95000000 {
		return float64(23) / 100
	} else if taxable < 110000000 {
		return float64(24) / 100
	} else if taxable < 134000000 {
		return float64(25) / 100
	} else if taxable < 169000000 {
		return float64(26) / 100
	} else if taxable < 221000000 {
		return float64(27) / 100
	} else if taxable < 390000000 {
		return float64(28) / 100
	} else if taxable < 463000000 {
		return float64(39) / 100
	} else if taxable < 561000000 {
		return float64(30) / 100
	} else if taxable < 709000000 {
		return float64(31) / 100
	} else if taxable < 965000000 {
		return float64(32) / 100
	} else if taxable < 1419000000 {
		return float64(33) / 100
	} else {
		return float64(34) / 100
	}
}
