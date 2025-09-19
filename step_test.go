// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"testing"

	"go.yaml.in/yaml/v3"
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
		"With an 'id:'": {
			yaml: `id: foobar`,
			model: Step{
				Id: "foobar",
			},
		},
		"With an 'if:'": {
			yaml: `if: ${{ 'foo' == 'bar' }}`,
			model: Step{
				If: "${{ 'foo' == 'bar' }}",
			},
		},
		"With 'continue-on-error:'": {
			yaml: `continue-on-error: true`,
			model: Step{
				ContinueOnError: true,
			},
		},
		"With 'timeout-minutes:'": {
			yaml: `timeout-minutes: 10`,
			model: Step{
				TimeoutMinutes: 10,
			},
		},
		"With a 'working-directory:'": {
			yaml: `working-directory: /foo/bar`,
			model: Step{
				WorkingDirectory: "/foo/bar",
			},
		},
		"With a 'shell:'": {
			yaml: `shell: bash`,
			model: Step{
				Shell: "bash",
			},
		},
		"With a 'run:'": {
			yaml: `run: echo 'foobar'`,
			model: Step{
				Run: "echo 'foobar'",
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
		t.Run(name, func(t *testing.T) {
			var got Step
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkStep(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"invalid 'name' value": {
			yaml: `
name:
  foo: bar
`,
		},
		"invalid 'uses' value": {
			yaml: `
uses: ['foo', 'bar']
`,
		},
		"invalid 'id' value": {
			yaml: `
id:
  foo: bar
`,
		},
		"invalid 'if' value": {
			yaml: `
if:
  foo: bar
`,
		},
		"invalid 'continue-on-error' value": {
			yaml: `continue-on-error: foobar`,
		},
		"invalid 'timeout-minutes' value": {
			yaml: `timeout-minutes: foobar`,
		},
		"invalid 'working-directory' value": {
			yaml: `
working-directory:
  foo: bar
`,
		},
		"invalid 'shell' value": {
			yaml: `
shell:
  foo: bar
`,
		},
		"invalid 'run' value": {
			yaml: `
run: ['foo', 'bar']
`,
		},
		"invalid 'with' value": {
			yaml: `
with: not a map
`,
		},
		"invalid 'env' value": {
			yaml: `
env: not a map
`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			if err := yaml.Unmarshal([]byte(tt.yaml), &tt.model); err == nil {
				t.Error("Want an error, got none")
			}
		})
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
		t.Run(name, func(t *testing.T) {
			var got Uses
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkUses(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"missing name": {
			yaml: `@bar`,
		},
		"empty ref": {
			yaml: `foo@`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			if err := yaml.Unmarshal([]byte(tt.yaml), &tt.model); err == nil {
				t.Error("Want an error, got none")
			}
		})
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

	if got, want := got.Id, want.Id; got != want {
		t.Errorf("Unexpected id (got %q, want %q)", got, want)
	}

	if got, want := got.If, want.If; got != want {
		t.Errorf("Unexpected if (got %q, want %q)", got, want)
	}

	if got, want := got.ContinueOnError, want.ContinueOnError; got != want {
		t.Errorf("Unexpected continue-on-error (got %t, want %t)", got, want)
	}

	if got, want := got.TimeoutMinutes, want.TimeoutMinutes; got != want {
		t.Errorf("Unexpected timeout-minutes (got %d, want %d)", got, want)
	}

	if got, want := got.WorkingDirectory, want.WorkingDirectory; got != want {
		t.Errorf("Unexpected working-directory (got %q, want %q)", got, want)
	}

	if got, want := got.Shell, want.Shell; got != want {
		t.Errorf("Unexpected shell (got %q, want %q)", got, want)
	}

	if got, want := got.Run, want.Run; got != want {
		t.Errorf("Unexpected run (got %q, want %q)", got, want)
	}

	checkUses(t, &got.Uses, &want.Uses)
	checkMap(t, got.With, want.With)
	checkMap(t, got.Env, want.Env)
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
