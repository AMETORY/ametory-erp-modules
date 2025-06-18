package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"math/rand"
	mathRand "math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/dgrijalva/jwt-go"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/ttacon/libphonenumber"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {

}

func GenerateRandomString(length int) string {
	return stringWithCharset(length, charset)
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const charsetNumber = "0123456789"

var seededRand *mathRand.Rand = mathRand.New(
	mathRand.NewSource(time.Now().UnixNano()))

// RandString generates a random string of length n
func RandString(n int, uppercase bool) string {
	r := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(r)
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	if uppercase {
		return strings.ToUpper(string(b))
	}
	return string(b)
}

// RandomStringNumber generates a random string of length n including numbers
func RandomStringNumber(n int, uppercase bool) string {
	r := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(r)
	letters := []rune("0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	if uppercase {
		return strings.ToUpper(string(b))
	}
	return string(b)
}

// CreateUsernameFromFullName creates a username from a full name
func CreateUsernameFromFullName(fullName string) string {
	letters := []rune(fullName)
	username := make([]rune, 0)
	for _, letter := range letters {
		if letter == ' ' {
			continue
		}
		username = append(username, letter)
	}
	if len(username) > 25 {
		username = username[:25]
	}
	username = append(username, '-')
	username = append(username, []rune(RandString(5, false))...)
	return string(username)
}

func LogJson(data interface{}) {
	jsonString, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonString))
}
func SaveJson(data interface{}) {
	jsonString, _ := json.MarshalIndent(data, "", "  ")
	log.Println(string(jsonString))
}

func FixRequest(request *http.Request) {
	req := request.URL.Query()
	pageStr := request.URL.Query().Get("page")
	if pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		// request.URL.Query().Set("page", strconv.Itoa(page-1))
		req.Set("page", strconv.Itoa(page-1))
		request.URL.RawQuery = req.Encode()
	}

}

func Uuid() string {
	return uuid.New().String()
}

func CalculateIncludeTax(subtotalInclTax float64, taxRate float64) float64 {
	// Rumus: pajak = (subtotal_incl_tax * tax_rate) / (1 + tax_rate)
	// jika tax_rate dalam persen maka konversi ke pecahan terlebih dahulu
	taxRateFraction := taxRate / 100
	tax := (subtotalInclTax * taxRateFraction) / (1 + taxRateFraction)
	return tax
}

func CalculateExcludeTax(subtotalExclTax float64, taxRate float64) float64 {
	// Rumus: pajak = subtotal_excl_tax * tax_rate
	// jika tax_rate dalam persen maka konversi ke pecahan terlebih dahulu
	taxRateFraction := taxRate / 100
	tax := subtotalExclTax * taxRateFraction
	// fmt.Println("taxRate", taxRate)
	// fmt.Println("taxRateFraction", taxRateFraction)
	// fmt.Println("subtotalExclTax", subtotalExclTax)
	// fmt.Println("tax", tax)
	return tax
}

func ContainsString(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func FormatCurrency(amount float64) string {
	return FormatFloatWithThousandSeparator(amount)
}

func FormatFloatWithThousandSeparator(number float64) string {
	// Format angka menjadi string dengan 2 digit desimal
	formatted := fmt.Sprintf("%.2f", number)

	// Pisahkan bagian integer dan desimal
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	// decimalPart := parts[1]

	// Tambahkan separator ribuan ke bagian integer
	var result strings.Builder
	length := len(integerPart)
	for i, char := range integerPart {
		result.WriteRune(char)
		if (length-i-1)%3 == 0 && i != length-1 {
			result.WriteString(",")
		}
	}

	// Gabungkan bagian integer dan desimal
	// result.WriteString(".")
	// result.WriteString(decimalPart)

	return result.String()
}

func URLify(str string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\t':
			return '-'
		case '&', '=', '?', '#', '+':
			return '_'
		default:
			return r
		}
	}, str)
}

