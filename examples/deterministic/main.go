package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/0verkilll/f5prng"
)

func main() {
	fmt.Println("=== Determinism Demonstration ===")
	fmt.Println()

	password := "test-password"

	// Generate first sequence
	fmt.Println("1. First generation with password:", password)
	bytes1 := generateSequence(password)
	fmt.Printf("   Bytes: %s\n", hex.EncodeToString(bytes1))

	// Generate second sequence with SAME password
	fmt.Println("\n2. Second generation with SAME password:", password)
	bytes2 := generateSequence(password)
	fmt.Printf("   Bytes: %s\n", hex.EncodeToString(bytes2))

	// Verify they are identical
	fmt.Println("\n3. Verification:")
	if bytes.Equal(bytes1, bytes2) {
		fmt.Println("   PASS: Both sequences are IDENTICAL")
	} else {
		fmt.Println("   FAIL: Sequences differ!")
	}

	// Generate with DIFFERENT password
	differentPassword := "different-password"
	fmt.Println("\n4. Third generation with DIFFERENT password:", differentPassword)
	bytes3 := generateSequence(differentPassword)
	fmt.Printf("   Bytes: %s\n", hex.EncodeToString(bytes3))

	// Verify they are different
	fmt.Println("\n5. Verification:")
	if !bytes.Equal(bytes1, bytes3) {
		fmt.Println("   PASS: Different password produces DIFFERENT output")
	} else {
		fmt.Println("   FAIL: Outputs should differ!")
	}

	fmt.Println()
	fmt.Println("=== Why This Matters ===")
	fmt.Println("F5 steganography requires IDENTICAL permutations during")
	fmt.Println("encoding and decoding. This determinism ensures that")
	fmt.Println("the same password produces the exact same PRNG sequence.")
}

func generateSequence(password string) []byte {
	hasher := f5prng.NewSHA1Hasher()
	prng := f5prng.NewSecureRandom(hasher)
	defer prng.Clear()

	err := prng.Seed([]byte(password))
	if err != nil {
		fmt.Printf("Error seeding: %v\n", err)
		return nil
	}

	return prng.NextBytes(32)
}
