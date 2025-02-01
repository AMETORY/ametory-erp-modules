package models

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentModel adalah model database untuk payment
type PaymentModel struct {
	shared.BaseModel
	Code                string      `gorm:"type:varchar(50);not null;uniqueIndex:idx_code,unique" json:"code"`
	Name                string      `gorm:"type:varchar(255);not null" json:"name"`
	Email               string      `gorm:"type:varchar(255);not null" json:"email"`
	Phone               string      `gorm:"type:varchar(50);not null" json:"phone"`
	Total               float64     `gorm:"type:decimal(10,2);not null" json:"total"`
	PaymentProvider     string      `gorm:"type:varchar(255);not null" json:"payment_provider"`
	PaymentLink         string      `gorm:"type:varchar(255);not null" json:"payment_link"`
	PaymentData         string      `gorm:"type:json" json:"-"`
	PaymentDataResponse interface{} `gorm:"-" json:"payment_data_response"`
	RefID               string      `gorm:"type:varchar(255);not null" json:"ref_id"`
	Status              string      `gorm:"type:varchar(50);default:PENDING;not null" json:"status"`
}

func (s *PaymentModel) TableName() string {
	return "payments"
}

func (pm *PaymentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if pm.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (pm *PaymentModel) AfterFind(tx *gorm.DB) (err error) {
	if pm.PaymentData != "" {
		var paymentData map[string]interface{}
		json.Unmarshal([]byte(pm.PaymentData), &paymentData)

		paymentMethod, ok := paymentData["payment_method"]
		if ok {
			if paymentMethod == "VA" && paymentData["sender_bank"] != nil {
				paymentData["bank"] = BankCodes[paymentData["sender_bank"].(string)]
			}
		}
		pm.PaymentDataResponse = paymentData
	}

	return
}

var BankCodes = map[string]string{
	"002": "Bank BRI",
	"008": "Bank Mandiri",
	"009": "Bank BNI",
	"200": "Bank BTN",
	"110": "Bank BJB",
	"111": "Bank DKI",
	"112": "Bank BPD D.I.Y",
	"113": "Bank Jateng",
	"114": "Bank Jatim",
	"115": "Bank Jambi",
	"116": "Bank Aceh",
	"117": "Bank Sumut",
	"118": "Bank Sumbar",
	"119": "Bank Kepri",
	"120": "Bank Sumsel dan Babel",
	"121": "Bank Lampung",
	"122": "Bank kalsel",
	"123": "Bank Kalbar",
	"124": "Bank Kaltim",
	"125": "Bank Kalteng",
	"126": "Bank Sulsel",
	"127": "Bank Sulut",
	"128": "Bank Ntb",
	"129": "Bank Bali",
	"130": "Bank Ntt",
	"131": "Bank Maluku",
	"132": "Bank Papua",
	"133": "Bank Bengkulu",
	"134": "Bank Sulteng",
	"135": "Bank Sultra",
	"137": "Bank Banten",
	"003": "Bank Ekspor Indonesia",
	"011": "Bank Danamon Indonesia",
	"013": "Bank Permata",
	"014": "Bank BCA",
	"016": "Bank Maybank",
	"019": "Bank Panin",
	"020": "Bank Arta Niaga Kencana",
	"022": "Bank CIMB Niaga",
	"023": "Bank UOB Indonesia",
	"026": "Bank Lippo",
	"028": "Bank OCBC NISP",
	"037": "Bank Artha Graha",
	"047": "Bank Pesona Perdania",
	"052": "Bank ABN Amro",
	"053": "Bank Keppel Tatlee Buana",
	"057": "Bank BNP Paribas Indonesia",
	"068": "Bank Woori Indonesia",
	"076": "Bank Bumi Arta",
	"087": "Bank Ekonomi",
	"089": "Bank Haga",
	"093": "Bank IFI",
	"095": "Bank Century/Bank J Trust Indonesia",
	"097": "Bank Mayapada",
	"145": "Bank Nusantara Parahyangan",
	"146": "Bank Swadesi",
	"151": "Bank Mestika",
	"157": "Bank Maspion",
	"159": "Bank Hagakita",
	"161": "Bank Ganesha",
	"162": "Bank Windu Kentjana",
	"164": "Bank ICBC Indonesia",
	"166": "Bank Harmoni Internasional",
	"167": "Bank QNB",
	"405": "Bank Swaguna",
	"426": "Bank Mega",
	"441": "Bank Bukopin",
	"459": "Bank Bisnis Internasional",
	"466": "Bank Sri Partha",
	"484": "Bank KEB Hana Indonesia",
	"485": "Bank MNC Internasional",
	"490": "Bank Neo",
	"494": "Bank BNI Agro",
	"503": "Bank Nobu",
	"513": "Bank Ina Perdana",
	"523": "Bank Sahabat Sampoerna",
	"535": "SeaBank",
	"542": "Bank Jago",
	"553": "Bank Mayora",
	"555": "Bank Index Selindo",
	"567": "AlloBank",
	"030": "Bank American Express Bank LTD",
	"032": "Bank JP. Morgan Chase Bank, N.A",
	"031": "Bank Citibank",
	"033": "Bank of America, N.A",
	"034": "Bank ING Indonesia Bank",
	"036": "Bank China Construction Bank Indonesia",
	"039": "Bank Credit Agricole Indosuez",
	"040": "Bank Bangkok",
	"042": "Bank of Tokyo Mitsubishi",
	"045": "Bank Sumitomo Mitsui Indonesia",
	"046": "Bank DBS Indonesia",
	"048": "Bank Mihuzo Indonesia",
	"050": "Bank Standard Chartered",
	"054": "Bank Capital Indonesia",
	"061": "Bank ANZ Indonesia",
	"069": "Bank of China Indonesia",
	"067": "Bank Deutsche",
	"152": "Bank Shinhan Indonesia",
	"212": "Bank Woori Saudara Indonesia",
	"950": "Bank Commonwealth",
	"147": "Bank Muamalat",
	"425": "Bank BJB Syariah",
	"451": "Bank Syariah Indonesia (BSI)",
	"506": "Bank Mega Syariah",
	"517": "Bank Panin Dubai Syariah",
	"521": "Bank Bukopin Syariah",
	"536": "Bank BCA Syariah",
	"547": "Bank BTPN Syariah",
	"947": "Bank Aladin Syariah",
}
