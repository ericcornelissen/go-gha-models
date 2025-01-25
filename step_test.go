// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"strconv"
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestStep(t *testing.T) {
	roundtrip := func(s Step) bool {
		b, err := yaml.Marshal(s)
		if err != nil {
			return true
		}

		if err = yaml.Unmarshal(b, &s); err != nil {
			return false
		}

		return true
	}

	if err := quick.Check(roundtrip, nil); err != nil {
		t.Error(err)
	}
}

func TestUses(t *testing.T) {
	roundtrip := func(u Uses) bool {
		if (u.Name == "" && u.Ref != "") || (u.Name != "" && u.Ref == "") {
			return true
		}

		if _, err := yaml.Marshal(u); err != nil {
			return false
		}

		return true
	}

	if err := quick.Check(roundtrip, nil); err != nil {
		t.Error(err)
	}
}

func checkSteps(t *testing.T, id string, got, want []Step) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of steps in %s (got %d, want %d)", id, got, want)
		return
	}

	for i, got := range got {
		want := want[i]
		sid := strconv.Itoa(i)

		if got, want := got.Name, want.Name; got != want {
			t.Errorf("Unexpected name for step %q in %q (got %q, want %q)", sid, id, got, want)
		} else if got != "" {
			sid = got
		}

		if got, want := got.Env, want.Env; len(got) != len(want) {
			t.Errorf("Unexpected number of items in env for step %q in %q (got %d, want %d)", sid, id, len(got), len(want))
		} else {
			for k, got := range got {
				if want, ok := want[k]; !ok {
					t.Errorf("Unexpected key %s in env for step %d", k, i)
				} else if got != want {
					t.Errorf("Incorrect value for key %s in env for step %q in %q (got %q, want %q)", k, sid, id, got, want)
				}
			}

			for k, want := range want {
				if _, ok := got[k]; !ok {
					t.Errorf("Missing key %s in env for step %q in %q (want %q)", k, sid, id, want)
				}
			}
		}

		if got, want := got.Run, want.Run; got != want {
			t.Errorf("Unexpected run for step %q in %q (got %q, want %q)", sid, id, got, want)
		}

		if got, want := got.Shell, want.Shell; got != want {
			t.Errorf("Unexpected shell for step %q in %q (got %q, want %q)", sid, id, got, want)
		}

		if got, want := got.Uses.Name, want.Uses.Name; got != want {
			t.Errorf("Unexpected uses name for step %q in %q (got %q, want %q)", sid, id, got, want)
		}

		if got, want := got.Uses.Ref, want.Uses.Ref; got != want {
			t.Errorf("Unexpected uses ref for step %q in %q (got %q, want %q)", sid, id, got, want)
		}

		if got, want := got.Uses.Annotation, want.Uses.Annotation; got != want {
			t.Errorf("Unexpected uses annotation for step %q in %q (got %q, want %q)", sid, id, got, want)
		}

		if got, want := got.With, want.With; len(got) != len(want) {
			t.Errorf("Unexpected number of items in with for step %q in %q (got %d, want %d)", sid, id, len(got), len(want))
		} else {
			for k, got := range got {
				if want, ok := want[k]; !ok {
					t.Errorf("Unexpected key %s in with for step %d", k, i)
				} else if got != want {
					t.Errorf("Incorrect value for key %s in with for step %q in %q (got %q, want %q)", k, sid, id, got, want)
				}
			}

			for k, want := range want {
				if _, ok := got[k]; !ok {
					t.Errorf("Missing key %s in with for step %q in %q (want %q)", k, sid, id, want)
				}
			}
		}
	}
}
