package f5prng

import (
	"github.com/0verkilll/sha1"
)

// DefaultFactory is the default PRNGFactory implementation that creates
// SecureRandom instances using SHA-1 hashing for Java compatibility.
//
// This factory creates instances that are compatible with Java's
// SecureRandom("SHA1PRNG") implementation, making it suitable for
// F5 steganography operations.
type DefaultFactory struct{}

// NewDefaultFactory creates a new DefaultFactory instance.
//
// Example:
//
//	factory := NewDefaultFactory()
//	prng := factory.NewPRNG()
//	defer prng.Clear()
//	if err := prng.Seed([]byte("password")); err != nil {
//	    // handle error
//	}
//	bytes := prng.NextBytes(20)
func NewDefaultFactory() PRNGFactory {
	return &DefaultFactory{}
}

// NewPRNG creates and returns a new SecureRandom instance.
// The returned instance uses SHA-1 hashing and is compatible with
// Java's SecureRandom("SHA1PRNG").
//
// The returned RandomSource must be seeded before use.
//
// Example:
//
//	factory := NewDefaultFactory()
//	prng := factory.NewPRNG()
//	defer prng.Clear()
//	if err := prng.Seed([]byte("password")); err != nil {
//	    // handle error
//	}
//	randomInt := prng.NextInt()
func (f *DefaultFactory) NewPRNG() RandomSource {
	hasher := sha1.NewSHA1(sha1.NewBigEndian())
	return NewSecureRandom(hasher)
}

// Verify that DefaultFactory implements PRNGFactory at compile time
var _ PRNGFactory = (*DefaultFactory)(nil)
