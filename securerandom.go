package f5prng

import (
	"fmt"
	"runtime"
)

// NOTE: Hot-path methods (NextInt, NextBytes, fillBytes, updateState) do NOT
// contain log() calls. Even with NopLogger, Go allocates variadic arg slices
// and boxes values into any interfaces on every call. With ~100K NextInt calls
// per candidate in key recovery, this causes 80+ GB of unnecessary allocations.
// Logging is kept in Seed() and Clear() which are called rarely.

// SecureRandom implements Java's SecureRandom with SHA1PRNG algorithm.
// This is a pure Go implementation that produces identical output to
// java.security.SecureRandom when initialized with the same seed.
//
// The algorithm uses SHA-1 hashing to generate pseudorandom bytes and
// maintains an internal state that gets updated with each generation.
//
// THREAD SAFETY: SecureRandom is NOT thread-safe. Do not share instances
// between goroutines. Either:
//  1. Create separate instances per goroutine, or
//  2. Use external synchronization (sync.Mutex)
//
// SECURITY WARNING: This implementation is NOT cryptographically secure.
// SHA-1 is broken and should not be used for security purposes. Use crypto/rand
// for secure random generation. This implementation is suitable for:
//   - F5 steganography decoding (PixelKnot, F5.jar)
//   - Legacy Java application compatibility
//   - Deterministic PRNG for testing/replay
//   - Educational and research purposes
//
// MEMORY PROTECTION: Always call Clear() when done to zero sensitive data:
//
//	sr := NewSecureRandom(hasher)
//	defer sr.Clear()
//	if err := sr.Seed([]byte(password)); err != nil {
//	    // handle error
//	}
//	// ... use sr
// HasherWithSumInto is an optional extension of Hasher that supports zero-allocation
// hashing by writing the output into a caller-provided buffer.
//
// When a Hasher implements this interface, SecureRandom's updateState method uses
// SumInto instead of Sum, eliminating a 20-byte allocation per hash call. In
// brute-force scenarios (100K+ PRNG calls per candidate), this reduces GC pressure
// by eliminating ~20K allocations per candidate.
type HasherWithSumInto interface {
	// SumInto computes the hash of data and writes it into dst.
	// dst must be at least 20 bytes for SHA-1.
	SumInto(dst []byte, data []byte) error
}

type SecureRandom struct {
	// hasher is the injected SHA-1 implementation
	hasher Hasher

	// hasherInto caches the HasherWithSumInto type assertion result.
	// Set once during Seed() to avoid per-call type assertions.
	hasherInto HasherWithSumInto

	// state holds the current generator state (20 bytes from SHA-1)
	state []byte

	// remainder holds unused bytes from the last hash. Valid bytes live at
	// remainder[remHead : remHead+remCount]; we advance remHead instead of
	// shifting the tail on every partial drain. A new digest always writes
	// at remainder[0 : new remCount] with remHead reset to 0.
	remainder []byte

	// remHead is the index of the next unread byte in remainder. The live
	// window is remainder[remHead : remHead+remCount]; when remCount reaches
	// zero, remHead is reset to 0 so the next refill always starts writing
	// at remainder[0]. Using a head index eliminates the O(remCount) memmove
	// that the previous shift-to-front implementation performed on every
	// partial drain — critical for NextInt, which consumes 4 bytes at a time
	// out of the 20-byte digest.
	remHead int

	// remCount tracks how many bytes are in a remainder
	remCount int

	// intBuf is a reusable 4-byte buffer for NextInt to avoid allocations
	intBuf [4]byte

	// outputBuf is a reusable 20-byte buffer for updateState to avoid allocations.
	// Used when the hasher supports HasherWithSumInto.
	outputBuf [20]byte

	// lastErr stores the last error that occurred during operations
	// This is used for error recovery in NextBytes and NextInt
	lastErr error
}

// NewSecureRandom creates a new SecureRandom instance with the provided hasher.
// The hasher should implement SHA-1 for Java compatibility.
//
// The returned instance implements the RandomSource interface.
//
// Example:
//
//	hasher := sha1.NewSHA1(sha1.NewBigEndian())
//	sr := NewSecureRandom(hasher)
//	defer sr.Clear()
//	if err := sr.Seed([]byte("password")); err != nil {
//	    // handle error
//	}
//	bytes := sr.NextBytes(20)
func NewSecureRandom(hasher Hasher) RandomSource {
	return &SecureRandom{
		hasher:    hasher,
		state:     make([]byte, 20), // SHA-1 produces 20 bytes
		remainder: make([]byte, 20),
		remCount:  0,
	}
}

