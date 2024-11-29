package hash

import (
	"testing"
)

func TestHash(t *testing.T) {
	key := "my_secret_key"
	data := []byte("my_data")
	hasher := New(key)

	expectedHashHex := "cac1b4db68fec5ca9b45231ce0723aafb8e16112b16ced7691e327d5f6d8433c"

	actualHash, err := hasher.Hash(data)
	if err != nil {
		t.Fatalf("Failed to hash data: %v", err)
	}

	if actualHash != expectedHashHex {
		t.Errorf("Expected hash %s, got %s", expectedHashHex, actualHash)
	}
}

func TestCheck(t *testing.T) {
	key := "my_secret_key"
	data := []byte("my_data")
	hasher := New(key)

	hash, err := hasher.Hash(data)
	if err != nil {
		t.Fatalf("Failed to hash data: %v", err)
	}

	tests := []struct {
		name          string
		input         []byte
		expectedHash  string
		expectedEqual bool
	}{
		{"Valid hash", data, hash, true},
		{"Invalid hash", []byte("other_data"), hash, false},
		{"Different hash", data, "wronghash", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := hasher.Check(test.input, []byte(test.expectedHash))
			if err != nil {
				t.Fatalf("Check failed: %v", err)
			}
			if result != test.expectedEqual {
				t.Errorf("Expected %v, got %v", test.expectedEqual, result)
			}
		})
	}
}
