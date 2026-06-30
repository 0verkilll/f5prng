# Bounded Example

Using `NextIntN` for bounded random integers in the range [0, n).

## What This Demonstrates

- Generating random integers within a specific range
- Common use cases: dice rolls, array indices, shuffling
- Proper error handling for invalid bounds
- Converting [0, n) to other ranges like [a, b]

## Running

```bash
go run main.go
```

## Expected Output

```
=== Bounded Random Integers with NextIntN ===

1. Simulating dice rolls (1-6):
   4 2 6 1 3 5 2 4 1 6

2. Random indices (0-99):
   42 17 89 3 56

3. Simulating Fisher-Yates shuffle indices for array[10]:
   Swap position 9 with position 3
   Swap position 8 with position 7
   ...

4. Error handling for invalid bounds:
   NextIntN(prng, 0): Error - bound must be greater than zero
   NextIntN(prng, -5): Error - bound must be greater than zero

=== Usage Notes ===
NextIntN(prng, n) returns a value in [0, n)
For range [a, b], use: a + NextIntN(prng, b-a+1)
```

## Code Walkthrough

```go
// Create and seed PRNG
hasher := f5prng.NewSHA1Hasher()
prng := f5prng.NewSecureRandom(hasher)
defer prng.Clear()

err := prng.Seed([]byte("seed"))

// Dice roll (1-6): Add 1 to convert [0,6) to [1,6]
roll, err := f5prng.NextIntN(prng, 6)
diceValue := roll + 1

// Random array index (0-99)
idx, err := f5prng.NextIntN(prng, 100)

// Range [a, b]: Use a + NextIntN(prng, b-a+1)
// Example: [10, 20]
value := 10 + NextIntN(prng, 11)
```

## Key Concepts

**Range [0, n):** `NextIntN(prng, n)` returns a value from 0 (inclusive) to n (exclusive). This is the standard convention matching Java and most programming languages.

**Error Handling:** `NextIntN` returns an error for invalid bounds (n <= 0). Always check the error return.

**Fisher-Yates Usage:** The bounded random function is essential for implementing the Fisher-Yates shuffle algorithm, which is the core of F5 steganography's permutation generation.

**Uniform Distribution:** The implementation handles modulo bias to ensure uniform distribution across the range.
