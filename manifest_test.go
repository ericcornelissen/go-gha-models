// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestParseManifest(t *testing.T) {
	type TestCase struct {
		yaml string
		want Manifest
	}

	okCases := map[string]TestCase{
		"Metadata": {
			yaml: `
name: Action name
author: Action author
description: Action description
branding:
  color: black
  icon: coffee
`,
			want: Manifest{
				Name:        "Action name",
				Author:      "Action author",
				Description: "Action description",
				Branding: Branding{
					Color: "black",
					Icon:  "coffee",
				},
			},
		},
		"Composite manifest": {
			yaml: `
runs:
  using: composite
  steps:
  - name: Checkout repository
    uses: actions/checkout@v3
    with:
      fetch-depth: 1
  - name: Echo value (bash)
    shell: bash
    run: echo '${{ inputs.value }}'
  - name: Echo value (JavaScript)
    uses: actions/github-script@v6
    with:
      script: console.log('${{ inputs.value }}')
`,
			want: Manifest{
				Runs: Runs{
					Using: "composite",
					Steps: []Step{
						{
							Name: "Checkout repository",
							Uses: Uses{
								Name: "actions/checkout",
								Ref:  "v3",
							},
							With: map[string]string{
								"fetch-depth": "1",
							},
						},
						{
							Name:  "Echo value (bash)",
							Run:   "echo '${{ inputs.value }}'",
							Shell: "bash",
						},
						{
							Name: "Echo value (JavaScript)",
							Uses: Uses{
								Name: "actions/github-script",
								Ref:  "v6",
							},
							With: map[string]string{
								"script": "console.log('${{ inputs.value }}')",
							},
						},
					},
				},
			},
		},
		"Docker-based manifest": {
			yaml: `
runs:
  using: docker
  pre-entrypoint: pre.sh
  entrypoint: entry.sh
  post-entrypoint: post.sh
  args:
  - foo
  - bar
  env:
    foo: bar
`,
			want: Manifest{
				Runs: Runs{
					Using:          "docker",
					Args:           []string{"foo", "bar"},
					Env:            map[string]string{"foo": "bar"},
					PreEntrypoint:  "pre.sh",
					Entrypoint:     "entry.sh",
					PostEntrypoint: "post.sh",
				},
			},
		},
		"Node-based manifest": {
			yaml: `
runs:
  using: node22
  pre: pre.js
  pre-if: ${{ always() }}
  main: main.js
  post: post.js
  post-if: ${{ never() }}
`,
			want: Manifest{
				Runs: Runs{
					Using:  "node22",
					Pre:    "pre.js",
					PreIf:  "${{ always() }}",
					Main:   "main.js",
					Post:   "post.js",
					PostIf: "${{ never() }}",
				},
			},
		},
		"Inputs": {
			yaml: `
inputs:
  foo:
    default: bar
    description: Hello world!
    required: true
  deprecated:
    deprecationMessage: Hello world!
  optional:
    default: value
    required: false
`,
			want: Manifest{
				Inputs: map[string]Input{
					"foo": {
						Default:     "bar",
						Description: "Hello world!",
						Required:    true,
					},
					"deprecated": {
						DeprecationMessage: "Hello world!",
					},
					"optional": {
						Required: false,
						Default:  "value",
					},
				},
			},
		},
		"Outputs": {
			yaml: `
outputs:
  described:
    description: Hello world!
  valued:
    value: ${{ steps.random-number-generator.outputs.random-id }}
  described-value:
    description: A random number
    value: ${{ steps.random-number-generator.outputs.random-id }}
`,
			want: Manifest{
				Outputs: map[string]Output{
					"described": {
						Description: "Hello world!",
					},
					"valued": {
						Value: "${{ steps.random-number-generator.outputs.random-id }}",
					},
					"described-value": {
						Description: "A random number",
						Value:       "${{ steps.random-number-generator.outputs.random-id }}",
					},
				},
			},
		},
	}

	for name, tt := range okCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseManifest([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			checkManifest(t, &got, &tt.want)
		})
	}

	errCases := map[string]TestCase{
		"Invalid 'branding' value": {
			yaml: `branding: 3.14`,
		},
		"Invalid 'input' value": {
			yaml: `inputs: 3.14`,
		},
		"Invalid 'required' value for input": {
			yaml: `
inputs:
  foo:
    required: bar
`,
		},
		"Invalid 'output' value": {
			yaml: `outputs: 3.14`,
		},
		"Invalid 'runs' value": {
			yaml: `runs: 3.14`,
		},
		"Invalid 'args' value in runs": {
			yaml: `
runs:
  using: docker
  args: foobar
`,
		},
		"Invalid 'env' value in runs": {
			yaml: `
runs:
  using: docker
  env: foobar
`,
		},
		"Invalid 'steps' value in runs": {
			yaml: `
runs:
  using: composite
  steps: 3.14
`,
		},
		"Invalid 'env' value in step": {
			yaml: `
runs:
  using: composite
  steps:
  - env: 1.618
`,
		},
		"Invalid 'uses' value in step": {
			yaml: `
runs:
  using: composite
  steps:
  - uses: foobar
`,
		},
		"Invalid 'with' value in step": {
			yaml: `
runs:
  using: composite
  steps:
  - with: 1.618
`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseManifest([]byte(tt.yaml)); err == nil {
				t.Fatal("Want an error, got none")
			}
		})
	}

	roundtrip := func(m Manifest) bool {
		b, err := yaml.Marshal(m)
		if err != nil {
			return true
		}

		if err := yaml.Unmarshal(b, &m); err != nil {
			return false
		}

		return true
	}

	if err := quick.Check(roundtrip, nil); err != nil {
		t.Error(err)
	}
}

