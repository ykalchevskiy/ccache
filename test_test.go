package ccache_test

import "testing"

func assert[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func require[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func requireSlice[S ~[]T, T comparable](t *testing.T, want, got S) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("slice: expected %v, got %v", want, got)
	}
	for i := 0; i < len(want); i++ {
		if want[i] != got[i] {
			t.Fatalf("slice[%d]: expected %v, got %v", i, want[i], got[i])
		}
	}
}
