// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestStep(t *testing.T) {
	type TestCase struct {
		yaml string
		want Step
	}

	okCases := map[string]TestCase{
		"With a 'name:'": {
			yaml: `name: foobar`,
			want: Step{
				Name: "foobar",
			},
		},
		"With a 'uses:'": {
			yaml: `uses: foo@bar`,
			want: Step{
				Uses: Uses{
					Name: "foo",
					Ref:  "bar",
				},
			},
		},
		"With a 'run:'": {
			yaml: `run: echo 'foobar'`,
			want: Step{
				Run: "echo 'foobar'",
			},
		},
		"With a 'shell:'": {
			yaml: `shell: bash`,
			want: Step{
				Shell: "bash",
			},
		},
		"With a 'with:'": {
			yaml: `
with:
  foo: bar
`,
			want: Step{
				With: map[string]string{
					"foo": "bar",
				},
			},
		},
		"With an 'env:'": {
			yaml: `
env:
  foo: bar
`,
			want: Step{
				Env: map[string]string{
					"foo": "bar",
				},
			},
		},
		"With a 'name:' and 'uses:'": {
			yaml: `
name: foobar
uses: foo@bar
`,
			want: Step{
				Name: "foobar",
				Uses: Uses{
					Name: "foo",
					Ref:  "bar",
				},
			},
		},
		"With a 'name:' and 'run:'": {
			yaml: `
name: foobar
run: echo 'foobaz'
`,
			want: Step{
				Name: "foobar",
				Run:  "echo 'foobaz'",
			},
		},
		"With a 'run:' and 'shell:'": {
			yaml: `
shell: powershell
run: echo 'foobaz'
`,
			want: Step{
				Run:   "echo 'foobaz'",
				Shell: "powershell",
			},
		},
	}

	for name, tt := range okCases {
		t.Run(name, func(t *testing.T) {
			var got Step
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			checkStep(t, &got, &tt.want)
		})
	}

	errCases := map[string]TestCase{
		"Invalid 'env' value": {
			yaml: `
name: foobar
env: not a map
`,
		},
		"Invalid 'uses' value": {
			yaml: `
name: foo
uses: bar
`,
		},
		"Invalid 'with' value": {
			yaml: `
name: foobar
with: not a map
`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			var got Step
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err == nil {
				t.Fatal("Want an error, got none")
			}
		})
	}

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
	type TestCase struct {
		yaml string
		want Uses
	}

	okCases := map[string]TestCase{
		"Without annotation": {
			yaml: `foo@bar`,
			want: Uses{
				Name: "foo",
				Ref:  "bar",
			},
		},
		"With annotation": {
			yaml: `foo@bar # foobaz`,
			want: Uses{
				Name:       "foo",
				Ref:        "bar",
				Annotation: "foobaz",
			},
		},
	}

	for name, tt := range okCases {
		t.Run(name, func(t *testing.T) {
			var got Uses
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			checkUses(t, &got, &tt.want)
		})
	}

	errCases := map[string]TestCase{
		"Missing ref": {
			yaml: `foobar`,
		},
		"Empty ref": {
			yaml: `foo@`,
		},
		"Empty name": {
			yaml: `@bar`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			var got Uses
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err == nil {
				t.Fatal("Want an error, got none")
			}
		})
	}

	roundtrip := func(u Uses) bool {
		if u.Name == "" || u.Ref == "" {
			return true
		}

		b, err := yaml.Marshal(u)
		if err != nil {
			return false
		}

		if err = yaml.Unmarshal(b, &u); err != nil {
			return false
		}

		return true
	}

	if err := quick.Check(roundtrip, nil); err != nil {
		t.Error(err)
	}
}

func checkStep(t *testing.T, got, want *Step) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected name (got %q, want %q)", got, want)
	}

	checkUses(t, &got.Uses, &want.Uses)

	if got, want := got.Run, want.Run; got != want {
		t.Errorf("Unexpected run (got %q, want %q)", got, want)
	}

	if got, want := got.Shell, want.Shell; got != want {
		t.Errorf("Unexpected shell (got %q, want %q)", got, want)
	}

	if got, want := got.Env, want.Env; len(got) != len(want) {
		t.Errorf("Unexpected number of items in env (got %d, want %d)", len(got), len(want))
	} else {
		for k, got := range got {
			if want, ok := want[k]; !ok {
				t.Errorf("Unexpected key %q in env", k)
			} else if got != want {
				t.Errorf("Incorrect value for key %q in env (got %q, want %q)", k, got, want)
			}
		}

		for k, want := range want {
			if _, ok := got[k]; !ok {
				t.Errorf("Missing key %q in env (want %q)", k, want)
			}
		}
	}

	if got, want := got.With, want.With; len(got) != len(want) {
		t.Errorf("Unexpected number of items in with (got %d, want %d)", len(got), len(want))
	} else {
		for k, got := range got {
			if want, ok := want[k]; !ok {
				t.Errorf("Unexpected key %q in with", k)
			} else if got != want {
				t.Errorf("Incorrect value for key %q in with (got %q, want %q)", k, got, want)
			}
		}

		for k, want := range want {
			if _, ok := got[k]; !ok {
				t.Errorf("Missing key %q in with (want %q)", k, want)
			}
		}
	}
}

func checkUses(t *testing.T, got, want *Uses) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected uses name (got %q, want %q)", got, want)
	}

	if got, want := got.Ref, want.Ref; got != want {
		t.Errorf("Unexpected uses ref (got %q, want %q)", got, want)
	}

	if got, want := got.Annotation, want.Annotation; got != want {
		t.Errorf("Unexpected uses annotation (got %q, want %q)", got, want)
	}
}
