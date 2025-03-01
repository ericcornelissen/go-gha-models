// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestStep(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
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
uses: ['foo', 'bar']
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
	})

	t.Run("Roundtrip", func(t *testing.T) {
		f := func(s Step) bool {
			b, err := yaml.Marshal(s)
			if err != nil {
				return true
			}

			if err = yaml.Unmarshal(b, &s); err != nil {
				return false
			}

			return true
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestUses(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		type TestCase struct {
			uses Uses
			want string
		}

		okCases := map[string]TestCase{
			"Versioned action, specific commit": {
				uses: Uses{
					Name: "actions/checkout",
					Ref:  "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
				},
				want: `actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3`,
			},
			"Versioned action, specific commit, with annotation": {
				uses: Uses{
					Name:       "actions/checkout",
					Ref:        "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
					Annotation: "v4.2.0",
				},
				want: `actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3 # v4.2.0`,
			},
			"Versioned action, major version": {
				uses: Uses{
					Name: "actions/checkout",
					Ref:  "v4",
				},
				want: `actions/checkout@v4`,
			},
			"Versioned action, specific version": {
				uses: Uses{
					Name: "actions/checkout",
					Ref:  "v4.2.0",
				},
				want: `actions/checkout@v4.2.0`,
			},
			"Versioned action, branch": {
				uses: Uses{
					Name: "actions/checkout",
					Ref:  "main",
				},
				want: `actions/checkout@main`,
			},
			"Versioned action in a subdirectory": {
				uses: Uses{
					Name: "actions/aws/ec2",
					Ref:  "main",
				},
				want: `actions/aws/ec2@main`,
			},
			"In the same repository as the workflow": {
				uses: Uses{
					Name: "./.github/actions/hello-world-action",
				},
				want: `./.github/actions/hello-world-action`,
			},
			"Docker Hub action": {
				uses: Uses{
					Name: "docker://alpine:3.8",
				},
				want: `docker://alpine:3.8`,
			},
			"GitHub Packages Container registry action": {
				uses: Uses{
					Name: "docker://ghcr.io/foo/bar",
				},
				want: `docker://ghcr.io/foo/bar`,
			},
			"Docker public registry action": {
				uses: Uses{
					Name: "docker://gcr.io/cloud-builders/gradle",
				},
				want: `docker://gcr.io/cloud-builders/gradle`,
			},
		}

		for name, tt := range okCases {
			t.Run(name, func(t *testing.T) {
				got, err := yaml.Marshal(tt.uses)
				if err != nil {
					t.Fatalf("Want no error, got %#v", err)
				}

				if got, want := string(got), tt.want+"\n"; got != want {
					t.Errorf("Unexpected result (got %q, want %q)", got, want)
				}
			})
		}

		errCases := map[string]TestCase{
			"Missing name": {
				uses: Uses{
					Name: "",
					Ref:  "bar",
				},
			},
		}

		for name, tt := range errCases {
			t.Run(name, func(t *testing.T) {
				if _, err := yaml.Marshal(tt.uses); err == nil {
					t.Fatal("Want an error, got none")
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		type TestCase struct {
			yaml string
			want Uses
		}

		okCases := map[string]TestCase{
			"Versioned action, specific commit": {
				yaml: `actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3`,
				want: Uses{
					Name: "actions/checkout",
					Ref:  "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
				},
			},
			"Versioned action, specific commit, with annotation": {
				yaml: `actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3 # v4.2.0`,
				want: Uses{
					Name:       "actions/checkout",
					Ref:        "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
					Annotation: "v4.2.0",
				},
			},
			"Versioned action, major version": {
				yaml: `actions/checkout@v4`,
				want: Uses{
					Name: "actions/checkout",
					Ref:  "v4",
				},
			},
			"Versioned action, specific version": {
				yaml: `actions/checkout@v4.2.0`,
				want: Uses{
					Name: "actions/checkout",
					Ref:  "v4.2.0",
				},
			},
			"Versioned action, branch": {
				yaml: `actions/checkout@main`,
				want: Uses{
					Name: "actions/checkout",
					Ref:  "main",
				},
			},
			"Versioned action in a subdirectory": {
				yaml: `actions/aws/ec2@main`,
				want: Uses{
					Name: "actions/aws/ec2",
					Ref:  "main",
				},
			},
			"In the same repository as the workflow": {
				yaml: `./.github/actions/hello-world-action`,
				want: Uses{
					Name: "./.github/actions/hello-world-action",
				},
			},
			"Docker Hub action": {
				yaml: `docker://alpine:3.8`,
				want: Uses{
					Name: "docker://alpine:3.8",
				},
			},
			"GitHub Packages Container registry action": {
				yaml: `docker://ghcr.io/foo/bar`,
				want: Uses{
					Name: "docker://ghcr.io/foo/bar",
				},
			},
			"Docker public registry action": {
				yaml: `docker://gcr.io/cloud-builders/gradle`,
				want: Uses{
					Name: "docker://gcr.io/cloud-builders/gradle",
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
	})

	t.Run("Roundtrip", func(t *testing.T) {
		f := func(u Uses) bool {
			b, err := yaml.Marshal(u)
			if err != nil {
				return true
			}

			var got Uses
			if err = yaml.Unmarshal(b, &got); err != nil {
				return false
			}

			return true
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
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