func FuzzParseManifest(f *testing.F) {
	seeds := []string{
		`
runs:
  using: node16
  main: index.js
`,
		`
runs:
  using: composite
  steps:
  - name: Checkout repository
    uses: actions/checkout@v3
    with:
      fetch-depth: 1
  - name: Echo value
    run: echo '${{ inputs.value }}'
`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ParseManifest(data)
	})
}

func checkManifest(t *testing.T, got, want *Manifest) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected manifest name (got %q, want %q)", got, want)
	}

	if got, want := got.Author, want.Author; got != want {
		t.Errorf("Unexpected manifest author (got %q, want %q)", got, want)
	}

	if got, want := got.Description, want.Description; got != want {
		t.Errorf("Unexpected manifest description (got %q, want %q)", got, want)
	}

	checkBranding(t, &got.Branding, &want.Branding)
	checkInputs(t, got.Inputs, want.Inputs)
	checkOutputs(t, got.Outputs, want.Outputs)
	checkRuns(t, &got.Runs, &want.Runs)
}

func checkBranding(t *testing.T, got, want *Branding) {
	t.Helper()

	if got, want := got.Color, want.Color; got != want {
		t.Errorf("Unexpected branding color (got %q, want %q)", got, want)
	}

	if got, want := got.Icon, want.Icon; got != want {
		t.Errorf("Unexpected branding icon (got %q, want %q)", got, want)
	}
}

func checkInputs(t *testing.T, got, want map[string]Input) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of inputs (got %d, want %d)", got, want)
		return
	}

	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("Unexpected input named %q", k)
		}
	}

	for k, want := range want {
		got, ok := got[k]
		if !ok {
			t.Errorf("Missing input named %q", k)
			continue
		}

		if got, want := got.Default, want.Default; got != want {
			t.Errorf("Unexpected input default (got %q, want %q)", got, want)
		}

		if got, want := got.DeprecationMessage, want.DeprecationMessage; got != want {
			t.Errorf("Unexpected input deprecation message (got %q, want %q)", got, want)
		}

		if got, want := got.Description, want.Description; got != want {
			t.Errorf("Unexpected input description (got %q, want %q)", got, want)
		}

		if got, want := got.Required, want.Required; got != want {
			t.Errorf("Unexpected input default (got %t, want %t)", got, want)
		}
	}
}

func checkOutputs(t *testing.T, got, want map[string]Output) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of output (got %d, want %d)", got, want)
		return
	}

	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("Unexpected output named %q", k)
		}
	}

	for k, want := range want {
		got, ok := got[k]
		if !ok {
			t.Errorf("Missing input named %q", k)
			continue
		}

		if got, want := got.Description, want.Description; got != want {
			t.Errorf("Unexpected output description (got %q, want %q)", got, want)
		}

		if got, want := got.Value, want.Value; got != want {
			t.Errorf("Unexpected output (got %q, want %q)", got, want)
		}
	}
}

func checkRuns(t *testing.T, got, want *Runs) {
	t.Helper()

	if got, want := got.Using, want.Using; got != want {
		t.Errorf("Unexpected runs using (got %q, want %q)", got, want)
	}

	/* using: composite */

	if got, want := got.Steps, want.Steps; len(got) != len(want) {
		t.Errorf("Unexpected number of runs steps (got %d, want %d)", len(got), len(want))
	} else {
		for i, got := range got {
			want := want[i]
			checkStep(t, &got, &want)
		}
	}

	/* using: docker */

	if got, want := got.Image, want.Image; got != want {
		t.Errorf("Unexpected runs image (got %q, want %q)", got, want)
	}

	if got, want := got.Args, want.Args; len(got) != len(want) {
		t.Errorf("Unexpected number of runs args (got %d, want %d)", len(got), len(want))
	} else {
		for i, got := range got {
			want := want[i]
			if got != want {
				t.Errorf("Unexpected %dth runs arg (got %q, want %q)", i, got, want)
			}
		}
	}

	if got, want := got.Env, want.Env; len(got) != len(want) {
		t.Errorf("Unexpected number of items in env (got %d, want %d)", len(got), len(want))
	} else {
		for k := range got {
			if _, ok := want[k]; !ok {
				t.Errorf("Unexpected key %q in env", k)
			}
		}

		for k, want := range want {
			got, ok := got[k]
			if !ok {
				t.Errorf("Missing key %q from env", k)
			}

			if got != want {
				t.Errorf("Incorrect value for key %q in env (got %q, want %q)", k, got, want)
			}
		}
	}

	if got, want := got.PreEntrypoint, want.PreEntrypoint; got != want {
		t.Errorf("Unexpected runs pre-entrypoint (got %q, want %q)", got, want)
	}

	if got, want := got.Entrypoint, want.Entrypoint; got != want {
		t.Errorf("Unexpected runs entrypoint (got %q, want %q)", got, want)
	}

	if got, want := got.PostEntrypoint, want.PostEntrypoint; got != want {
		t.Errorf("Unexpected runs post-entrypoint (got %q, want %q)", got, want)
	}

	/* using: docker */

	if got, want := got.Pre, want.Pre; got != want {
		t.Errorf("Unexpected runs pre (got %q, want %q)", got, want)
	}

	if got, want := got.PreIf, want.PreIf; got != want {
		t.Errorf("Unexpected runs pre-if (got %q, want %q)", got, want)
	}

	if got, want := got.Main, want.Main; got != want {
		t.Errorf("Unexpected runs main (got %q, want %q)", got, want)
	}

	if got, want := got.Post, want.Post; got != want {
		t.Errorf("Unexpected runs post (got %q, want %q)", got, want)
	}

	if got, want := got.PostIf, want.PostIf; got != want {
		t.Errorf("Unexpected runs post-if (got %q, want %q)", got, want)
	}
}
