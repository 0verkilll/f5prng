package f5prng

import (
	"testing"
)

func TestNewDefaultHasher(t *testing.T) {
	hasher := NewDefaultHasher()

	// Verify it's a valid Hasher
	if hasher == nil {
		t.Fatal("NewDefaultHasher returned nil")
	}

	// Test basic operations
	data := []byte("test")
	hash, err := hasher.Sum(data)
	if err != nil {
		t.Fatalf("Sum() returned error: %v", err)
	}

	// SHA-1 produces 20 bytes
	if len(hash) != 20 {
		t.Errorf("Hash length = %d, want 20", len(hash))
	}

	// Verify Size() returns 20
	if hasher.Size() != 20 {
		t.Errorf("Size() = %d, want 20", hasher.Size())
	}

	// Verify BlockSize() returns 64
	if hasher.BlockSize() != 64 {
		t.Errorf("BlockSize() = %d, want 64", hasher.BlockSize())
	}
}

func TestNewDefaultHasher_UsableWithSecureRandom(t *testing.T) {
	// Verify NewDefaultHasher works with SecureRandom
	hasher := NewDefaultHasher()
	sr := NewSecureRandom(hasher)

	if err := sr.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}
	b := sr.NextBytes(20)

	if len(b) != 20 {
		t.Errorf("NextBytes(20) returned %d bytes", len(b))
	}

	sr.Clear()
}
