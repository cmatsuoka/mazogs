package graphics

import "testing"

func resetInputStateForTest(t *testing.T) {
	t.Helper()
	keyPressed = false
	keyValue = ""
	keyLatch = ""
	t.Cleanup(func() {
		keyPressed = false
		keyValue = ""
		keyLatch = ""
	})
}

// TestInKeyReturnsHeldKeyWithoutConsumingLatch checks that InKey prefers the
// currently held key and leaves the latch untouched.
func TestInKeyReturnsHeldKeyWithoutConsumingLatch(t *testing.T) {
	resetInputStateForTest(t)
	keyValue = "a"
	keyLatch = "b"

	if got := InKey(); got != "a" {
		t.Fatalf("InKey() with held key = %q, want %q", got, "a")
	}
	if keyLatch != "b" {
		t.Fatalf("keyLatch was consumed unexpectedly: got %q want %q", keyLatch, "b")
	}
}

// TestInKeyConsumesLatchWhenNoHeldKey checks that InKey returns and clears the
// latch when there is no currently held key.
func TestInKeyConsumesLatchWhenNoHeldKey(t *testing.T) {
	resetInputStateForTest(t)
	keyLatch = "x"

	if got := InKey(); got != "x" {
		t.Fatalf("InKey() from latch = %q, want %q", got, "x")
	}
	if keyLatch != "" {
		t.Fatalf("keyLatch not cleared after InKey: got %q want empty", keyLatch)
	}
	if got := InKey(); got != "" {
		t.Fatalf("second InKey() after consuming latch = %q, want empty", got)
	}
}

// TestClearLatchPreservesHeldKey checks that ClearLatch only clears the latch
// and keeps current held-key state intact.
func TestClearLatchPreservesHeldKey(t *testing.T) {
	resetInputStateForTest(t)
	keyValue = "left"
	keyLatch = "right"

	ClearLatch()

	if keyValue != "left" {
		t.Fatalf("keyValue changed unexpectedly: got %q want %q", keyValue, "left")
	}
	if keyLatch != "" {
		t.Fatalf("keyLatch not cleared: got %q want empty", keyLatch)
	}
	if !HasKey() {
		t.Fatalf("HasKey() = false, want true with held key present")
	}
}

// TestClearKeysClearsAllInputState checks that ClearKeys clears the pressed,
// held, and latched key state.
func TestClearKeysClearsAllInputState(t *testing.T) {
	resetInputStateForTest(t)
	keyPressed = true
	keyValue = "up"
	keyLatch = "up"

	ClearKeys()

	if keyPressed {
		t.Fatalf("keyPressed not cleared: got true want false")
	}
	if keyValue != "" {
		t.Fatalf("keyValue not cleared: got %q want empty", keyValue)
	}
	if keyLatch != "" {
		t.Fatalf("keyLatch not cleared: got %q want empty", keyLatch)
	}
	if HasKey() {
		t.Fatalf("HasKey() = true, want false after ClearKeys")
	}
}

// TestConvertCodeUsesLookupTable checks that convertCode performs a direct
// translation through the ZX lookup table.
func TestConvertCodeUsesLookupTable(t *testing.T) {
	for i := 0; i < len(zxCharCode); i++ {
		c := byte(i)
		if got, want := convertCode(c), zxCharCode[c]; got != want {
			t.Fatalf("convertCode(%#x) = %#x, want %#x", c, got, want)
		}
	}
}
