// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"strings"
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestStep(t *testing.T) {
	type TestCase struct {
		yaml  string
		model Step
	}

	okCases := map[string]TestCase{
		"With a 'name:'": {
			yaml: `name: foobar`,
			model: Step{
				Name: "foobar",
			},
		},
		"With a 'uses:'": {
			yaml: `uses: foo@bar`,
			model: Step{
				Uses: Uses{
					Name: "foo",
					Ref:  "bar",
				},
			},
		},
		"With a 'run:'": {
			yaml: `run: echo 'foobar'`,
			model: Step{
				Run: "echo 'foobar'",
			},
		},
		"With a 'shell:'": {
			yaml: `shell: bash`,
			model: Step{
				Shell: "bash",
			},
		},
		"With a 'with:'": {
			yaml: `
with:
    foo: bar
`,
			model: Step{
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
			model: Step{
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
			model: Step{
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
			model: Step{
				Name: "foobar",
				Run:  "echo 'foobaz'",
			},
		},
		"With a 'run:' and 'shell:'": {
			yaml: `
shell: powershell
run: echo 'foobaz'
`,
			model: Step{
				Run:   "echo 'foobaz'",
				Shell: "powershell",
			},
		},
	}

	for name, tt := range okCases {
		t.Run("Marshal: "+name, func(t *testing.T) {
			got, err := yaml.Marshal(tt.model)
			if err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			if got, want := string(got), strings.TrimSpace(tt.yaml)+"\n"; got != want {
				t.Errorf("Unexpected result\n=== got ===\n%s\n=== want ===\n%s", got, want)
			}
		})

		t.Run("Unmarshal: "+name, func(t *testing.T) {
			var got Step
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkStep(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"yaml: invalid 'name' value": {
			yaml: `
name:
  foo: bar
`,
		},
		"yaml: invalid 'uses' value": {
			yaml: `
uses: ['foo', 'bar']
`,
		},
		"yaml: invalid 'shell' value": {
			yaml: `
shell:
  foo: bar
`,
		},
		"yaml: invalid 'run' value": {
			yaml: `
run: ['foo', 'bar']
`,
		},
		"yaml: invalid 'with' value": {
			yaml: `
with: not a map
`,
		},
		"yaml: invalid 'env' value": {
			yaml: `
env: not a map
`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			var err error
			if strings.HasPrefix(name, "model:") {
				_, err = yaml.Marshal(tt.model)
			} else if strings.HasPrefix(name, "yaml:") {
				err = yaml.Unmarshal([]byte(tt.yaml), &tt.model)
			} else {
				t.Fatalf("Incorrect test name %q", name)
			}

			if err == nil {
				t.Error("Want an error, got none")
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
		yaml  string
		model Uses
	}

	okCases := map[string]TestCase{
		"Versioned action, specific commit": {
			yaml: `actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3`,
			model: Uses{
				Name: "actions/checkout",
				Ref:  "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
			},
		},
		"Versioned action, specific commit, with annotation": {
			yaml: `actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3 # v4.2.0`,
			model: Uses{
				Name:       "actions/checkout",
				Ref:        "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
				Annotation: "v4.2.0",
			},
		},
		"Versioned action, major version": {
			yaml: `actions/checkout@v4`,
			model: Uses{
				Name: "actions/checkout",
				Ref:  "v4",
			},
		},
		"Versioned action, specific version": {
			yaml: `actions/checkout@v4.2.0`,
			model: Uses{
				Name: "actions/checkout",
				Ref:  "v4.2.0",
			},
		},
		"Versioned action, branch": {
			yaml: `actions/checkout@main`,
			model: Uses{
				Name: "actions/checkout",
				Ref:  "main",
			},
		},
		"Versioned action in a subdirectory": {
			yaml: `actions/aws/ec2@main`,
			model: Uses{
				Name: "actions/aws/ec2",
				Ref:  "main",
			},
		},
		"In the same repository as the workflow": {
			yaml: `./.github/actions/hello-world-action`,
			model: Uses{
				Name: "./.github/actions/hello-world-action",
			},
		},
		"Docker Hub action": {
			yaml: `docker://alpine:3.8`,
			model: Uses{
				Name: "docker://alpine:3.8",
			},
		},
		"GitHub Packages Container registry action": {
			yaml: `docker://ghcr.io/foo/bar`,
			model: Uses{
				Name: "docker://ghcr.io/foo/bar",
			},
		},
		"Docker public registry action": {
			yaml: `docker://gcr.io/cloud-builders/gradle`,
			model: Uses{
				Name: "docker://gcr.io/cloud-builders/gradle",
			},
		},
	}

	for name, tt := range okCases {
		t.Run("Marshal: "+name, func(t *testing.T) {
			got, err := yaml.Marshal(tt.model)
			if err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			if got, want := string(got), tt.yaml+"\n"; got != want {
				t.Errorf("Unexpected result (got %q, want %q)", got, want)
			}
		})

		t.Run("Unmarshal: "+name, func(t *testing.T) {
			var got Uses
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkUses(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"model: missing name": {
			model: Uses{
				Ref: "bar",
			},
		},
		"yaml: missing name": {
			yaml: `@bar`,
		},
		"yaml: empty ref": {
			yaml: `foo@`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			var err error
			if strings.HasPrefix(name, "model:") {
				_, err = yaml.Marshal(tt.model)
			} else if strings.HasPrefix(name, "yaml:") {
				err = yaml.Unmarshal([]byte(tt.yaml), &tt.model)
			} else {
				t.Fatalf("Incorrect test name %q", name)
			}

			if err == nil {
				t.Error("Want an error, got none")
			}
		})
	}

	roundtrip := func(u Uses) bool {
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

	if err := quick.Check(roundtrip, nil); err != nil {
		t.Error(err)
	}
}

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

func checkSteps(t *testing.T, got, want []Step) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of steps (got %d, want %d)", got, want)
		return
	}

	for i, got := range got {
		want := want[i]
		checkStep(t, &got, &want)
	}
}

func checkStep(t *testing.T, got, want *Step) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected name (got %q, want %q)", got, want)
	}

	if got, want := got.Run, want.Run; got != want {
		t.Errorf("Unexpected run (got %q, want %q)", got, want)
	}

	if got, want := got.Shell, want.Shell; got != want {
		t.Errorf("Unexpected shell (got %q, want %q)", got, want)
	}

	checkUses(t, &got.Uses, &want.Uses)
	checkMap(t, got.Env, want.Env)
	checkMap(t, got.With, want.With)
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
