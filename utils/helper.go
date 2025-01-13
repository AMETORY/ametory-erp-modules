package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
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
