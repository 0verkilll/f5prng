package f5prng

import (
	"bytes"
	"testing"
)

func TestPRNGFactory_CreatesValidInstances(t *testing.T) {
	factory := NewDefaultFactory()

	// Create multiple instances
	prng1 := factory.NewPRNG()
	prng2 := factory.NewPRNG()

	// Both should be independent
	if err := prng1.Seed([]byte("seed1")); err != nil {
		t.Fatalf("prng1.Seed returned unexpected error: %v", err)
	}
	if err := prng2.Seed([]byte("seed2")); err != nil {
		t.Fatalf("prng2.Seed returned unexpected error: %v", err)
	}

	bytes1 := prng1.NextBytes(20)
	bytes2 := prng2.NextBytes(20)

	// Different seeds should produce different output
	if bytes.Equal(bytes1, bytes2) {
		t.Error("Different seeds should produce different output")
	}

	// Same seed should produce same output
	prng3 := factory.NewPRNG()
	if err := prng3.Seed([]byte("seed1")); err != nil {
		t.Fatalf("prng3.Seed returned unexpected error: %v", err)
	}
	bytes3 := prng3.NextBytes(20)

	if !bytes.Equal(bytes1, bytes3) {
		t.Error("Same seed should produce same output")
	}

	// Cleanup
	prng1.Clear()
	prng2.Clear()
	prng3.Clear()
}
