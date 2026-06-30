package main

import (
	"fmt"
	"os"

	"github.com/0verkilll/f5prng"
)

func main() {
	// Show supported locales
	fmt.Println("=== f5prng Internationalization Example ===")
	fmt.Println()
	fmt.Printf("Supported locales: %v\n", f5prng.GetSupportedLocales())
	fmt.Println()

	// Test with a few locales
	locales := []string{"en-US", "es-ES", "fr-FR"}

	for _, locale := range locales {
		fmt.Printf("=== Locale: %s ===\n", locale)

		// Create and set translator for this locale
		translator, err := f5prng.NewTranslator(locale)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error creating translator: %v\n", err)
			continue
		}
		f5prng.SetTranslator(translator)

		// Trigger an error to see localized message
		// Try NextIntN with invalid bound
		_, err = f5prng.NextIntN(nil, 0)
		if err != nil {
			fmt.Printf("Invalid bound: %v\n", err)
		}

		// Try NextIntN with negative bound
		hasher := f5prng.NewSHA1Hasher()
		prng := f5prng.NewSecureRandom(hasher)
		_, err = f5prng.NextIntN(prng, -1)
		if err != nil {
			fmt.Printf("Negative bound: %v\n", err)
		}
		prng.Clear()

		fmt.Println()
	}

	// Auto-detect locale (uses system locale)
	fmt.Println("=== Auto-detect locale ===")
	translator, err := f5prng.NewTranslator("")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	f5prng.SetTranslator(translator)
	fmt.Println("Translator configured with system locale")
	fmt.Printf("Detected locale: %s\n", translator.GetLocale())

	// Reset to nil to show default behavior
	fmt.Println()
	fmt.Println("=== Default behavior (no translator) ===")
	f5prng.SetTranslator(nil)

	_, err = f5prng.NextIntN(nil, 0)
	if err != nil {
		fmt.Printf("Error message: %v\n", err)
	}
	fmt.Println("(Default English message when no translator is set)")
}
