package main

import (
	"encoding/hex"
	"fmt"

	"github.com/0verkilll/f5prng"
)

func main() {
	// Create the SHA-1 hasher and PRNG
	hasher := f5prng.NewSHA1Hasher()
	prng := f5prng.NewSecureRandom(hasher)
	defer prng.Clear() // Always clean up sensitive data

	// Seed with a password (like F5 steganography)
	password := "secret123"
	err := prng.Seed([]byte(password))
	if err != nil {
		fmt.Printf("Error seeding: %v\n", err)
		return
	}

	// Generate random bytes
	randomBytes := prng.NextBytes(20)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Random bytes (20): %s\n", hex.EncodeToString(randomBytes))

	// Generate random integers
	fmt.Println("\nRandom integers:")
	for i := 0; i < 5; i++ {
		value := prng.NextInt()
		fmt.Printf("  NextInt(): %d\n", value)
	}

	fmt.Println()
	fmt.Println("This output is deterministic:")
	fmt.Println("Same password always produces the same sequence.")
}
