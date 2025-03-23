package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ttacon/libphonenumber"
)

func init() {

}

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
	return strings.ReplaceAll(filename, " ", "-")
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
