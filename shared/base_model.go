package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared/constants"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string         `gorm:"type:char(36);primary_key" bson:"-" json:"id,omitempty"`
	CreatedAt *time.Time     `json:"created_at" bson:"-"`
	UpdatedAt *time.Time     `json:"updated_at" bson:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-" bson:"-"`
}

type InvoiceBillSettingModel struct {
	ParentTemplate        string `json:"parent_template"`
	Template              string `json:"template"`
	DefaultFooter         string `json:"default_footer"`
	StaticCharacter       string `json:"static_character"`
	NumberFormat          string `json:"number_format"`
	AutoNumericLength     int    `json:"auto_numeric_length"`
	RandomNumericLength   int    `json:"random_numeric_length"`
	RandomCharacterLength int    `json:"random_character_length"`
	ShowFooter            bool   `json:"show_footer"`
	ShowSKU               bool   `json:"show_sku"`
	Title                 string `json:"title"`
	StoreID               string `json:"store_id"`
	SourceID              string `json:"source_id"`
	DestinationID         string `json:"destination_id"`
	DestinationCashID     string `json:"destination_cash_id"`
	TaxID                 string `json:"tax_id"`
	TaxMethod             string `json:"tax_method"`
	SecondaryTaxID        string `json:"secondary_tax_id"`
	SecondaryTaxMethod    string `json:"secondary_tax_method"`
	Notes                 string `json:"notes"`
	IsAutoStock           bool   `json:"is_auto_stock"`
	// HideTax               bool   `json:"hide_tax"`
	// HideSecondaryTax      bool   `json:"hide_secondary_tax"`
	// HideTotalBeforeTax    bool   `json:"hide_total_before_tax"`
}

func GenerateInvoiceBillNumber(data InvoiceBillSettingModel, before string) string {

	re := regexp.MustCompile(`{(.*?)}`)
	values := []any{}
	for _, v := range re.FindAllStringSubmatch(data.NumberFormat, -1) {
		if len(v) > 0 {
			if v[1] == constants.STATIC_CHARACTER {
				values = append(values, data.StaticCharacter)
			} else if v[1] == constants.AUTO_NUMERIC {

				numberBefore, err := strconv.Atoi(before)
				if err != nil {
					values = append(values, before)
				} else {
					if data.AutoNumericLength == 0 {
						values = append(values, fmt.Sprintf("%d", numberBefore+1))
					} else {
						length := strconv.Itoa(data.AutoNumericLength)

						values = append(values, fmt.Sprintf("%0"+length+"d", numberBefore+1))
					}

				}
			} else if v[1] == constants.MONTH_ROMAN {
				intMonth, _ := strconv.Atoi(time.Now().Format("1"))
				values = append(values, utils.IntegerToRoman(intMonth))
			} else if v[1] == constants.MONTH_MM {
				values = append(values, time.Now().Format("01"))
			} else if v[1] == constants.MONTH_MMM {
				values = append(values, time.Now().Format("Jan"))
			} else if v[1] == constants.MONTH_MMMM {
				values = append(values, time.Now().Format("January"))
			} else if v[1] == constants.YEAR_YY {
				values = append(values, time.Now().Format("06"))
			} else if v[1] == constants.YEAR_YYYY {
				values = append(values, time.Now().Format("2006"))
			} else if v[1] == constants.RANDOM_NUMERIC {
				values = append(values, utils.GenerateRandomNumber(data.RandomNumericLength))
			} else if v[1] == constants.RANDOM_CHARACTER {
				values = append(values, strings.ToUpper(utils.GenerateRandomString(data.RandomCharacterLength)))
			} else {
				values = append(values, v[0])
			}
		}
	}

	return fmt.Sprintf(re.ReplaceAllString(data.NumberFormat, "%s"), values...)
}

func ExtractNumber(data InvoiceBillSettingModel, number string) string {
	re2 := regexp.MustCompile(`{(.*?)}`)
	pattern := "0.{1}"
	if data.AutoNumericLength > 0 {
		pattern = fmt.Sprintf("(\\d{%d})", data.AutoNumericLength)

	}
	re := regexp.MustCompile(pattern)
	getNumber := re.FindAllString(number, -1)
	if len(getNumber) == 1 {
		num, _ := strconv.Atoi(getNumber[0])
		if num > 0 {
			return GenerateInvoiceBillNumber(data, getNumber[0])
		}
	}
	values := []any{}
	for _, v := range re2.FindAllStringSubmatch(data.NumberFormat, -1) {
		if len(v) > 0 {

			if v[1] == constants.STATIC_CHARACTER {
				values = append(values, data.StaticCharacter)
			} else if v[1] == constants.AUTO_NUMERIC {
				values = append(values, "(\\d+)")
			} else if v[1] == constants.MONTH_ROMAN {
				intMonth, _ := strconv.Atoi(time.Now().Format("1"))
				values = append(values, utils.IntegerToRoman(intMonth))
			} else if v[1] == constants.MONTH_MM {
				values = append(values, time.Now().Format("01"))
			} else if v[1] == constants.MONTH_MMM {
				values = append(values, time.Now().Format("Jan"))
			} else if v[1] == constants.MONTH_MMMM {
				values = append(values, time.Now().Format("January"))
			} else if v[1] == constants.YEAR_YY {
				values = append(values, time.Now().Format("06"))
			} else if v[1] == constants.YEAR_YYYY {
				values = append(values, time.Now().Format("2006"))
			} else if v[1] == constants.RANDOM_NUMERIC {
				values = append(values, utils.GenerateRandomNumber(data.RandomNumericLength))
			} else if v[1] == constants.RANDOM_CHARACTER {
				values = append(values, strings.ToUpper(utils.GenerateRandomString(data.RandomCharacterLength)))
			} else {
				values = append(values, v[0])
			}
		}
	}
	pattern2 := strings.ReplaceAll(fmt.Sprintf(re2.ReplaceAllString(data.NumberFormat, "%s"), values...), "/", "\\/")
	re3 := regexp.MustCompile(pattern2)

	for _, v := range re3.FindAllStringSubmatch(number, -1) {
		if len(v) > 0 {
			return GenerateInvoiceBillNumber(data, v[1])
		}
	}

	return GenerateInvoiceBillNumber(data, "00")
}
