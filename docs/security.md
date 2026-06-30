# Security

This document explains the security characteristics of f5prng, including warnings about its limitations and guidance for safe usage.

## Security Warning

**f5prng is NOT cryptographically secure.** This implementation should NOT be used for:

- Generating cryptographic keys
- Creating session tokens or authentication credentials
- Password generation
- Any security-critical random number generation
- Nonces for cryptographic protocols
- IV/salt generation for encryption

## Why It's Not Secure

### SHA-1 is Broken

SHA-1 has known cryptographic weaknesses:

- **Collision attacks**: Practical collision attacks exist (SHAttered, 2017)
- **Preimage resistance**: Weakening, though not fully broken
- **Industry deprecation**: SHA-1 is deprecated for security use since 2017

### Deterministic Output

The SHA1PRNG algorithm is **fully deterministic**:

- Same seed always produces the same output
- Output sequence is predictable if the seed is known or guessable
- No entropy mixing from system sources

### Java Compatibility vs Security

This implementation prioritizes **exact Java compatibility** over security:

- Matches Java's deprecated SHA1PRNG algorithm
- Uses signed byte arithmetic for bit-exact output
- Cannot add security improvements without breaking compatibility

## Safe Use Cases

f5prng is designed for specific, non-security applications:

### F5 Steganography

The primary use case is compatibility with F5 steganography tools:

```go
// Safe: Generating permutation for F5 steganography
factory := f5prng.NewDefaultFactory()
prng := factory.NewPRNG()
defer prng.Clear()

if err := prng.Seed([]byte(password)); err != nil {
    // handle error
}
perm := fisheryates.Generate(size, prng)
```

This is safe because:
- The security model relies on the password, not the PRNG
- F5 requires exact permutation matching with Java implementations
- The randomness quality doesn't affect message confidentiality

### Legacy Java Interoperability

When you must match Java's `SecureRandom` output:

```go
// Safe: Cross-platform deterministic behavior
if err := prng.Seed([]byte("shared-seed")); err != nil {
    // handle error
}
bytes := prng.NextBytes(20)
// Same bytes as Java's SecureRandom.nextBytes(20)
```

### Research and Education

For studying F5 steganography or Java compatibility:

```go
// Safe: Academic research on F5 algorithm
if err := prng.Seed(testVector); err != nil {
    // handle error
}
// Analyze output patterns
```

### Deterministic Testing

For reproducible test scenarios:

```go
// Safe: Reproducible test setup
if err := prng.Seed([]byte("test-seed")); err != nil {
    // handle error
}
testData := prng.NextBytes(100)
// Same data every test run
```

## What to Use Instead

For cryptographic applications, use Go's standard library:

### Secure Random Bytes

```go
import "crypto/rand"

// Generate 32 secure random bytes
bytes := make([]byte, 32)
_, err := rand.Read(bytes)
if err != nil {
    panic(err)
}
```

### Secure Random Integers

```go
import "crypto/rand"
import "math/big"

// Generate secure random integer in [0, max)
max := big.NewInt(100)
n, err := rand.Int(rand.Reader, max)
if err != nil {
    panic(err)
}
```

### Secure Token Generation

```go
import (
    "crypto/rand"
    "encoding/base64"
)

// Generate secure token
bytes := make([]byte, 32)
rand.Read(bytes)
token := base64.URLEncoding.EncodeToString(bytes)
```

## Memory Protection

Even for non-security uses, protect sensitive data:

### Always Clear After Use

```go
prng := f5prng.NewSecureRandom(hasher)
defer prng.Clear() // Zeros all internal state

prng.Seed([]byte(password))
// ... use prng
// Clear() called automatically on function exit
```

### Why Clear Matters

- Passwords may be used as seeds
- Memory dumps could expose seed material
- Process crashes may leave state in memory
- `Clear()` zeros all internal buffers

### Clear Implementation

`Clear()` performs:
1. Zeros the 20-byte state array
2. Zeros the 20-byte remainder buffer
3. Zeros the 4-byte intBuf array
4. Resets remCount to 0
5. Hints to garbage collector for immediate reclamation

## Thread Safety

**SecureRandom is NOT thread-safe.** Do not share instances between goroutines.

### Safe Pattern

```go
// Option 1: One instance per goroutine
func worker(seed []byte) {
    prng := factory.NewPRNG()
    defer prng.Clear()
    prng.Seed(seed)
    // ... use prng locally
}

// Option 2: External synchronization
var mu sync.Mutex

func getRandomBytes(n int) []byte {
    mu.Lock()
    defer mu.Unlock()
    return sharedPRNG.NextBytes(n)
}
```

## Error Handling

f5prng provides proper error handling for edge cases:

### Seed Errors

```go
err := prng.Seed(seed)
if err != nil {
    // Handle error - extremely rare, requires massive seed
    log.Printf("Seed failed: %v", err)
}
```

### LastError for NextBytes/NextInt

```go
bytes := prng.NextBytes(20)
if err := prng.LastError(); err != nil {
    // Handle error
    log.Printf("NextBytes failed: %v", err)
}
```

## Input Validation

### NextIntN Bound Checking

```go
// Safe: Validates bound > 0
result, err := f5prng.NextIntN(prng, 100)
if err != nil {
    // Handle invalid bound
}
```

## Security Checklist

When using f5prng, verify:

- [ ] NOT using for cryptographic purposes
- [ ] Using `defer prng.Clear()` for cleanup
- [ ] NOT sharing instances between goroutines
- [ ] Checking errors from `Seed()` and `LastError()`
- [ ] Understanding output is deterministic and predictable

## Reporting Security Issues

For security concerns: security@example.com

See [SECURITY.md](../.github/SECURITY.md) for full security policy.

## References

- [SHA-1 Collision (SHAttered)](https://shattered.io/)
- [NIST SP 800-131A](https://csrc.nist.gov/publications/detail/sp/800-131a/rev-2/final) - SHA-1 deprecation
- [Go crypto/rand](https://pkg.go.dev/crypto/rand) - Secure random generation
