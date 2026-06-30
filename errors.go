package f5prng

// ErrorCode represents an error code for i18n support.
// Error codes enable localized error messages and programmatic error handling.
type ErrorCode int

const (
	// ErrCodeInvalidBound indicates n <= 0 was passed to NextIntN.
	ErrCodeInvalidBound ErrorCode = iota + 1

	// ErrCodeSeedHashFailed indicates the hash computation failed during Seed().
	// This is extremely rare and would require a seed exceeding 2^61-1 bytes.
	ErrCodeSeedHashFailed

	// ErrCodeStateUpdateFailed indicates the hash computation failed during state update.
	// This should never happen with proper hasher implementations.
	ErrCodeStateUpdateFailed

	// ErrCodeNilHasher indicates a nil hasher was provided to NewSecureRandom.
	ErrCodeNilHasher
)

// errorTranslationKeys maps error codes to their i18n translation keys.
var errorTranslationKeys = map[ErrorCode]string{
	ErrCodeInvalidBound:      "error.invalid_bound",
	ErrCodeSeedHashFailed:    "error.seed_hash_failed",
	ErrCodeStateUpdateFailed: "error.state_update_failed",
	ErrCodeNilHasher:         "error.nil_hasher",
}

// errorDefaultMessages maps error codes to their default English messages.
var errorDefaultMessages = map[ErrorCode]string{
	ErrCodeInvalidBound:      "bound must be greater than zero",
	ErrCodeSeedHashFailed:    "seed hash computation failed",
	ErrCodeStateUpdateFailed: "state update hash computation failed",
	ErrCodeNilHasher:         "hasher cannot be nil",
}

// Error represents a f5prng error with a code for i18n support.
// The Code field allows callers to map errors to localized messages.
//
// Example:
//
//	if err != nil {
//	    var e f5prng.Error
//	    if errors.As(err, &e) {
//	        switch e.Code {
//	        case f5prng.ErrCodeInvalidBound:
//	            // Handle invalid bound error
//	        case f5prng.ErrCodeSeedHashFailed:
//	            // Handle seed hash failure
//	        }
//	    }
//	}
type Error struct {
	Code    ErrorCode
	Message string
}

// Error implements the error interface.
// If a translator is set via SetTranslator(), returns the localized message.
// Otherwise, returns the default English message.
func (e Error) Error() string {
	// If a custom message was set, use it
	if e.Message != "" {
		return e.Message
	}

	// Try to get the translation key for this error code
	key, hasKey := errorTranslationKeys[e.Code]
	if !hasKey {
		return "f5prng.Error"
	}

	// Get the default message for this error code
	defaultMsg, hasDefault := errorDefaultMessages[e.Code]
	if !hasDefault {
		defaultMsg = "f5prng.Error"
	}

	// Use the translate function which handles translator presence
	return translate(key, defaultMsg)
}

// Sentinel errors for common error conditions.
// These can be used with errors.Is() for error checking.
var (
	// ErrInvalidBound is returned when NextIntN is called with n <= 0.
	ErrInvalidBound = Error{Code: ErrCodeInvalidBound}

	// ErrSeedHashFailed is returned when the hash computation fails during Seed().
	ErrSeedHashFailed = Error{Code: ErrCodeSeedHashFailed}

	// ErrStateUpdateFailed is returned when the hash computation fails during state update.
	ErrStateUpdateFailed = Error{Code: ErrCodeStateUpdateFailed}

	// ErrNilHasher is returned when a nil hasher is provided to NewSecureRandom.
	ErrNilHasher = Error{Code: ErrCodeNilHasher}
)
