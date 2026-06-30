package f5prng

import (
	"reflect"
	"testing"
)

// TestPRNGOutputIsStable_v1 is a cross-version determinism regression lock.
//
// F5 / SHA1PRNG consumers rely on *byte-exact* reproducibility of the PRNG
// stream for a given seed — PixelKnot and F5.jar emit steganographic payloads
// whose decoding depends on producing the same permutation indices / byte
// masks as the original Java code. Any accidental change to the PRNG output
// (e.g. a refactor that replaces signed-byte arithmetic with unsigned but
// gets the sign-extension wrong, or re-orders the byte assembly in NextInt)
// silently breaks decode compatibility with every image ever embedded.
//
// This test captures the observed output of the current implementation for
// one fixed seed and freezes it. If you change PRNG behaviour for ANY reason,
// you MUST justify it in a spec change and regenerate `want` deliberately.
// Do NOT update `want` just to make a red test go green.
//
// The `want` values below were captured on 2026-04-23 against the current
// sign-extending NextInt assembly (see the "Why sign-extended ... is
// load-bearing" note in securerandom.go). An audit attempt to refactor the
// NextInt byte assembly to a clean unsigned-OR form was discovered to
// change these numbers — the refactor was reverted precisely because this
// test caught it. Do not "optimise" NextInt without first updating these
// values deliberately as part of a spec change.
func TestPRNGOutputIsStable_v1(t *testing.T) {
	// Use the concrete SecureRandom so we can call LastError() — it isn't
	// part of the RandomSource interface (by design; error recovery is an
	// implementation detail), but is available on the concrete type.
	sr := NewSecureRandom(NewSHA1Hasher()).(*SecureRandom)
	defer sr.Clear()

	if err := sr.Seed([]byte("deterministic-seed-for-regression-v1")); err != nil {
		t.Fatalf("Seed returned unexpected error: %v", err)
	}

	got := make([]int32, 16)
	for i := range got {
		got[i] = sr.NextInt()
	}

	// FROZEN: captured output of SecureRandom seeded with
	// "deterministic-seed-for-regression-v1" as of 2026-04-23. If this test
	// fails, you changed PRNG output. Do NOT update these numbers without
	// a deliberate, reviewed spec change — Java SHA1PRNG parity depends on
	// them being stable.
	want := []int32{
		811078780,
		-10157,
		-7823,
		-6011829,
		-15778,
		-18925,
		941579530,
		-3203750,
		-2158767,
		-11236,
		-2806452,
		-88,
		-14,
		-100306675,
		-55,
		-10209,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PRNG output diverged from frozen reference.\n  got:  %v\n  want: %v\nIf this change is intentional you MUST document why in a spec change and update the want[] slice deliberately.", got, want)
	}

	if err := sr.LastError(); err != nil {
		t.Errorf("LastError after regression stream: %v", err)
	}
}
