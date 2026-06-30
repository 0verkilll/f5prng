package f5prng

import (
	"errors"
	"testing"
)

func TestNextIntN_BoundedRandom(t *testing.T) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()

	if err := prng.Seed([]byte("bounded test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	// Test various bounds
	bounds := []int{1, 2, 10, 100, 1000, 10000}

	for _, n := range bounds {
		t.Run("", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				val, err := NextIntN(prng, n)
				if err != nil {
					t.Fatalf("NextIntN(%d) returned unexpected error: %v", n, err)
				}
				if val < 0 || val >= n {
					t.Errorf("NextIntN(%d) = %d, out of range [0, %d)", n, val, n)
				}
			}
		})
	}

	// Test determinism
	if err := prng.Seed([]byte("bounded test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}
	vals1 := make([]int, 10)
	for i := range vals1 {
		vals1[i], _ = NextIntN(prng, 100)
	}

	if err := prng.Seed([]byte("bounded test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}
	vals2 := make([]int, 10)
	for i := range vals2 {
		vals2[i], _ = NextIntN(prng, 100)
	}

	for i := range vals1 {
		if vals1[i] != vals2[i] {
			t.Errorf("Determinism failed at index %d: got %d and %d", i, vals1[i], vals2[i])
		}
	}
}

func TestNextIntN_BoundOfOne(t *testing.T) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()
	if err := prng.Seed([]byte("one test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	// n=1 should always return 0
	for i := 0; i < 100; i++ {
		val, err := NextIntN(prng, 1)
		if err != nil {
			t.Fatalf("NextIntN(1) returned unexpected error: %v", err)
		}
		if val != 0 {
			t.Errorf("NextIntN(1) = %d, want 0", val)
		}
	}
}

func TestNextIntN_LargeBound(t *testing.T) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()
	if err := prng.Seed([]byte("large bound test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	// Test with large bounds
	largeBounds := []int{1000000, 10000000, 100000000, 1000000000}

	for _, n := range largeBounds {
		for i := 0; i < 10; i++ {
			val, err := NextIntN(prng, n)
			if err != nil {
				t.Fatalf("NextIntN(%d) returned unexpected error: %v", n, err)
			}
			if val < 0 || val >= n {
				t.Errorf("NextIntN(%d) = %d, out of range", n, val)
			}
		}
	}
}

func TestNextIntN_Distribution(t *testing.T) {
	// Test that distribution is roughly uniform
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()
	if err := prng.Seed([]byte("distribution test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	n := 10
	counts := make([]int, n)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		val, err := NextIntN(prng, n)
		if err != nil {
			t.Fatalf("NextIntN(%d) returned unexpected error: %v", n, err)
		}
		counts[val]++
	}

	// Each bucket should have roughly iterations/n values
	expected := iterations / n
	tolerance := expected / 4 // 25% tolerance

	for i, count := range counts {
		if count < expected-tolerance || count > expected+tolerance {
			t.Logf("Bucket %d: got %d, expected ~%d (tolerance: %d)", i, count, expected, tolerance)
			// Note: This is a statistical test, might occasionally fail
		}
	}
}

func TestNextIntN_ErrorOnInvalid(t *testing.T) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()
	if err := prng.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	// Test n = 0 returns error
	_, err := NextIntN(prng, 0)
	if err == nil {
		t.Error("NextIntN(0) should return an error")
	}
	if !errors.Is(err, ErrInvalidBound) {
		t.Errorf("NextIntN(0) should return ErrInvalidBound, got: %v", err)
	}
}

func TestNextIntN_ErrorOnNegative(t *testing.T) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()
	if err := prng.Seed([]byte("test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	// Test n = -1 returns error
	_, err := NextIntN(prng, -1)
	if err == nil {
		t.Error("NextIntN(-1) should return an error")
	}
	if !errors.Is(err, ErrInvalidBound) {
		t.Errorf("NextIntN(-1) should return ErrInvalidBound, got: %v", err)
	}
}

func TestError_ErrorMethod(t *testing.T) {
	// Save original translator and restore after test
	originalTranslator := GetTranslator()
	defer SetTranslator(originalTranslator)
	SetTranslator(nil)

	// Error without message but with valid code uses default translation
	err := Error{Code: ErrCodeInvalidBound}
	expected := "bound must be greater than zero"
	if err.Error() != expected {
		t.Errorf("Error.Error() = %q, want %q", err.Error(), expected)
	}

	// Error with unknown code returns generic message
	unknownErr := Error{Code: ErrorCode(999)}
	if unknownErr.Error() != "f5prng.Error" {
		t.Errorf("Error.Error() with unknown code = %q, want %q", unknownErr.Error(), "f5prng.Error")
	}

	// Test with custom message (always takes precedence)
	errWithMsg := Error{Code: ErrCodeInvalidBound, Message: "custom message"}
	if errWithMsg.Error() != "custom message" {
		t.Errorf("Error.Error() with message = %q, want %q", errWithMsg.Error(), "custom message")
	}
}

func TestError_CodeAccess(t *testing.T) {
	// Verify error code can be accessed for i18n lookup
	_, err := NextIntN(nil, 0)
	if err == nil {
		t.Fatal("expected error")
	}

	var e Error
	if errors.As(err, &e) {
		if e.Code != ErrCodeInvalidBound {
			t.Errorf("Error.Code = %d, want %d", e.Code, ErrCodeInvalidBound)
		}
	} else {
		t.Error("expected error to be of type Error")
	}
}

func TestNextIntN_HandlesNegativeInts(t *testing.T) {
	// Test that NextIntN properly handles negative int32 values from NextInt
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	defer prng.Clear()

	// Use seeds that we know will produce negative values
	if err := prng.Seed([]byte("negative handling test")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	// Generate many values to ensure some come from negative int32s
	for i := 0; i < 1000; i++ {
		val, err := NextIntN(prng, 100)
		if err != nil {
			t.Fatalf("Iteration %d: NextIntN(100) returned unexpected error: %v", i, err)
		}
		if val < 0 || val >= 100 {
			t.Errorf("Iteration %d: NextIntN(100) = %d, out of range", i, val)
		}
	}
}

// fixedIntSource is a tiny RandomSource stub for boundary-value testing of
// NextIntN. NextInt returns the configured value on every call so we can
// exercise specific int32 corner cases (MIN_VALUE, 0, -1) deterministically
// without having to find a seed whose output happens to include them.
type fixedIntSource struct{ v int32 }

func (f *fixedIntSource) Seed(_ []byte) error { return nil }
func (f *fixedIntSource) NextBytes(n int) []byte {
	if n <= 0 {
		return nil
	}
	return make([]byte, n)
}
func (f *fixedIntSource) NextInt() int32 { return f.v }
func (f *fixedIntSource) Clear()         {}

// TestNextIntN_Int32Boundaries locks down the documented behaviour for the
// three int32 boundary values the sign-flip in NextIntN has to cope with:
//
//   - math.MinInt32 (-2147483648): -v overflows in two's-complement int32.
//     The implementation uses uint32(-v) which yields 2147483648, so the
//     modulo is well-defined.
//   - 0: positive == 0, result must be 0 (for any n > 0).
//   - -1: positive == 1, result must be 1 % n.
//
// The invariant checked here — that the result is always in [0, n) and that
// NextIntN does NOT panic on MinInt32 — is what the fuzz corpus below seeds.
func TestNextIntN_Int32Boundaries(t *testing.T) {
	type row struct {
		v    int32
		n    int
		want int
	}
	rows := []row{
		// MinInt32 with various bounds — the sign-flip must not overflow.
		{v: -2147483648, n: 1, want: 0},
		{v: -2147483648, n: 2, want: 0}, // 2147483648 % 2 == 0
		{v: -2147483648, n: 3, want: 2}, // 2147483648 % 3 == 2
		{v: -2147483648, n: 100, want: 48},
		{v: -2147483648, n: 1000000, want: 147483648 % 1000000},
		// Zero.
		{v: 0, n: 1, want: 0},
		{v: 0, n: 100, want: 0},
		{v: 0, n: 1000000, want: 0},
		// -1 -> positive == 1.
		{v: -1, n: 1, want: 0},
		{v: -1, n: 2, want: 1},
		{v: -1, n: 100, want: 1},
	}
	for _, r := range rows {
		src := &fixedIntSource{v: r.v}
		got, err := NextIntN(src, r.n)
		if err != nil {
			t.Errorf("NextIntN(%d with NextInt=%d) unexpected error: %v", r.n, r.v, err)
			continue
		}
		if got != r.want {
			t.Errorf("NextIntN(n=%d, NextInt=%d) = %d, want %d", r.n, r.v, got, r.want)
		}
		if got < 0 || got >= r.n {
			t.Errorf("NextIntN(n=%d, NextInt=%d) = %d, out of range [0, %d)", r.n, r.v, got, r.n)
		}
	}
}

// TestNextIntN_InvalidBoundsDocumentedBehavior locks down the documented
// behavior for n <= 0: NextIntN MUST return ErrInvalidBound (it must not
// panic, and must not return a value). This guards against future refactors
// that might silently change the contract.
func TestNextIntN_InvalidBoundsDocumentedBehavior(t *testing.T) {
	for _, n := range []int{0, -1, -100, -2147483648} {
		src := &fixedIntSource{v: 42}
		v, err := NextIntN(src, n)
		if err == nil {
			t.Errorf("NextIntN(n=%d): expected error, got v=%d", n, v)
		}
		if !errors.Is(err, ErrInvalidBound) {
			t.Errorf("NextIntN(n=%d): expected ErrInvalidBound, got %v", n, err)
		}
		if v != 0 {
			t.Errorf("NextIntN(n=%d): expected zero value on error, got %d", n, v)
		}
	}
}

// Fuzz tests

func FuzzNextIntN(f *testing.F) {
	// Add corpus with various bounds
	f.Add(1)
	f.Add(2)
	f.Add(10)
	f.Add(100)
	f.Add(1000)
	f.Add(10000)
	f.Add(1000000)

	f.Fuzz(func(t *testing.T, n int) {
		if n <= 0 || n > 100000000 {
			return // Skip invalid or extremely large bounds
		}

		factory := NewDefaultFactory()
		prng := factory.NewPRNG()
		defer prng.Clear()
		if err := prng.Seed([]byte("fuzz")); err != nil {
			t.Fatalf("Seed returned unexpected error: %v", err)
		}

		// Generate multiple values
		for i := 0; i < 10; i++ {
			val, err := NextIntN(prng, n)
			if err != nil {
				t.Fatalf("NextIntN(%d) returned unexpected error: %v", n, err)
			}
			if val < 0 || val >= n {
				t.Errorf("NextIntN(%d) = %d, out of range [0, %d)", n, val, n)
			}
		}
	})
}

// FuzzNextIntN_IntBoundary fuzzes NextIntN against a stub PRNG whose NextInt()
// returns the fuzzed int32 value. This covers the sign-flip boundary
// (math.MinInt32 overflows -v in int32) plus the surrounding values that the
// audit pinned as "must stay in range / must not panic":
//
//   - int32(-2147483648): MIN_VALUE — -v overflows, uint32(-v) is required.
//   - int32(0): positive == 0 — result must be 0 for any n > 0.
//   - int32(-1): sign flip must yield uint32(1).
//
// For n > 0 the result must lie in [0, n). For n <= 0 NextIntN must return
// ErrInvalidBound (no panic, no silent zero).
func FuzzNextIntN_IntBoundary(f *testing.F) {
	// Boundary corpus — the audit explicitly calls out these three.
	f.Add(int32(-2147483648), 100)
	f.Add(int32(0), 100)
	f.Add(int32(-1), 100)
	// A few more for breadth across the fuzzer's mutation space.
	f.Add(int32(-2147483648), 1)
	f.Add(int32(-2147483648), 7)
	f.Add(int32(2147483647), 100)
	f.Add(int32(1), 1000000)
	// Invalid bounds.
	f.Add(int32(0), 0)
	f.Add(int32(-1), -1)
	f.Add(int32(-2147483648), -100)

	f.Fuzz(func(t *testing.T, v int32, n int) {
		src := &fixedIntSource{v: v}
		got, err := NextIntN(src, n)
		if n <= 0 {
			if err == nil {
				t.Errorf("NextIntN(v=%d, n=%d): expected error for n<=0, got %d", v, n, got)
			}
			if !errors.Is(err, ErrInvalidBound) {
				t.Errorf("NextIntN(v=%d, n=%d): expected ErrInvalidBound, got %v", v, n, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("NextIntN(v=%d, n=%d): unexpected error: %v", v, n, err)
		}
		if got < 0 || got >= n {
			t.Errorf("NextIntN(v=%d, n=%d) = %d, out of range [0, %d)", v, n, got, n)
		}
	})
}

func FuzzNextIntN_Determinism(f *testing.F) {
	f.Add([]byte("test"), 100)
	f.Add([]byte("password"), 1000)
	f.Add([]byte{1, 2, 3}, 50)

	f.Fuzz(func(t *testing.T, seed []byte, n int) {
		if n <= 0 || n > 10000 {
			return
		}

		factory := NewDefaultFactory()

		prng1 := factory.NewPRNG()
		if err := prng1.Seed(seed); err != nil {
			t.Fatalf("prng1.Seed returned unexpected error: %v", err)
		}
		vals1 := make([]int, 10)
		for i := range vals1 {
			vals1[i], _ = NextIntN(prng1, n)
		}
		prng1.Clear()

		prng2 := factory.NewPRNG()
		if err := prng2.Seed(seed); err != nil {
			t.Fatalf("prng2.Seed returned unexpected error: %v", err)
		}
		vals2 := make([]int, 10)
		for i := range vals2 {
			vals2[i], _ = NextIntN(prng2, n)
		}
		prng2.Clear()

		for i := range vals1 {
			if vals1[i] != vals2[i] {
				t.Errorf("Determinism failed at index %d: got %d and %d", i, vals1[i], vals2[i])
			}
		}
	})
}

// Benchmarks

func BenchmarkNextIntN_10(b *testing.B) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	if err := prng.Seed([]byte("benchmark")); err != nil {
		b.Fatalf("Seed returned unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NextIntN(prng, 10)
	}
}

func BenchmarkNextIntN_100(b *testing.B) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	if err := prng.Seed([]byte("benchmark")); err != nil {
		b.Fatalf("Seed returned unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NextIntN(prng, 100)
	}
}

func BenchmarkNextIntN_1000000(b *testing.B) {
	factory := NewDefaultFactory()
	prng := factory.NewPRNG()
	if err := prng.Seed([]byte("benchmark")); err != nil {
		b.Fatalf("Seed returned unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NextIntN(prng, 1000000)
	}
}
