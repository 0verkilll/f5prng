package main

import (
	"fmt"

	"github.com/0verkilll/f5prng"
)

func main() {
	fmt.Println("=== Bounded Random Integers with NextIntN ===")
	fmt.Println()

	// Create and seed PRNG
	hasher := f5prng.NewSHA1Hasher()
	prng := f5prng.NewSecureRandom(hasher)
	defer prng.Clear()

	err := prng.Seed([]byte("bounded-example"))
	if err != nil {
		fmt.Printf("Error seeding: %v\n", err)
		return
	}

	// Example 1: Dice roll (1-6)
	fmt.Println("1. Simulating dice rolls (1-6):")
	fmt.Print("   ")
	for i := 0; i < 10; i++ {
		roll, err := f5prng.NextIntN(prng, 6)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("%d ", roll+1) // Add 1 to get 1-6 instead of 0-5
	}
	fmt.Println()

	// Example 2: Random index (0-99)
	fmt.Println("\n2. Random indices (0-99):")
	fmt.Print("   ")
	for i := 0; i < 5; i++ {
		idx, err := f5prng.NextIntN(prng, 100)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("%d ", idx)
	}
	fmt.Println()

	// Example 3: Array shuffle simulation
	fmt.Println("\n3. Simulating Fisher-Yates shuffle indices for array[10]:")
	for i := 9; i > 0; i-- {
		j, err := f5prng.NextIntN(prng, i+1) // Random index from 0 to i
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("   Swap position %d with position %d\n", i, j)
	}

	// Example 4: Error handling
	fmt.Println("\n4. Error handling for invalid bounds:")
	_, err = f5prng.NextIntN(prng, 0)
	if err != nil {
		fmt.Printf("   NextIntN(prng, 0): Error - %v\n", err)
	}

	_, err = f5prng.NextIntN(prng, -5)
	if err != nil {
		fmt.Printf("   NextIntN(prng, -5): Error - %v\n", err)
	}

	fmt.Println()
	fmt.Println("=== Usage Notes ===")
	fmt.Println("NextIntN(prng, n) returns a value in [0, n)")
	fmt.Println("For range [a, b], use: a + NextIntN(prng, b-a+1)")
}
