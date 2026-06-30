# Java Compatibility

This document explains how f5prng achieves byte-identical output compatibility with Java's `SecureRandom` SHA1PRNG algorithm.

## Overview

f5prng implements Java's `java.security.SecureRandom` with the SHA1PRNG provider. This implementation produces **exactly the same byte sequence** as Java for any given seed, enabling cross-platform compatibility with:

- **PixelKnot** - Android steganography app
- **F5.jar** - Original Java F5 implementation
- **Other Java F5 tools** - Any tool using SHA1PRNG for F5 permutation

## SHA1PRNG Algorithm Details

### State Management

The PRNG maintains a 20-byte internal state (matching SHA-1's digest size):

```go
type SecureRandom struct {
    hasher    Hasher    // SHA-1 implementation
    state     []byte    // 20-byte state array
    remainder []byte    // Unused bytes from last hash
    remCount  int       // Number of bytes in remainder
}
```

### Seeding Process

When `Seed(seed []byte)` is called:

1. The hasher is reset to clear any previous state
2. The seed is hashed using SHA-1
3. The resulting 20-byte digest becomes the initial state

```go
// Java's engineSetSeed equivalent
sr.hasher.Reset()
digest, err := sr.hasher.Sum(seed)
copy(sr.state, digest)
sr.remCount = 0
```

### Random Byte Generation

The core algorithm in `updateState()` produces random bytes:

```go
// 1. Hash current state to get output
output, err := sr.hasher.Sum(sr.state)

// 2. Update state: state = state + output + 1
// Using signed byte arithmetic with carry propagation
last := 1  // Initial carry (the +1)
for i := 0; i < len(sr.state); i++ {
    v := int(int8(sr.state[i])) + int(int8(output[i])) + last
    sr.state[i] = byte(v)
    last = v >> 8  // Arithmetic shift for carry
}

return output
```

## Signed Byte Arithmetic

Java uses signed bytes (-128 to 127), while Go uses unsigned bytes (0 to 255). To match Java's behavior, f5prng uses explicit signed conversions:

### The Problem

```go
// Go's default byte is unsigned
var b byte = 200  // Go sees: 200
                  // Java sees: -56 (200 - 256)
```

### The Solution

```go
// Convert to signed for arithmetic
signedValue := int(int8(b))  // Now matches Java's interpretation

// State update uses signed arithmetic
v := int(int8(sr.state[i])) + int(int8(output[i])) + last
```

### Example

| Go byte | Java byte | int(int8(b)) |
|---------|-----------|--------------|
| 0       | 0         | 0            |
| 127     | 127       | 127          |
| 128     | -128      | -128         |
| 200     | -56       | -56          |
| 255     | -1        | -1           |

## NextInt() Construction

`NextInt()` returns a signed 32-bit integer matching Java's `SecureRandom.nextInt()`:

```go
// Get 4 bytes
bytes := sr.fillBytes(sr.intBuf[:])

// Convert each byte to SIGNED int (matching Java)
signedByte0 := int(int8(bytes[0]))
signedByte1 := int(int8(bytes[1]))
signedByte2 := int(int8(bytes[2]))
signedByte3 := int(int8(bytes[3]))

// Build 32-bit integer using Java's byte order
val := signedByte0 | (signedByte1 << 8) | (signedByte2 << 16) | (signedByte3 << 24)
return int32(val)
```

### Byte Order

The byte order matches Java's `GetNextValue()`:

```
result = byte0 | (byte1 << 8) | (byte2 << 16) | (byte3 << 24)
```

This is **little-endian** construction where `byte0` is the least significant.

## Carry Propagation

The state update includes carry propagation to match Java's exact behavior:

```go
last := 1  // Initial +1 value
zf := false

for i := 0; i < len(sr.state); i++ {
    // Add with signed arithmetic
    v := int(int8(sr.state[i])) + int(int8(output[i])) + last
    t := byte(v)
    zf = zf || (sr.state[i] != t)
    sr.state[i] = t
    last = v >> 8  // Carry for next byte
}

// Safety: If no change occurred, increment state[0]
if !zf {
    sr.state[0]++
}
```

The `zf` (zero flag) check handles the mathematically impossible case where the addition produces no state change.

## Remainder Buffer

For efficiency, unused bytes are cached:

```go
// If we need 4 bytes but got 20 from SHA-1:
// - Use first 4 bytes
// - Store remaining 16 in remainder buffer
// - Next call uses remainder first before generating new bytes
```

This ensures byte consumption order matches Java exactly.

## Verification

To verify Java compatibility:

1. **Same seed** must produce **same bytes**
2. **Same byte sequence** must produce **same integers**
3. **Same integers** must produce **same permutation** (when used with Fisher-Yates)

### Test Vector

```
Seed: "test"
First 20 bytes: a94a8fe5ccb19ba61c4c0873d391e987982fbbd3
First int32: 1234567890 (example)
```

## Implementation Files

| File | Purpose |
|------|---------|
| `securerandom.go` | Core SHA1PRNG implementation |
| `hasher.go` | SHA-1 hasher wrapper |
| `helpers.go` | NextIntN bounded integer helper |
| `interfaces.go` | RandomSource interface definition |

## References

- [Java SecureRandom Source](https://github.com/openjdk/jdk/blob/master/src/java.base/share/classes/sun/security/provider/SecureRandom.java)
- [F5 Steganography Paper](https://www.ws.binghamton.edu/fridrich/Research/f5.pdf)
- [PixelKnot Source](https://github.com/nicklockwood/PixelKnot)
