package thirdparty

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

// Bpjs menghitung iuran BPJS berdasarkan peraturan perundangan yang berlaku
//
// Bpjs menghitung iuran BPJS Kesehatan, BPJS Ketenagakerjaan JHT, BPJS JP,
// BPJS JKM, dan BPJS JKK berdasarkan upah dan resiko pekerjaan.
//
// Bpjs menggunakan batas atas dan bawah upah yang ditentukan dalam peraturan
// perundangan yang berlaku untuk menghitung iuran BPJS.
//
// Bpjs juga menggunakan tarif iuran BPJS yang ditentukan oleh pemerintah untuk
// menghitung iuran BPJS.
//
// Bpjs diinisialisasi dengan menggunakan fungsi InitBPJS.
//
// Contoh penggunaan Bpjs:
// bpjs := InitBPJS()
// employerContribution, employeeContribution, totalContribution := bpjs.CalculateBPJSKes(salary)

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

	return employerContribution, employeeContribution, totalContribution
}

// CalculateBPJSTkJht menghitung iuran BPJS Ketenagakerjaan JHT berdasarkan upah yang diberikan.
//
// Fungsi ini akan memastikan bahwa upah berada dalam batas yang ditentukan untuk BPJS Ketenagakerjaan.
//
// Contoh penggunaan CalculateBPJSTkJht:
// bpjs := InitBPJS()
// employerContribution, employeeContribution, totalContribution := bpjs.CalculateBPJSTkJht(salary)
func (m Bpjs) CalculateBPJSTkJht(salary float64) (float64, float64, float64) {
	// Pastikan gaji berada dalam batas yang ditentukan untuk BPJS Ketenagakerjaan
	if salary > m.MaxSalaryTk {
		salary = m.MaxSalaryTk
	}

	// Hitung iuran BPJS Ketenagakerjaan JHT
	employerContribution := salary * m.BpjsTkJhtRateEmployer
	employeeContribution := salary * m.BpjsTkJhtRateEmployee
	totalContribution := employerContribution + employeeContribution

	return employerContribution, employeeContribution, totalContribution
}

// CalculateBPJSTkJp menghitung iuran BPJS Ketenagakerjaan Jp berdasarkan upah yang diberikan.
//
// Fungsi ini akan menghitung iuran BPJS Ketenagakerjaan Jp yang dibayar oleh pemberi kerja dan
// pekerja, serta total iuran yang harus dibayar.
//
// Contoh penggunaan CalculateBPJSTkJp:
// bpjs := InitBPJS()
// employerContribution, employeeContribution, totalContribution := bpjs.CalculateBPJSTkJp(salary)
func (m Bpjs) CalculateBPJSTkJp(salary float64) (float64, float64, float64) {

	// Hitung iuran BPJS Ketenagakerjaan Jp
	employerContribution := salary * m.BpjsTkJpRateEmployer
	employeeContribution := salary * m.BpjsTkJpRateEmployee
	totalContribution := employerContribution + employeeContribution

	return employerContribution, employeeContribution, totalContribution
}

// CalculateBPJSTkJkm menghitung iuran BPJS Ketenagakerjaan Kmk berdasarkan upah yang diberikan.
//
// Fungsi ini akan menghitung iuran BPJS Ketenagakerjaan Kmk yang dibayar oleh pekerja.
//
// Contoh penggunaan CalculateBPJSTkJkm:
// bpjs := InitBPJS()
// employerContribution := bpjs.CalculateBPJSTkJkm(salary)
func (m Bpjs) CalculateBPJSTkJkm(salary float64) float64 {
	return salary * m.BpjsTkJkmEmployee
}

// CalculateBPJSTkJkk menghitung iuran BPJS Ketenagakerjaan Kkk berdasarkan upah yang diberikan dan
// resiko yang dihadapi.
//
// Fungsi ini akan menghitung iuran BPJS Ketenagakerjaan Kkk yang dibayar oleh pekerja berdasarkan
// resiko yang dihadapi.
//
// Contoh penggunaan CalculateBPJSTkJkk:
// bpjs := InitBPJS()
// employerContribution, employeeContribution := bpjs.CalculateBPJSTkJkk(salary, risk)
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
