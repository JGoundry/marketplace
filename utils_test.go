package main

import (
	"testing"
)

func FuzzPasswordHashing(f *testing.F) {
	f.Add("")
	f.Add("hello")
	f.Add("你好")

	f.Fuzz(func(t *testing.T, password string) {
		if len(password) > 72 {
			password = password[:72]
		}
		hashedPassword, err := hashPassword(password)
		if err != nil ||
			!checkPasswordHash(password, hashedPassword) ||
			checkPasswordHash(password, password) {
			t.FailNow()
		}
	})
}

func TestConvertMoneyPrintable(t *testing.T) {
	inputToAnswer := map[int]float64{
		137: 1.37,
		100: 1,
	}

	for input, answer := range inputToAnswer {
		if convertMoneyPrintable(input) != answer {
			t.FailNow()
		}
	}
}
