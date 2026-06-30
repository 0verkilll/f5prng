package f5prng

import (
	"bytes"
	"testing"
)

func TestSHA1Hasher_Implementation(t *testing.T) {
	hasher := NewSHA1Hasher()

	// Test basic hashing
	data := []byte("test data")
	hash, err := hasher.Sum(data)
	if err != nil {
		t.Fatalf("Sum() returned error: %v", err)
	}

	// SHA-1 produces 20 bytes
	if len(hash) != 20 {
		t.Errorf("Hash length = %d, want 20", len(hash))
	}

	// Test Size()
	if hasher.Size() != 20 {
		t.Errorf("Size() = %d, want 20", hasher.Size())
	}

	// Test BlockSize()
	if hasher.BlockSize() != 64 {
		t.Errorf("BlockSize() = %d, want 64", hasher.BlockSize())
	}

	// Test Reset() doesn't panic
	hasher.Reset()

	// After reset, same input should produce same hash
	hash2, err := hasher.Sum(data)
	if err != nil {
		t.Fatalf("Sum() after Reset() returned error: %v", err)
	}
	if !bytes.Equal(hash, hash2) {
		t.Error("Hash should be same after Reset()")
	}
}
