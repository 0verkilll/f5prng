# Basic Example

Simple PRNG usage demonstrating seeding, NextBytes, and NextInt.

## What This Demonstrates

- Creating the SHA-1 based PRNG
- Seeding with a password
- Generating random bytes with NextBytes
- Generating random integers with NextInt
- Cleaning up with Clear()

## Running

```bash
go run main.go
```

## Expected Output

```
Password: secret123
Random bytes (20): 7c9c6c9e...

Random integers:
  NextInt(): 1234567890
  NextInt(): -987654321
  NextInt(): 123456789
  NextInt(): -123456789
  NextInt(): 987654321

This output is deterministic:
Same password always produces the same sequence.
```

## Code Walkthrough

```go
// 1. Create the SHA-1 hasher and PRNG
hasher := f5prng.NewSHA1Hasher()
prng := f5prng.NewSecureRandom(hasher)
defer prng.Clear() // Always clean up sensitive data

// 2. Seed with a password
password := "secret123"
err := prng.Seed([]byte(password))
if err != nil {
    // Handle error
}

// 3. Generate random bytes
randomBytes := prng.NextBytes(20)

// 4. Generate random integers
value := prng.NextInt()
```

## Key Concepts

**Determinism:** The same seed always produces the same sequence of random values. This is essential for F5 steganography where both encoder and decoder must generate identical permutations.

**Memory Safety:** Always call `Clear()` when done to zero sensitive state. Using `defer prng.Clear()` ensures cleanup even if a panic occurs.

**Java Compatibility:** This PRNG produces byte-identical output to Java's `SecureRandom` with SHA1PRNG algorithm, enabling interoperability with PixelKnot and F5.jar.
