package main

import (
	"testing"
)

var passwordSeeds []string = []string{
	"",
	"hello",
	"你好",
}

func FuzzPasswordHashing(f *testing.F) {
	for _, seed := range passwordSeeds {
		f.Add(seed)
	}

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

var testConvertMoneyPrintableTable = map[string]struct {
	input    int
	expected float64
}{
	"decimal": {
		input:    137,
		expected: 1.37,
	},
	"integer": {
		input:    100,
		expected: 1,
	},
}

func TestConvertMoneyPrintable(t *testing.T) {
	for name, args := range testConvertMoneyPrintableTable {
		t.Run(name, func(t *testing.T) {
			answer := convertMoneyPrintable(args.input)
			if answer != args.expected {
				t.Errorf("input %v, got %v, expected %v", args.input, answer, args.expected)
			}
		})
	}
}