func GenerateJWT(userID string, expiredAt int64, secretKey string) (string, error) {
	claims := jwt.StandardClaims{
		Id:        userID,
		ExpiresAt: expiredAt,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	// fmt.Println("token: ", config.App.Server.SecretKey)
	return signedToken, nil
}

func FileHeaderToBytes(fileHeader *multipart.FileHeader) ([]byte, error) {
	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Use a buffer to read the file content
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Return the content as a byte slice
	return buf.Bytes(), nil
}

func FilenameTrimSpace(filename string) string {
	return strings.ReplaceAll(strings.ReplaceAll(filename, " ", "-"), "%", "")
}

func ReduceMap(data map[string]interface{}, keys []string) map[string]interface{} {
	res := make(map[string]interface{})
	for _, key := range keys {
		if val, exists := data[key]; exists {
			res[key] = val
		}
	}

	return res
}

func ParsePhoneNumber(value string, country string) string {

	if country == "" {
		country = "ID"
	}
	num, err := libphonenumber.Parse(value, country)
	if err != nil {
		return value
	}
	countryCode := num.CountryCode
	nationalNumber := num.NationalNumber

	return fmt.Sprintf("%d%d", *countryCode, *nationalNumber)
}

func PtrInt(value int) *int {
	return &value
}

// Helper function to return pointer to a float64
func PtrFloat64(value float64) *float64 {
	return &value
}

func AmountRound(x float64, decimalPlace int) float64 {
	multiplier := math.Pow(10, float64(decimalPlace))
	return math.Round(x*multiplier) / multiplier
}

func IntegerToRoman(number int) string {
	maxRomanNumber := 3999
	if number > maxRomanNumber {
		return strconv.Itoa(number)
	}

	conversions := []struct {
		value int
		digit string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var roman strings.Builder
	for _, conversion := range conversions {
		for number >= conversion.value {
			roman.WriteString(conversion.digit)
			number -= conversion.value
		}
	}

	return roman.String()
}

func GenerateRandomNumber(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charsetNumber[seededRand.Intn(len(charsetNumber))]
	}
	return string(b)
}

func GenerateOrderReceipt(data ReceiptData, templatePath string) ([]byte, error) {
	if templatePath == "" {
		templatePath = "templates/invoice.html"
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return nil, err
	}

	// 2. Generate PDF dari HTML string
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}
	// fmt.Println(htmlBuf.String())
	page := wkhtmltopdf.NewPageReader(strings.NewReader(htmlBuf.String()))
	page.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(page)
	page.DisableSmartShrinking.Set(true)
	page.FooterFontSize.Set(8)

	pdfg.Dpi.Set(300)
	pdfg.PageWidth.Set(57) // Set to receipt width in millimeters
	pdfg.MarginLeft.Set(3)
	pdfg.MarginRight.Set(3)
	pdfg.MarginBottom.Set(3)
	pdfg.MarginTop.Set(3)

	if err := pdfg.Create(); err != nil {
		return nil, err
	}

	// 3. Return PDF sebagai []byte
	return pdfg.Bytes(), nil
}
func GenerateInvoicePDF(data InvoicePDF, templatePath string, footer string) ([]byte, error) {
	// 1. Render HTML dari template
	if templatePath == "" {
		templatePath = "templates/invoice.html"
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return nil, err
	}

	// 2. Generate PDF dari HTML string
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}
	// fmt.Println(htmlBuf.String())
	page := wkhtmltopdf.NewPageReader(strings.NewReader(htmlBuf.String()))
	page.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(page)
	page.DisableSmartShrinking.Set(true)
	page.FooterRight.Set("[page]")
	if footer != "" {
		page.FooterLeft.Set(footer)
	}
	page.FooterFontSize.Set(8)

	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.MarginLeft.Set(15)
	pdfg.MarginRight.Set(15)
	pdfg.MarginBottom.Set(15)
	pdfg.MarginTop.Set(15)

	if err := pdfg.Create(); err != nil {
		return nil, err
	}

	// 3. Return PDF sebagai []byte
	return pdfg.Bytes(), nil
}

func FormatRupiah(amount float64) string {
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("%.0f", amount)
}

func NumToAlphabet(num int) string {
	b := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	res := []rune{}
	for num > 0 {
		res = append(res, b[num%26-1])
		num /= 26
	}
	return string(resverse(res))
}

func resverse(r []rune) []rune {
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return r
}

// LogPrint print error message with RFC3339 timestamp.
// This function is used to log error message from goroutine.
func LogPrint(v ...any) {
	log.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), v)
}
func LogPrintf(format string, v ...any) {
	content := fmt.Sprintf(format, v...)
	log.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), content)
}

func GetMimeType(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	mtype := mimetype.Detect(fileBytes)

	return mtype.String()

}
