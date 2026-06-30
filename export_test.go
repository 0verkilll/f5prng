package f5prng

// export_test.go exposes private methods for testing purposes only.
// This file follows the standard Go testing pattern for testing unexported
// methods without making them public in the production API.

// TestFillBytes exposes the private fillBytes method for testing.
// This allows us to test the "use all output" branch which is unreachable
// in normal usage since fillBytes is only called from NextInt with a 4-byte buffer.
// Note: This ignores the error for backward compatibility in tests that don't check errors.
func (sr *SecureRandom) TestFillBytes(buf []byte) {
	_ = sr.fillBytes(buf)
}
