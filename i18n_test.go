package f5prng

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestNewTranslator(t *testing.T) {
	t.Run("creates translator with specific locale", func(t *testing.T) {
		translator, err := NewTranslator("en-US")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		if translator == nil {
			t.Fatal("NewTranslator() returned nil translator")
		}
		if got := translator.GetLocale(); got != "en-US" {
			t.Errorf("GetLocale() = %v, want %v", got, "en-US")
		}
	})

	t.Run("creates translator with empty locale for auto-detect", func(t *testing.T) {
		translator, err := NewTranslator("")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		if translator == nil {
			t.Fatal("NewTranslator() returned nil translator")
		}
	})

	t.Run("creates translator with Spanish locale", func(t *testing.T) {
		translator, err := NewTranslator("es-ES")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		if got := translator.GetLocale(); got != "es-ES" {
			t.Errorf("GetLocale() = %v, want %v", got, "es-ES")
		}
	})

	t.Run("creates translator with French locale", func(t *testing.T) {
		translator, err := NewTranslator("fr-FR")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		if got := translator.GetLocale(); got != "fr-FR" {
			t.Errorf("GetLocale() = %v, want %v", got, "fr-FR")
		}
	})
}

func TestSetTranslatorAndGetTranslator(t *testing.T) {
	// Save original translator and restore after test
	originalTranslator := GetTranslator()
	defer SetTranslator(originalTranslator)

	t.Run("sets and gets translator", func(t *testing.T) {
		translator, err := NewTranslator("en-US")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}

		SetTranslator(translator)
		got := GetTranslator()

		if got != translator {
			t.Errorf("GetTranslator() did not return the set translator")
		}
	})

	t.Run("sets nil translator", func(t *testing.T) {
		SetTranslator(nil)
		got := GetTranslator()

		if got != nil {
			t.Errorf("GetTranslator() = %v, want nil", got)
		}
	})
}