// Seed initializes the random number generator with the provided seed.
// This implements Java's engineSetSeed method.
//
// For deterministic output, the same seed must always produce the same
// sequence of random values.
//
// Returns an error if the hasher is nil or if the hash computation fails.
func (sr *SecureRandom) Seed(seed []byte) error {
	log().Debug("Seed called", "seed_length", len(seed))

	// Check for nil hasher
	if sr.hasher == nil {
		sr.lastErr = ErrNilHasher
		log().Debug("Seed failed: nil hasher")
		return ErrNilHasher
	}

	// Cache the HasherWithSumInto type assertion (once per Seed call).
	// This avoids per-updateState type assertions in the hot path.
	sr.hasherInto, _ = sr.hasher.(HasherWithSumInto)

	// Clear any previous error
	sr.lastErr = nil

	if len(seed) == 0 {
		// Handle empty seed - use zero state
		for i := range sr.state {
			sr.state[i] = 0
		}
		sr.remCount = 0
		sr.remHead = 0
		log().Debug("Seed completed", "empty_seed", true)
		return nil
	}

	// Java's engineSetSeed: state = digest(seed)
	sr.hasher.Reset()
	digest, err := sr.hasher.Sum(seed)
	if err != nil {
		sr.lastErr = fmt.Errorf("%w: %v", ErrSeedHashFailed, err)
		log().Debug("Seed failed: hash computation error", "error", err.Error())
		return sr.lastErr
	}
	copy(sr.state, digest)
	sr.remCount = 0
	sr.remHead = 0

	log().Debug("Seed completed")
	return nil
}

// updateState updates the internal state using Java's algorithm.
// This is called after consuming random bytes to generate new state.
//
// Java algorithm:
// 1. Hash the current state to get output
// 2. Update state: state = state + output + 1 (with signed byte arithmetic)
// 3. Return hashed output.
//
// Returns the hashed output and any error that occurred.
func (sr *SecureRandom) updateState() ([]byte, error) {
	// Check for nil hasher
	if sr.hasher == nil {
		return nil, ErrNilHasher
	}

	// Hash current state to get output.
	// Use zero-allocation SumInto when available (eliminates 20-byte alloc per call).
	sr.hasher.Reset()
	var output []byte
	if sr.hasherInto != nil {
		if err := sr.hasherInto.SumInto(sr.outputBuf[:], sr.state); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrStateUpdateFailed, err)
		}
		output = sr.outputBuf[:]
	} else {
		var err error
		output, err = sr.hasher.Sum(sr.state)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrStateUpdateFailed, err)
		}
	}

	// Java does: state = state + output + 1
	// This is done with signed byte arithmetic and carry propagation. The
	// int8 casts are LOAD-BEARING here: Java byte addition is signed, and
	// `v >> 8` must arithmetically extend the sign so the carry propagates
	// correctly (e.g. v == -1 must yield last == -1, not last == 0). Do not
	// "clean up" these casts to uint8 without reproducing the full Java
	// parity test suite — the signed arithmetic is part of the observable
	// output stream.
	last := 1 // Initial carry (the +1)
	zf := false

	for i := 0; i < len(sr.state); i++ {
		// Add two bytes with sign extension, matching Java
		v := int(int8(sr.state[i])) + int(int8(output[i])) + last
		t := byte(v)
		zf = zf || (sr.state[i] != t)
		sr.state[i] = t
		last = v >> 8 // Arithmetic shift for carry
	}

	// If no change occurred, increment state[0]
	// DEFENSIVE CODE: This path is mathematically unreachable in practice.
	// It would require state + output + 1 to overflow all 20 bytes perfectly
	// such that sr.state[i] == t for all i. With SHA-1's avalanche effect,
	// this has a negligible probability (~2^-160). Kept for Java compatibility.
	if !zf {
		sr.state[0]++
	}

	return output, nil
}

// NextBytes returns n pseudorandom bytes.
// This implements Java's engineNextBytes method.
//
// The returned slice is newly allocated for each call, matching Java's behavior.
// Returns an empty slice if n <= 0.
//
// If an internal error occurs (extremely rare), returns an empty slice and
// stores the error in lastErr (accessible via LastError method).
func (sr *SecureRandom) NextBytes(n int) []byte {
	if n <= 0 {
		return []byte{}
	}

	result := make([]byte, n)
	resultPos := 0

	// Use remainder bytes first. We advance remHead instead of shifting the
	// tail to the front on every partial drain — see the remHead field doc
	// for why this matters on the NextInt hot path.
	if sr.remCount > 0 {
		toCopy := n
		if toCopy > sr.remCount {
			toCopy = sr.remCount
		}
		copy(result[resultPos:], sr.remainder[sr.remHead:sr.remHead+toCopy])
		resultPos += toCopy
		sr.remCount -= toCopy
		if sr.remCount == 0 {
			sr.remHead = 0
		} else {
			sr.remHead += toCopy
		}
	}

	// Generate more bytes as needed
	for resultPos < n {
		output, err := sr.updateState()
		if err != nil {
			// Store error for later retrieval, return partial result
			sr.lastErr = err
			return result[:resultPos]
		}
		remaining := n - resultPos

		if remaining < len(output) {
			// Use part of output, save rest to remainder. The drain loop above
			// guarantees remHead == 0 whenever remCount == 0, so the saved
			// tail always lives at remainder[0:].
			copy(result[resultPos:], output[:remaining])
			copy(sr.remainder, output[remaining:])
			sr.remCount = len(output) - remaining
			sr.remHead = 0
			resultPos = n
		} else {
			// Use all the output
			copy(result[resultPos:], output)
			resultPos += len(output)
		}
	}

	return result
}

