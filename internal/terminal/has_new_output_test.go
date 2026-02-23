package terminal

import (
	"testing"
)

func TestHasNewOutput_DefaultFalse(t *testing.T) {
	w := &Window{}
	if w.HasNewOutput.Load() {
		t.Error("HasNewOutput should be false by default")
	}
}

func TestHasNewOutput_SwapReturnsAndClears(t *testing.T) {
	w := &Window{}
	w.HasNewOutput.Store(true)

	// First swap should return true
	if !w.HasNewOutput.Swap(false) {
		t.Error("first Swap should return true after Store(true)")
	}
	// Second swap should return false (already cleared)
	if w.HasNewOutput.Swap(false) {
		t.Error("second Swap should return false (already cleared)")
	}
}

func TestHasNewOutput_MultipleStores(t *testing.T) {
	w := &Window{}
	// Multiple stores coalesce into a single true
	w.HasNewOutput.Store(true)
	w.HasNewOutput.Store(true)
	w.HasNewOutput.Store(true)

	if !w.HasNewOutput.Swap(false) {
		t.Error("Swap should return true after multiple stores")
	}
	if w.HasNewOutput.Swap(false) {
		t.Error("second Swap should return false")
	}
}
