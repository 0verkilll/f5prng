package f5prng

import (
	"github.com/0verkilll/sha1"
)

// SHA1Hasher wraps the sha1 package to implement the Hasher interface.
// This provides SHA-1 hashing compatible with Java's SHA1PRNG algorithm.
//
// The wrapper holds the concrete *sha1.SHA1 so it can expose SumInto directly
// and satisfy HasherWithSumInto, avoiding a 20-byte allocation per hash call on
// SecureRandom's state-update hot path.
type SHA1Hasher struct {
	inner *sha1.SHA1
}

// NewSHA1Hasher creates a new SHA-1 hasher suitable for use with SecureRandom.
//
// Example:
//
//	hasher := NewSHA1Hasher()
//	sr := NewSecureRandom(hasher)
func NewSHA1Hasher() Hasher {
	return &SHA1Hasher{
		inner: sha1.NewSHA1(sha1.NewBigEndian()),
	}
}

// Sum computes and returns the SHA-1 hash of the provided data.
func (h *SHA1Hasher) Sum(data []byte) ([]byte, error) {
	return h.inner.Sum(data)
}

// SumInto computes the SHA-1 hash of data and writes it into dst. dst must be
// at least 20 bytes. This is the zero-allocation path used by SecureRandom in
// hot loops (e.g. brute-force key recovery) where the per-call Sum allocation
// dominates GC pressure.
func (h *SHA1Hasher) SumInto(dst, data []byte) error {
	return h.inner.SumInto(dst, data)
}

// Reset clears the internal state of the hasher.
func (h *SHA1Hasher) Reset() {
	h.inner.Reset()
}

// BlockSize returns SHA-1's block size (64 bytes).
func (h *SHA1Hasher) BlockSize() int {
	return h.inner.BlockSize()
}

// Size returns SHA-1's output size (20 bytes).
func (h *SHA1Hasher) Size() int {
	return h.inner.Size()
}

// Compile-time interface checks.
var (
	_ Hasher            = (*SHA1Hasher)(nil)
	_ HasherWithSumInto = (*SHA1Hasher)(nil)
)
