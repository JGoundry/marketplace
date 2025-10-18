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
		t.Parallel()
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
	t.Parallel()
	for name, args := range testConvertMoneyPrintableTable {
		input := args.input
		expected := args.expected
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			answer := convertMoneyPrintable(input)
			if answer != expected {
				t.Errorf("input %v, got %v, expected %v", input, answer, expected)
			}
		})
	}
}
