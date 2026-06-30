# Logging Example

Demonstrates debug logging integration with a custom logger.

## What This Demonstrates

- Silent by default behavior
- Enabling debug logging with SetLogger
- Using structured fields with the logger
- Filtering log levels
- Disabling logging by setting nil

## Running

```bash
go run main.go
```

## Expected Output

```
=== f5prng Logging Example ===

1. Default behavior (silent - no logger set):
   Generated 10 bytes
   (No debug output - this is the default)

2. With debug logging enabled:
   [14:30:45.123] DEBUG: Seed called seed_length=4
   [14:30:45.123] DEBUG: Seed completed
   [14:30:45.123] DEBUG: NextBytes called n=10
   [14:30:45.124] DEBUG: NextBytes completed n=10
   [14:30:45.124] DEBUG: NextInt called
   [14:30:45.124] DEBUG: NextInt completed value=123456789
   Result: 10 bytes generated

3. With structured fields:
   [14:30:45.125] DEBUG: Seed called seed_length=6 component=steganography version=1.0
   [14:30:45.125] DEBUG: Seed completed component=steganography version=1.0
   ...
   Result: 20 bytes generated

4. With level filtering (Warn and above only):
   (Debug messages filtered out)

5. Disable logging:
   (Back to silent mode)
```

## Code Walkthrough

```go
// Enable debug logging with a custom logger
f5prng.SetLogger(&SimpleLogger{level: logger.LevelDebug})

// Use the PRNG - now with debug output
prng.Seed([]byte("test"))
prng.NextBytes(10)
prng.NextInt()

// Add structured fields to the logger
fieldLogger := &SimpleLogger{
    level:  logger.LevelDebug,
    fields: map[string]any{"component": "steganography"},
}
f5prng.SetLogger(fieldLogger)

// Filter to only show warnings and above
f5prng.SetLogger(&SimpleLogger{level: logger.LevelWarn})

// Disable logging (reset to NopLogger)
f5prng.SetLogger(nil)
```

## Key Concepts

**Silent by Default:** The package uses `logger.NopLogger{}` by default, producing no output. This is the expected behavior for library code.

**Logger Interface:** The logger must implement `logger.Logger` interface from `github.com/0verkilll/logger`. You can use zerolog, zap, or any compatible logger.

**Thread Safety:** The logger is stored with `sync.RWMutex` protection, making it safe to change from any goroutine.

**Log Levels:** Debug logs are emitted for all PRNG operations (Seed, NextBytes, NextInt, Clear). Use level filtering to control verbosity.

## Using with zerolog

```go
import (
    "github.com/rs/zerolog"
    "github.com/0verkilll/logger"
)

// Create zerolog adapter
type ZerologAdapter struct {
    log zerolog.Logger
}

func (z *ZerologAdapter) Debug(msg string, args ...any) {
    z.log.Debug().Fields(toMap(args)).Msg(msg)
}
// ... implement other methods

f5prng.SetLogger(&ZerologAdapter{log: zerolog.New(os.Stdout)})
```
