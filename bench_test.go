package f5prng

import (
	"testing"

	"github.com/0verkilll/sha1"
)

// BenchmarkClear measures the cost of SecureRandom.Clear() on an initialized
// (seeded) PRNG. This is the hot path in brute-force key-recovery scenarios
// where each candidate password is Clear()'d before the next Seed() — if
// Clear gets slower, every candidate pays for it.
//
// Note: the post-audit version of Clear() also zeroes outputBuf and calls
// hasher.Reset() to wipe the wrapped hasher's internal block state. The
// extra work is trivially small but this benchmark makes that cost visible.
func BenchmarkClear(b *testing.B) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	if err := prng.Seed([]byte("benchmark")); err != nil {
		b.Fatalf("Seed returned unexpected error: %v", err)
	}
	// Prime remainder / state by drawing some bytes so Clear has full buffers
	// to zero (otherwise it zeroes already-zero memory, which is unrealistic).
	_ = prng.NextBytes(17)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prng.Clear()
	}
}

// BenchmarkSeed_SHA1Digest measures Seed() with a typical 20-byte SHA-1
// digest as the seed. F5 key-recovery pipelines hash the candidate password
// with SHA-1 and feed the 20-byte digest to Seed() on every candidate, so
// 20 bytes is the realistic input length to optimise against. The existing
// BenchmarkSeed uses a short ASCII seed which under-reports the real cost.
func BenchmarkSeed_SHA1Digest(b *testing.B) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()

	// Produce a realistic 20-byte SHA-1 digest to feed into Seed().
	h := sha1.NewSHA1(sha1.NewBigEndian())
	seed, err := h.Sum([]byte("benchmark candidate password"))
	if err != nil {
		b.Fatalf("sha1.Sum returned unexpected error: %v", err)
	}
	if len(seed) != 20 {
		b.Fatalf("expected 20-byte SHA-1 digest, got %d bytes", len(seed))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := prng.Seed(seed); err != nil {
			b.Fatalf("Seed returned unexpected error: %v", err)
		}
	}
}

// BenchmarkNextIntN_100K benchmarks NextIntN against a 100000 bound. This
// mirrors the realistic Fisher-Yates permutation size used by F5 at decode
// time (~100K DCT coefficients per JPEG), making it the closest bound to
// production workloads. If this regresses, key-recovery end-to-end runtime
// regresses proportionally.
func BenchmarkNextIntN_100K(b *testing.B) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	if err := prng.Seed([]byte("benchmark")); err != nil {
		b.Fatalf("Seed returned unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NextIntN(prng, 100000)
	}
}
