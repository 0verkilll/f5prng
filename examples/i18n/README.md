# Internationalization Example

Demonstrates localized error messages in multiple languages.

## What This Demonstrates

- Setting up translators for different locales
- Viewing error messages in different languages
- Auto-detecting system locale
- Fallback to English when no translator is set

## Running

```bash
go run main.go
```

## Expected Output

```
=== f5prng Internationalization Example ===

Supported locales: [en-US es-ES fr-FR]

=== Locale: en-US ===
Invalid bound: bound must be greater than zero
Negative bound: bound must be greater than zero

=== Locale: es-ES ===
Invalid bound: el limite debe ser mayor que cero
Negative bound: el limite debe ser mayor que cero

=== Locale: fr-FR ===
Invalid bound: la limite doit etre superieure a zero
Negative bound: la limite doit etre superieure a zero

=== Auto-detect locale ===
Translator configured with system locale
Detected locale: en-US

=== Default behavior (no translator) ===
Error message: bound must be greater than zero
(Default English message when no translator is set)
```

## Code Walkthrough

```go
// Get list of supported locales
locales := f5prng.GetSupportedLocales()

// Create translator for a specific locale
translator, err := f5prng.NewTranslator("es-ES")
if err != nil {
    // Handle error
}

// Set the global translator
f5prng.SetTranslator(translator)

// Now errors will be in Spanish
_, err = f5prng.NextIntN(prng, 0)
// err.Error() returns Spanish message

// Auto-detect from system environment
translator, _ = f5prng.NewTranslator("")
f5prng.SetTranslator(translator)

// Reset to default (English fallback)
f5prng.SetTranslator(nil)
```

## Key Concepts

**Global Translator:** The translator is set globally using `SetTranslator()`. All error messages from the package will use this translator.

**Thread Safety:** The translator uses `sync.RWMutex` internally, making it safe to use from multiple goroutines.

**Fallback Behavior:** If no translator is set (nil), error messages default to English. This ensures the package works without requiring i18n setup.

**Locale Detection:** Passing an empty string to `NewTranslator("")` auto-detects the locale from environment variables (`LANG`, `LANGUAGE`, etc.).

## Adding New Locales

To add support for a new language:

1. Create a JSON file in the `locales/` directory (e.g., `de-DE.json`)
2. Copy the structure from `en-US.json`
3. Translate all message values
4. The new locale is automatically available via `GetSupportedLocales()`
