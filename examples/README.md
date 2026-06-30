# Examples

Runnable examples demonstrating f5prng PRNG usage for the F5 steganography ecosystem.

## Quick Run

```bash
# Basic PRNG usage
cd examples/basic && go run main.go

# Determinism demo (same seed = same output)
cd examples/deterministic && go run main.go

# Bounded random integers
cd examples/bounded && go run main.go

# Internationalization
cd examples/i18n && go run main.go

# Debug logging
cd examples/logging && go run main.go
```

## Available Examples

| Example | Description | README |
|---------|-------------|--------|
| [basic/](basic/) | Simple PRNG usage: seeding, NextBytes, NextInt | [README](basic/README.md) |
| [deterministic/](deterministic/) | Demonstrates byte-identical output for same seed | [README](deterministic/README.md) |
| [bounded/](bounded/) | Using NextIntN for bounded random integers | [README](bounded/README.md) |
| [i18n/](i18n/) | Localized error messages | [README](i18n/README.md) |
| [logging/](logging/) | Debug logging with custom logger | [README](logging/README.md) |

## IDE Support

Run configurations are included for:
- **IntelliJ IDEA / GoLand:** `.run/` directory
- **VS Code:** `.vscode/launch.json`

Click the play button in your IDE to run any example.

## Common Pattern

All examples follow the same basic setup:

```go
// 1. Create the PRNG with SHA-1 hasher
hasher := f5prng.NewSHA1Hasher()
prng := f5prng.NewSecureRandom(hasher)
defer prng.Clear() // Always clean up sensitive data

// 2. Seed the PRNG
err := prng.Seed([]byte("your-seed"))
if err != nil {
    // Handle error
}

// 3. Generate random data
bytes := prng.NextBytes(20)   // Get 20 random bytes
value := prng.NextInt()       // Get a random int32
bounded, _ := f5prng.NextIntN(prng, 100) // Get 0-99
```

Each example's README contains detailed explanations and expected output.
