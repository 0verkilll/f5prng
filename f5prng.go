// Package f5prng provides unified PRNG interfaces for F5 steganography packages.
//
// This package serves as the single source of truth for PRNG interfaces across
// the F5 steganography ecosystem, eliminating interface fragmentation and the
// need for adapter types in consuming packages.
//
// The interfaces are designed to match Java's SecureRandom behavior for
// compatibility with PixelKnot's F5 steganography algorithm.
//
// # Quick Start
//
//	factory := f5prng.NewDefaultFactory()
//	prng := factory.NewPRNG()
//	defer prng.Clear()
//
//	if err := prng.Seed([]byte("password")); err != nil {
//	    // handle error
//	}
//	randomBytes := prng.NextBytes(20)
//	randomInt := prng.NextInt()
//	boundedInt, _ := f5prng.NextIntN(prng, 100)
//
// # Security Warning
//
// The implementations in this package are NOT cryptographically secure.
// Do not use for cryptographic key generation, session tokens, or other
// security-critical purposes. Use crypto/rand for secure randomness.
//
// # Java Compatibility
//
// This package produces byte-identical output to Java's SecureRandom (SHA1PRNG)
// for compatibility with:
//   - PixelKnot Android app
//   - F5.jar reference implementation
//   - Other Java F5 steganography tools
package f5prng

// NewDefaultHasher creates a new SHA-1 hasher for use with SecureRandom.
// This is a convenience function equivalent to NewSHA1Hasher().
func NewDefaultHasher() Hasher {
	return NewSHA1Hasher()
}
