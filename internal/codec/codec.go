package codec

import (
	"strings"
)

const (
	CodeLength = 10
	Alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	PadChar    = 'a'
)

const base = int64(len(Alphabet))

// EncodeID converts numeric id to a fixed-length base63 code
func EncodeID(id int64) (string, error) {
	if id < 0 {
		return "", ErrInvalidID
	}

	if id == 0 {
		return strings.Repeat(string(PadChar), CodeLength), nil
	}

	digits := make([]byte, 0, CodeLength)
	for id > 0 {
		rem := id % base
		digits = append(digits, Alphabet[rem])
		id /= base
	}

	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}

	if len(digits) > CodeLength {
		return "", ErrInvalidID
	}

	if len(digits) < CodeLength {
		padding := strings.Repeat(string(PadChar), CodeLength-len(digits))
		return padding + string(digits), nil
	}

	return string(digits), nil
}

// DecodeCode converts a fixed-length base63 code to numeric id
func DecodeCode(code string) (int64, error) {
	if len(code) != CodeLength {
		return 0, ErrInvalidCode
	}

	var id int64
	for i := 0; i < len(code); i++ {
		idx := strings.IndexByte(Alphabet, code[i])
		if idx == -1 {
			return 0, ErrInvalidCode
		}

		id = id*base + int64(idx)
	}

	return id, nil
}
