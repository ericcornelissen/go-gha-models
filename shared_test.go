// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"testing"
)

func checkMap(t *testing.T, got, want map[string]string) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of items in map (got %d, want %d)", got, want)
		return
	}

	for k, got := range got {
		want, ok := want[k]
		if !ok {
			t.Errorf("Got key %q in map, but do want it", k)
			continue
		}

		if got != want {
			t.Errorf("Unexpected value for key %q in map (got %q, want %q)", k, got, want)
		}
	}

	for k, want := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("Want key %q(=%q) in map, but it is not present", k, want)
		}
	}
}
