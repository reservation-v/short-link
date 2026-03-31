package codec

import "testing"

func TestEncodeID_KnownCases(t *testing.T) {
	tests := []struct {
		name string
		id   int64
		want string
	}{
		{name: "zero", id: 0, want: "aaaaaaaaaa"},
		{name: "one", id: 1, want: "aaaaaaaaab"},
		{name: "sixty-two", id: 62, want: "aaaaaaaaa_"},
		{name: "sixty-three", id: 63, want: "aaaaaaaaba"},
		{name: "sixty-four", id: 64, want: "aaaaaaaabb"},
		{name: "max-10-digits", id: maxIDForCodeLength(), want: "__________"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeID(tt.id)
			if err != nil {
				t.Fatalf("EncodeID(%d) returned error: %v", tt.id, err)
			}
			if got != tt.want {
				t.Fatalf("EncodeID(%d) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestEncodeID_Errors(t *testing.T) {
	_, err := EncodeID(-1)
	if err != ErrInvalidID {
		t.Fatalf("EncodeID(-1) error = %v, want %v", err, ErrInvalidID)
	}

	_, err = EncodeID(maxIDForCodeLength() + 1)
	if err != ErrInvalidID {
		t.Fatalf("EncodeID(overflow) error = %v, want %v", err, ErrInvalidID)
	}
}

func TestDecodeCode_KnownCases(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int64
	}{
		{name: "all-padding", code: "aaaaaaaaaa", want: 0},
		{name: "one", code: "aaaaaaaaab", want: 1},
		{name: "sixty-two", code: "aaaaaaaaa_", want: 62},
		{name: "sixty-three", code: "aaaaaaaaba", want: 63},
		{name: "sixty-four", code: "aaaaaaaabb", want: 64},
		{name: "max-10-digits", code: "__________", want: maxIDForCodeLength()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeCode(tt.code)
			if err != nil {
				t.Fatalf("DecodeCode(%q) returned error: %v", tt.code, err)
			}
			if got != tt.want {
				t.Fatalf("DecodeCode(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestDecodeCode_Errors(t *testing.T) {
	tests := []string{
		"",
		"short",
		"aaaaaaaaaaa",
		"aaaaaaaaa-",
	}

	for _, code := range tests {
		t.Run(code, func(t *testing.T) {
			_, err := DecodeCode(code)
			if err != ErrInvalidCode {
				t.Fatalf("DecodeCode(%q) error = %v, want %v", code, err, ErrInvalidCode)
			}
		})
	}
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	ids := []int64{
		0,
		1,
		2,
		62,
		63,
		64,
		1024,
		999999,
		1234567890,
		maxIDForCodeLength(),
	}

	for _, id := range ids {
		code, err := EncodeID(id)
		if err != nil {
			t.Fatalf("EncodeID(%d) returned error: %v", id, err)
		}

		got, err := DecodeCode(code)
		if err != nil {
			t.Fatalf("DecodeCode(%q) returned error: %v", code, err)
		}
		if got != id {
			t.Fatalf("round-trip mismatch: id=%d code=%q decoded=%d", id, code, got)
		}
	}
}

func maxIDForCodeLength() int64 {
	var max int64 = 1
	for range CodeLength {
		max *= int64(len(Alphabet))
	}
	return max - 1
}
