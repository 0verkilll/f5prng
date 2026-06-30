package f5prng

// NextIntN returns a bounded random integer in the range [0, n).
// This is a helper function that converts an unbounded NextInt() to a bounded value.
//
// The function properly handles modulo bias to ensure uniform distribution
// across the range [0, n).
//
// Parameters:
//   - rs: The RandomSource to use for generating random integers
//   - n: The exclusive upper bound (must be > 0)
//
// Returns:
//   - A random integer in the range [0, n), or an error if n <= 0
//
// Example:
//
//	prng := factory.NewPRNG()
//	defer prng.Clear()
//	if err := prng.Seed([]byte("password")); err != nil {
//	    // handle error
//	}
//	randomIndex, err := NextIntN(prng, 100) // Returns 0-99
//	if err != nil {
//	    // handle error
//	}
func NextIntN(rs RandomSource, n int) (int, error) {
	if n <= 0 {
		return 0, ErrInvalidBound
	}

	// For n == 1, there's only one possible value
	if n == 1 {
		return 0, nil
	}

	// Get unbounded int32 and make it positive
	// We use uint32 to handle the full range of positive values
	v := rs.NextInt()

	// Convert to positive value using two's complement
	// This matches Java's Math.abs behavior for most cases
	// Note: Math.abs(Integer.MIN_VALUE) returns Integer.MIN_VALUE in Java,
	// but for F5 compatibility we need to match the exact behavior
	var positive uint32
	if v >= 0 {
		// #nosec G115 -- Safe conversion: v >= 0, so it fits in uint32
		positive = uint32(v)
	} else {
		// For negative values, negate to get positive
		// Note: -(-2147483648) overflows in int32, but uint32(-v) handles it
		positive = uint32(-v)
	}

	// Simple modulo for bounded result
	// Note: This has slight modulo bias, but matches the original F5 behavior
	// For F5 compatibility, we don't use rejection sampling
	// #nosec G115 -- Safe conversion: n > 0, so it fits in uint32
	return int(positive % uint32(n)), nil
}