func TestGetSupportedLocales(t *testing.T) {
	t.Run("returns embedded locales", func(t *testing.T) {
		locales := GetSupportedLocales()

		if len(locales) == 0 {
			t.Fatal("GetSupportedLocales() returned empty slice")
		}

		// Check that the expected locales are present
		expected := []string{"en-US", "es-ES", "fr-FR"}
		for _, exp := range expected {
			found := false
			for _, loc := range locales {
				if loc == exp {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GetSupportedLocales() missing %v, got %v", exp, locales)
			}
		}
	})

	t.Run("returns sorted locales", func(t *testing.T) {
		locales := GetSupportedLocales()

		for i := 1; i < len(locales); i++ {
			if locales[i-1] > locales[i] {
				t.Errorf("GetSupportedLocales() not sorted: %v before %v", locales[i-1], locales[i])
			}
		}
	})
}

func TestGetSupportedLocalesFromFS(t *testing.T) {
	t.Run("handles empty filesystem", func(t *testing.T) {
		emptyFS := fstest.MapFS{}
		locales := getSupportedLocalesFromFS(emptyFS, "locales")

		// Should return fallback
		if len(locales) != 1 || locales[0] != "en-US" {
			t.Errorf("getSupportedLocalesFromFS() = %v, want [en-US]", locales)
		}
	})

	t.Run("handles missing directory", func(t *testing.T) {
		emptyFS := fstest.MapFS{}
		locales := getSupportedLocalesFromFS(emptyFS, "nonexistent")

		// Should return fallback
		if len(locales) != 1 || locales[0] != "en-US" {
			t.Errorf("getSupportedLocalesFromFS() = %v, want [en-US]", locales)
		}
	})

	t.Run("ignores non-json files", func(t *testing.T) {
		testFS := fstest.MapFS{
			"locales/en-US.json": &fstest.MapFile{Data: []byte("{}")},
			"locales/readme.txt": &fstest.MapFile{Data: []byte("readme")},
			"locales/.hidden":    &fstest.MapFile{Data: []byte("hidden")},
		}
		locales := getSupportedLocalesFromFS(testFS, "locales")

		if len(locales) != 1 || locales[0] != "en-US" {
			t.Errorf("getSupportedLocalesFromFS() = %v, want [en-US]", locales)
		}
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		testFS := fstest.MapFS{
			"locales/en-US.json":    &fstest.MapFile{Data: []byte("{}")},
			"locales/subdir/x.json": &fstest.MapFile{Data: []byte("{}")},
		}
		locales := getSupportedLocalesFromFS(testFS, "locales")

		if len(locales) != 1 || locales[0] != "en-US" {
			t.Errorf("getSupportedLocalesFromFS() = %v, want [en-US]", locales)
		}
	})

	t.Run("handles filesystem read error", func(t *testing.T) {
		errorFS := &errorReadDirFS{}
		locales := getSupportedLocalesFromFS(errorFS, "locales")

		// Should return fallback
		if len(locales) != 1 || locales[0] != "en-US" {
			t.Errorf("getSupportedLocalesFromFS() = %v, want [en-US]", locales)
		}
	})
}

// errorReadDirFS is a mock filesystem that returns errors on ReadDir
type errorReadDirFS struct{}

func (errorReadDirFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

func (errorReadDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, fs.ErrNotExist
}

func TestTranslate(t *testing.T) {
	// Save original translator and restore after test
	originalTranslator := GetTranslator()
	defer SetTranslator(originalTranslator)

	t.Run("returns default when no translator set", func(t *testing.T) {
		SetTranslator(nil)
		got := translate("error.invalid_bound", "default message")

		if got != "default message" {
			t.Errorf("translate() = %v, want %v", got, "default message")
		}
	})

	t.Run("returns translated message when translator set", func(t *testing.T) {
		translator, err := NewTranslator("en-US")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := translate("error.invalid_bound", "default message")

		// Should return the translated message from en-US.json
		if got != "bound must be greater than zero" {
			t.Errorf("translate() = %v, want %v", got, "bound must be greater than zero")
		}
	})

	t.Run("returns default for missing key", func(t *testing.T) {
		translator, err := NewTranslator("en-US")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := translate("nonexistent.key", "default message")

		if got != "default message" {
			t.Errorf("translate() = %v, want %v", got, "default message")
		}
	})

	t.Run("returns Spanish translation", func(t *testing.T) {
		translator, err := NewTranslator("es-ES")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := translate("error.invalid_bound", "default message")

		if got != "El límite debe ser mayor que cero" {
			t.Errorf("translate() = %v, want %v", got, "El límite debe ser mayor que cero")
		}
	})

	t.Run("returns French translation", func(t *testing.T) {
		translator, err := NewTranslator("fr-FR")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := translate("error.invalid_bound", "default message")

		if got != "La borne doit être supérieure à zéro" {
			t.Errorf("translate() = %v, want %v", got, "La borne doit être supérieure à zéro")
		}
	})
}

func TestTranslateWithArgs(t *testing.T) {
	// Save original translator and restore after test
	originalTranslator := GetTranslator()
	defer SetTranslator(originalTranslator)

	t.Run("formats default with args when no translator set", func(t *testing.T) {
		SetTranslator(nil)
		got := translateWithArgs("some.key", "value is %d", 42)

		if got != "value is 42" {
			t.Errorf("translateWithArgs() = %v, want %v", got, "value is 42")
		}
	})

	t.Run("formats default for missing key", func(t *testing.T) {
		translator, err := NewTranslator("en-US")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := translateWithArgs("nonexistent.key", "value is %d", 42)

		if got != "value is 42" {
			t.Errorf("translateWithArgs() = %v, want %v", got, "value is 42")
		}
	})
}

func TestErrorTranslation(t *testing.T) {
	// Save original translator and restore after test
	originalTranslator := GetTranslator()
	defer SetTranslator(originalTranslator)

	t.Run("ErrInvalidBound uses default message without translator", func(t *testing.T) {
		SetTranslator(nil)
		got := ErrInvalidBound.Error()

		if got != "bound must be greater than zero" {
			t.Errorf("ErrInvalidBound.Error() = %v, want %v", got, "bound must be greater than zero")
		}
	})

	t.Run("ErrSeedHashFailed uses default message without translator", func(t *testing.T) {
		SetTranslator(nil)
		got := ErrSeedHashFailed.Error()

		if got != "seed hash computation failed" {
			t.Errorf("ErrSeedHashFailed.Error() = %v, want %v", got, "seed hash computation failed")
		}
	})

	t.Run("ErrStateUpdateFailed uses default message without translator", func(t *testing.T) {
		SetTranslator(nil)
		got := ErrStateUpdateFailed.Error()

		if got != "state update hash computation failed" {
			t.Errorf("ErrStateUpdateFailed.Error() = %v, want %v", got, "state update hash computation failed")
		}
	})

	t.Run("ErrNilHasher uses default message without translator", func(t *testing.T) {
		SetTranslator(nil)
		got := ErrNilHasher.Error()

		if got != "hasher cannot be nil" {
			t.Errorf("ErrNilHasher.Error() = %v, want %v", got, "hasher cannot be nil")
		}
	})

	t.Run("ErrInvalidBound uses Spanish translation", func(t *testing.T) {
		translator, err := NewTranslator("es-ES")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := ErrInvalidBound.Error()

		if got != "El límite debe ser mayor que cero" {
			t.Errorf("ErrInvalidBound.Error() = %v, want %v", got, "El límite debe ser mayor que cero")
		}
	})

	t.Run("ErrInvalidBound uses French translation", func(t *testing.T) {
		translator, err := NewTranslator("fr-FR")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		got := ErrInvalidBound.Error()

		if got != "La borne doit être supérieure à zéro" {
			t.Errorf("ErrInvalidBound.Error() = %v, want %v", got, "La borne doit être supérieure à zéro")
		}
	})

	t.Run("custom message overrides translation", func(t *testing.T) {
		translator, err := NewTranslator("es-ES")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}
		SetTranslator(translator)

		customErr := Error{Code: ErrCodeInvalidBound, Message: "custom message"}
		got := customErr.Error()

		if got != "custom message" {
			t.Errorf("Error.Error() = %v, want %v", got, "custom message")
		}
	})

	t.Run("unknown error code returns generic message", func(t *testing.T) {
		SetTranslator(nil)
		unknownErr := Error{Code: ErrorCode(999)}
		got := unknownErr.Error()

		if got != "f5prng.Error" {
			t.Errorf("Error.Error() = %v, want %v", got, "f5prng.Error")
		}
	})
}

func TestTranslatorThreadSafety(t *testing.T) {
	// Save original translator and restore after test
	originalTranslator := GetTranslator()
	defer SetTranslator(originalTranslator)

	t.Run("concurrent access is safe", func(t *testing.T) {
		translator, err := NewTranslator("en-US")
		if err != nil {
			t.Fatalf("NewTranslator() error = %v", err)
		}

		done := make(chan bool)

		// Concurrent setter
		go func() {
			for i := 0; i < 100; i++ {
				SetTranslator(translator)
				SetTranslator(nil)
			}
			done <- true
		}()

		// Concurrent getter
		go func() {
			for i := 0; i < 100; i++ {
				_ = GetTranslator()
			}
			done <- true
		}()

		// Concurrent translator
		go func() {
			for i := 0; i < 100; i++ {
				_ = translate("error.invalid_bound", "default")
			}
			done <- true
		}()

		// Wait for all goroutines
		<-done
		<-done
		<-done
	})
}
