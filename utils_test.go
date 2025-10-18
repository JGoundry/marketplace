package main

import "testing"

var passwords []string = []string{"password", "thispassword", "verysecure"}

func TestPasswordHashing(t *testing.T) {
	for _, password := range passwords {
		hashedPassword, err := hashPassword(password)
		if err != nil ||
			!checkPasswordHash(password, hashedPassword) ||
			checkPasswordHash(password, password) {
			t.FailNow()
		}
	}
}

func TestConvertMoneyPrintable(t *testing.T) {
	if convertMoneyPrintable(137) != 1.37 {
		t.FailNow()
	}
}
