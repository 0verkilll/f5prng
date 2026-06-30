# Deterministic Example

Demonstrates that the PRNG produces byte-identical output for the same seed.

## What This Demonstrates

- Same seed always produces identical output
- Different seeds produce different output
- Independent PRNG instances with same seed produce same sequence
- Why determinism is critical for F5 steganography

## Running

```bash
go run main.go
```

## Expected Output

```
=== Determinism Demonstration ===

1. First generation with password: test-password
   Bytes: 7c9c6c9e...

2. Second generation with SAME password: test-password
   Bytes: 7c9c6c9e...

3. Verification:
   PASS: Both sequences are IDENTICAL

4. Third generation with DIFFERENT password: different-password
   Bytes: a1b2c3d4...

5. Verification:
   PASS: Different password produces DIFFERENT output

=== Why This Matters ===
F5 steganography requires IDENTICAL permutations during
encoding and decoding. This determinism ensures that
the same password produces the exact same PRNG sequence.
```

## Code Walkthrough

```go
func generateSequence(password string) []byte {
    // Create fresh PRNG instance
    hasher := f5prng.NewSHA1Hasher()
    prng := f5prng.NewSecureRandom(hasher)
    defer prng.Clear()

    // Seed with password
    err := prng.Seed([]byte(password))
    if err != nil {
        return nil
    }

    // Generate bytes - this is deterministic
    return prng.NextBytes(32)
}

// Calling generateSequence("password") multiple times
// ALWAYS returns the exact same byte sequence
```

## Key Concepts

**Reproducibility:** The core requirement for F5 steganography is that both the encoder and decoder generate identical permutation sequences. This example proves that independent PRNG instances with the same seed produce identical output.

**Java Compatibility:** The output is byte-identical to Java's `SecureRandom` with SHA1PRNG algorithm. You can verify this by running the same seed through both implementations.

**State Independence:** Each call to `generateSequence` creates a new PRNG with fresh state. Despite being independent instances, they produce identical output when seeded with the same value.
