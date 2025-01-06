package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func init() {

}

// RandString generates a random string of length n
func RandString(n int) string {
	r := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(r)
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
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
	username = append(username, []rune(RandString(5))...)
	return string(username)
}

func LogJson(data interface{}) {
	jsonString, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonString))
}