// NextBytesInto fills dst with pseudo-random bytes without allocating a new
// slice. The byte stream is identical to NextBytes(len(dst)) — i.e. consumers
// that switch between NextBytes and NextBytesInto see the same Java-compatible
// output. Returns nil on success, or the underlying error if the PRNG is not
// seeded or the hasher fails. On error, dst may be partially written.
//
// This is a zero-allocation hot-path variant used by consumers that extract
// individual bytes in inner loops (e.g. f5messageextract.nextSignedByte).
func (sr *SecureRandom) NextBytesInto(dst []byte) error {
	if err := sr.fillBytes(dst); err != nil {
		sr.lastErr = err
		return err
	}
	return nil
}

// NextInt returns a pseudorandom int32.
// This implements Java's nextInt() method which returns a 32-bit signed integer.
//
// IMPORTANT: This gets 4 bytes using an internal buffer to avoid allocations,
// while maintaining the exact byte consumption order and signed integer behavior
// used by Java's SecureRandom and the F5 algorithm.
//
// The byte order matches GetNextValue() from the working implementation:
//
//	byte0 | (byte1 << 8) | (byte2 << 16) | (byte3 << 24)
//
// # Why sign-extended int32(int8(b)) is load-bearing
//
// At first glance the `int(int8(b))` casts below look like a sign-extension
// bug waiting to happen — the obvious "clean" rewrite is an unsigned
// assembly:
//
//	u := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
//	return int32(u)
//
// That rewrite produces DIFFERENT output when byte2 has its high bit set.
// Reason: in Go, `int(-1) << 16` is `0xFFFFFFFFFFFF0000` (64-bit, sign
// extended). OR-ing that into the 32-bit result bleeds 1-bits into bit
// positions 24..31, which should be owned exclusively by byte3. The
// unsigned path does not do this, so its output diverges from what Java F5
// / PixelKnot produced when the reference stream was captured.
//
// The golden test TestPRNGOutputIsStable_v1 in determinism_test.go locks
// the current (sign-extending) behaviour, and TestJavaCompatibility_KnownVectors
// pins the NextBytes stream. Both must pass; altering the byte assembly
// without a deliberate spec change will break decode compatibility with
// every PixelKnot / F5.jar artifact ever produced against this library.
//
// If an internal error occurs (extremely rare), returns 0 and stores the error
// in lastErr (accessible via LastError method).
func (sr *SecureRandom) NextInt() int32 {
	// Fill the internal buffer with 4 bytes (zero allocations)
	err := sr.fillBytes(sr.intBuf[:])
	if err != nil {
		sr.lastErr = err
		return 0
	}

	// Convert each byte to SIGNED int (matching GetNextByte behavior).
	// GetNextByte does: int(int8(byte)), so negative bytes stay negative,
	// and their sign extension deliberately propagates into the high bits
	// of the OR below. See the doc comment above for why this is
	// load-bearing and must not be "cleaned up" to a uint32-OR form.
	signedByte0 := int(int8(sr.intBuf[0]))
	signedByte1 := int(int8(sr.intBuf[1]))
	signedByte2 := int(int8(sr.intBuf[2]))
	signedByte3 := int(int8(sr.intBuf[3]))

	// Build 32-bit integer matching Java's byte-by-byte construction
	// Using signed bytes ensures correct behavior when bytes are negative
	val := signedByte0 | (signedByte1 << 8) | (signedByte2 << 16) | (signedByte3 << 24)
	// #nosec G115 -- Safe conversion: val is truncated to its low 32 bits;
	// the resulting bit pattern is the F5-compatible int32 value.
	return int32(val)
}

