package f5prng

// RandomSource defines the interface for pseudo-random number generation.
// This abstraction enables different PRNG implementations while maintaining
// the Interface Segregation Principle by providing only essential methods.
//
// The interface is designed to match Java's SecureRandom behavior for
// compatibility with F5 steganography's permutation and embedding algorithms.
//
// Method Signatures:
//   - NextBytes allocates and returns a new slice (matches Java behavior)
//   - NextInt returns an unbounded int32 (matches Java SecureRandom.nextInt())
//
// SECURITY NOTE: Implementations of this interface are NOT cryptographically
// secure. Do not use for cryptographic key generation, session tokens, or
// other security-critical purposes. Use crypto/rand for secure randomness.
type RandomSource interface {
	// Seed initializes or re-initializes the random source with the given seed.
	// For deterministic PRNGs, the same seed must produce the same sequence
	// of random values across all calls.
	//
	// Parameters:
	//   seed - The seed bytes for initializing the random state
	//
	// Returns:
	//   error - Error if seeding fails (e.g., hash computation error), nil otherwise
	Seed(seed []byte) error

	// NextBytes generates and returns n random bytes.
	// The bytes are generated from the internal PRNG state and advance
	// the state for subsequent calls.
	//
	// This method allocates a new slice for each call, matching Java's behavior.
	// For high-performance scenarios where allocation matters, consider using
	// a buffer-based approach in your implementation.
	//
	// Parameters:
	//   n - The number of random bytes to generate (must be >= 0)
	//
	// Returns:
	//   A newly allocated slice containing n pseudo-random bytes.
	//   Returns an empty slice if n <= 0.
	NextBytes(n int) []byte

	// NextInt generates and returns the next random 32-bit integer.
	// For Java compatibility, this returns a signed int32 value
	// in the range [-2147483648, 2147483647].
	//
	// This is an unbounded random int32, matching Java's SecureRandom.nextInt().
	// For bounded random integers, use the NextIntN helper function.
	//
	// Returns:
	//   A pseudo-random int32 value
	NextInt() int32

	// Clear securely zeros all internal state to prevent memory disclosure.
	// This method MUST be called when done using the RandomSource, especially
	// if seeded with sensitive data (passwords, keys, etc.).
	//
	// After calling Clear(), the instance is no longer usable and must be
	// reseeded before generating new random data.
	//
	// Multiple calls to Clear() are safe (idempotent).
	//
	// Best practice: Use defer to ensure Clear() is called:
	//   rs := factory.NewPRNG()
	//   defer rs.Clear()
	//   rs.Seed(sensitiveData)
	//   // ... use rs
	Clear()
}

// RandomSourceWithBytesInto is an optional extension of RandomSource that
// supports zero-allocation byte generation by writing into a caller-provided
// buffer. Consumers that extract bytes in tight loops (e.g. per-message-byte
// XOR masks in F5 extraction) should type-assert for this interface and use
// NextBytesInto when available, falling back to NextBytes otherwise.
//
// The byte stream produced by NextBytesInto(buf) is identical to the stream
// produced by NextBytes(len(buf)), so callers can mix the two freely.
type RandomSourceWithBytesInto interface {
	// NextBytesInto fills dst with pseudo-random bytes. Returns nil on success,
	// or the underlying error if the PRNG is not seeded or the hasher fails.
	NextBytesInto(dst []byte) error
}

// PRNGFactory defines the interface for creating RandomSource instances.
// This factory pattern enables dependency injection and allows different
// PRNG implementations to be plugged in without changing consuming code.
//
// Example usage:
//
//	factory := NewDefaultFactory()
//	prng := factory.NewPRNG()
//	defer prng.Clear()
//	if err := prng.Seed([]byte("password")); err != nil {
//	    // handle error
//	}
//	randomBytes := prng.NextBytes(20)
type PRNGFactory interface {
	// NewPRNG creates and returns a new RandomSource instance.
	// The returned instance is uninitialized and must be seeded before use.
	//
	// Each call creates a new, independent RandomSource. Multiple instances
	// can be used concurrently (though individual instances are not thread-safe).
	//
	// Returns:
	//   A new RandomSource ready to be seeded
	NewPRNG() RandomSource
}

// Hasher defines the interface for cryptographic hash functions.
// This abstraction follows the Single Responsibility Principle by focusing
// solely on hashing operations, and the Dependency Inversion Principle by
// allowing different hash implementations (SHA-1, SHA-256, etc.) to be used
// interchangeably.
//
// For F5 steganography compatibility, use a SHA-1 hasher implementation.
//
// Implementations must be stateful and maintain internal hash state between
// calls. The Reset method allows reuse of the same instance for multiple
// hash operations.
type Hasher interface {
	// Sum computes and returns the hash of the provided data.
	// The returned slice is the final hash value (e.g., 20 bytes for SHA-1).
	// Multiple calls to Sum with the same data must return identical results.
	//
	// Parameters:
	//   data - The input bytes to hash
	//
	// Returns:
	//   hash - The computed hash as a byte slice
	//   error - Error if hash computation fails, nil otherwise
	Sum(data []byte) ([]byte, error)

	// Reset clears the internal state of the hasher, allowing it to be reused
	// for a new hash computation. This is more efficient than creating a new
	// Hasher instance for each operation.
	Reset()

	// BlockSize returns the hash's underlying block size in bytes.
	// For SHA-1, this is 64 bytes. For SHA-256, this is also 64 bytes.
	//
	// This is useful for certain cryptographic operations that need to know
	// the block size, such as HMAC implementations.
	//
	// Returns:
	//   The block size in bytes
	BlockSize() int

	// Size returns the hash's output size in bytes.
	// For SHA-1, this is 20 bytes. For SHA-256, this is 32 bytes.
	//
	// This allows generic code to work with different hash algorithms
	// without hardcoding the output size.
	//
	// Returns:
	//   The hash output size in bytes
	Size() int
}
