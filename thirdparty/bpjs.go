package thirdparty

import "fmt"

// Konstanta untuk perhitungan BPJS dan PPh 21

type Bpjs struct {
	BpjsKesRateEmployer           float64 // 4% BPJS Kesehatan dibayar oleh pemberi kerja
	BpjsKesRateEmployee           float64 // 1% BPJS Kesehatan dibayar oleh pekerja
	BpjsTkJhtRateEmployer         float64 // 3.7% Jaminan Hari Tua  dibayar oleh pemberi kerja
	BpjsTkJhtRateEmployee         float64 // 2% Jaminan Hari Tua dibayar oleh pekerja
	MaxSalaryKes                  float64 // Batas atas gaji untuk BPJS Kesehatan
	MinSalaryKes                  float64 // Batas bawah gaji untuk BPJS Kesehatan
	MaxSalaryTk                   float64 // Batas atas gaji untuk BPJS Ketenagakerjaan
	BpjsTkJkkVeryLowRiskEmployee  float64 // BPJS JKK resiko sangat rendah
	BpjsTkJkkLowRiskEmployee      float64 // BPJS JKK resiko rendah
	BpjsTkJkkMiddleRiskEmployee   float64 // BPJS JKK resiko menengah
	BpjsTkJkkHighRiskEmployee     float64 // BPJS JKK resiko tinggi
	BpjsTkJkkVeryHighRiskEmployee float64 // BPJS JKK resiko sangat tinggi
	BpjsTkJkmEmployee             float64 // BPJS JKM
	BpjsTkJpRateEmployer          float64 // 2% Jaminan Pensiun  dibayar oleh pemberi kerja
	BpjsTkJpRateEmployee          float64 // 1% Jaminan Pensiun dibayar oleh pekerja
	BpjsKesEnabled                bool
	BpjsTkJhtEnabled              bool
	BpjsTkJkmEnabled              bool
	BpjsTkJpEnabled               bool
	BpjsTkJkkEnabled              bool
}

func InitBPJS() *Bpjs {
	return &Bpjs{
		BpjsKesRateEmployer:           0.04,
		BpjsKesRateEmployee:           0.01,
		BpjsTkJhtRateEmployer:         0.037,
		BpjsTkJhtRateEmployee:         0.02,
		MaxSalaryKes:                  12000000,
		MinSalaryKes:                  2000000,
		MaxSalaryTk:                   8800000,
		BpjsTkJkmEmployee:             0.003,
		BpjsTkJkkVeryLowRiskEmployee:  0.0024,
		BpjsTkJkkLowRiskEmployee:      0.0054,
		BpjsTkJkkMiddleRiskEmployee:   0.0089,
		BpjsTkJkkHighRiskEmployee:     0.0127,
		BpjsTkJkkVeryHighRiskEmployee: 0.0174,
	}
}

func (m Bpjs) CalculateBPJSKes(salary float64) (float64, float64, float64) {
	// Pastikan gaji berada dalam batas yang ditentukan untuk BPJS Kesehatan
	if salary > m.MaxSalaryKes {
		salary = m.MaxSalaryKes
	} else if salary < m.MinSalaryKes {
		salary = m.MinSalaryKes
	}

	// Hitung iuran BPJS Kesehatan
	employerContribution := salary * m.BpjsKesRateEmployer
	employeeContribution := salary * m.BpjsKesRateEmployee
	totalContribution := employerContribution + employeeContribution

	fmt.Println("CALCULATE_BPJSKES", salary, m.BpjsKesRateEmployer)
	fmt.Println("CALCULATE_BPJSKES", salary, m.BpjsKesRateEmployee)
	fmt.Println("CALCULATE_BPJSKES", employerContribution, employeeContribution)

	return employerContribution, employeeContribution, totalContribution
}

// Fungsi untuk menghitung iuran BPJS JHT Ketenagakerjaan
func (m Bpjs) CalculateBPJSTkJht(salary float64) (float64, float64, float64) {
	// Pastikan gaji berada dalam batas yang ditentukan untuk BPJS Ketenagakerjaan
	if salary > m.MaxSalaryTk {
		salary = m.MaxSalaryTk
	}

	// Hitung iuran BPJS Ketenagakerjaan JHT
	employerContribution := salary * m.BpjsTkJhtRateEmployer
	employeeContribution := salary * m.BpjsTkJhtRateEmployee
	totalContribution := employerContribution + employeeContribution

	fmt.Println("CALCULATE_BPJSTK_JHT", salary, m.BpjsTkJhtRateEmployer)
	fmt.Println("CALCULATE_BPJSTK_JHT", salary, m.BpjsTkJhtRateEmployee)
	fmt.Println("CALCULATE_BPJSTK_JHT", employerContribution, employeeContribution)

	return employerContribution, employeeContribution, totalContribution
}

// Fungsi untuk menghitung iuran BPJS JP Ketenagakerjaan
func (m Bpjs) CalculateBPJSTkJp(salary float64) (float64, float64, float64) {

	// Hitung iuran BPJS Ketenagakerjaan Jp
	employerContribution := salary * m.BpjsTkJpRateEmployer
	employeeContribution := salary * m.BpjsTkJpRateEmployee
	totalContribution := employerContribution + employeeContribution

	return employerContribution, employeeContribution, totalContribution
}

func (m Bpjs) CalculateBPJSTkJkm(salary float64) float64 {
	return salary * m.BpjsTkJkmEmployee
}

func (m Bpjs) CalculateBPJSTkJkk(salary float64, risk string) (float64, float64) {
	switch risk {
	case "low":
		return salary * m.BpjsTkJkkLowRiskEmployee, m.BpjsTkJkkLowRiskEmployee
	case "middle":
		return salary * m.BpjsTkJkkMiddleRiskEmployee, m.BpjsTkJkkMiddleRiskEmployee
	case "high":
		return salary * m.BpjsTkJkkHighRiskEmployee, m.BpjsTkJkkHighRiskEmployee
	case "very_high":
		return salary * m.BpjsTkJkkVeryHighRiskEmployee, m.BpjsTkJkkVeryHighRiskEmployee
	default:
		return salary * m.BpjsTkJkkVeryLowRiskEmployee, m.BpjsTkJkkVeryLowRiskEmployee
	}
}