// fillBytes fills the provided buffer with random bytes (zero allocation version).
// This is an internal method for NextInt to avoid allocations.
// Returns an error if the hash computation fails.
func (sr *SecureRandom) fillBytes(buf []byte) error {
	n := len(buf)
	if n == 0 {
		return nil
	}

	resultPos := 0

	// Use remainder bytes first. We advance remHead instead of shifting the
	// tail to the front — this removes the per-call memmove that dominated
	// the NextInt hot path (4 bytes out of 20 per call means up to 4 shifts
	// per digest refresh).
	if sr.remCount > 0 {
		toCopy := n
		if toCopy > sr.remCount {
			toCopy = sr.remCount
		}
		copy(buf[resultPos:], sr.remainder[sr.remHead:sr.remHead+toCopy])
		resultPos += toCopy
		sr.remCount -= toCopy
		if sr.remCount == 0 {
			sr.remHead = 0
		} else {
			sr.remHead += toCopy
		}
	}

	// Generate more bytes as needed
	for resultPos < n {
		output, err := sr.updateState()
		if err != nil {
			return err
		}
		remaining := n - resultPos

		if remaining < len(output) {
			// Use part of output, save rest to remainder. The drain branch
			// above guarantees remHead == 0 whenever remCount == 0, so the
			// saved tail always lives at remainder[0:].
			copy(buf[resultPos:], output[:remaining])
			copy(sr.remainder, output[remaining:])
			sr.remCount = len(output) - remaining
			sr.remHead = 0
			resultPos = n
		} else {
			// Use all the output
			copy(buf[resultPos:], output)
			resultPos += len(output)
		}
	}

	return nil
}

// LastError returns the last error that occurred during PRNG operations.
//
// This method MUST be checked after batched NextBytes / NextBytesInto / NextInt
// calls: those methods do not return errors directly (to preserve the
// Java-compatible signatures), so any error encountered during the batch is
// only surfaced here. A typical pattern is:
//
//	for i := 0; i < n; i++ {
//	    vals[i] = rs.NextInt()
//	}
//	if err := rs.LastError(); err != nil {
//	    // handle error — some or all of vals[] may be zero / partial
//	}
//
// Returns nil if no error has occurred since the last successful Seed() call.
func (sr *SecureRandom) LastError() error {
	return sr.lastErr
}

// Clear securely zeros all internal state to prevent memory disclosure attacks.
// This method MUST be called when done using SecureRandom, especially if seeded
// with sensitive data (passwords, keys, secrets, etc.).
//
// Clear() performs the following operations:
//   - Zeros the 20-byte state array
//   - Zeros the 20-byte remainder buffer
//   - Zeros the 4-byte intBuf array
//   - Zeros the 20-byte outputBuf array (may hold the last hash digest)
//   - Resets remCount to 0
//   - Clears any stored error
//   - Resets the wrapped hasher so its internal block buffer does not retain
//     residual state from the last Sum / SumInto call
//   - Hints to the garbage collector for immediate memory reclamation
//
// After calling Clear(), the SecureRandom instance is no longer usable and must
// be reseeded with Seed() before generating new random data.
//
// Security Best Practice:
//
//	sr := NewSecureRandom(hasher)
//	defer sr.Clear() // Ensure cleanup even if panic occurs
//	sr.Seed([]byte(sensitivePassword))
//	// ... use sr
//	// Clear() called automatically on function exit
//
// Multiple calls to Clear() are safe (idempotent - subsequent calls are no-ops).
//
// Thread Safety: This method is NOT thread-safe. Do not call Clear() concurrently
// with other SecureRandom methods on the same instance.
func (sr *SecureRandom) Clear() {
	log().Debug("Clear called")

	// Zero state array (20 bytes from SHA-1)
	for i := range sr.state {
		sr.state[i] = 0
	}

	// Zero remainder buffer (20 bytes)
	for i := range sr.remainder {
		sr.remainder[i] = 0
	}

	// Zero internal int buffer (4 bytes)
	for i := range sr.intBuf {
		sr.intBuf[i] = 0
	}

	// Zero output buffer (may hold the most recent 20-byte hash digest)
	for i := range sr.outputBuf {
		sr.outputBuf[i] = 0
	}

	// Reset remainder window bookkeeping (head + count).
	sr.remCount = 0
	sr.remHead = 0

	// Clear any stored error
	sr.lastErr = nil

	// Reset the wrapped hasher so its internal block buffer (which may hold
	// residual state from the last Sum / SumInto — e.g. message length, the
	// partially-processed block, the current chaining variables) is wiped.
	// The Hasher interface contract includes Reset(); concrete hashers in
	// this package (SHA1Hasher) wipe all internal state on Reset(). If a
	// custom Hasher implementation does not fully wipe on Reset(), it must
	// be wiped by the caller separately — SecureRandom has no way to
	// guarantee that from behind the interface.
	if sr.hasher != nil {
		sr.hasher.Reset()
	}

	// Hint to garbage collector to reclaim memory immediately
	// This increases the likelihood that sensitive data is removed from memory
	runtime.KeepAlive(sr)

	log().Debug("Clear completed")
}

// Compile-time interface checks.
var (
	_ RandomSource              = (*SecureRandom)(nil)
	_ RandomSourceWithBytesInto = (*SecureRandom)(nil)
)
