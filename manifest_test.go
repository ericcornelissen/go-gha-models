// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"strings"
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestParseManifest(t *testing.T) {
	type TestCase struct {
		yaml  string
		model Manifest
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
runs:
    using: composite
`,
			model: Manifest{
				Name:        "Action name",
				Author:      "Action author",
				Description: "Action description",
				Branding: Branding{
					Color: "black",
					Icon:  "coffee",
				},
				Runs: Runs{
					Using: "composite",
				},
			},
		},
		"Composite manifest": {
			yaml: `
name: Composite action example
description: An example of a composite action
runs:
    using: composite
    steps:
        - name: Checkout repository
          uses: actions/checkout@v3
          with:
            fetch-depth: "1"
        - name: Echo value (bash)
          shell: bash
          run: echo '${{ inputs.value }}'
        - name: Echo value (JavaScript)
          uses: actions/github-script@v6
          with:
            script: console.log('${{ inputs.value }}')
`,
			model: Manifest{
				Name:        "Composite action example",
				Description: "An example of a composite action",
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
name: Docker action example
description: An example of a Docker action
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
			model: Manifest{
				Name:        "Docker action example",
				Description: "An example of a Docker action",
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
name: Node.js action example
description: An example of a Node.js action
runs:
    using: node22
    pre: pre.js
    pre-if: ${{ always() }}
    main: main.js
    post: post.js
    post-if: ${{ never() }}
`,
			model: Manifest{
				Name:        "Node.js action example",
				Description: "An example of a Node.js action",
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
name: Action input example
description: An example of Action inputs
inputs:
    deprecated:
        description: This input is deprecated
        deprecationMessage: Hello world!
    optional:
        description: This input is optional
        default: value
    required:
        description: This input is required
        required: true
runs:
    using: composite
`,
			model: Manifest{
				Name:        "Action input example",
				Description: "An example of Action inputs",
				Inputs: map[string]Input{
					"deprecated": {
						Description:        "This input is deprecated",
						DeprecationMessage: "Hello world!",
					},
					"optional": {
						Description: "This input is optional",
						Default:     "value",
						Required:    false,
					},
					"required": {
						Description: "This input is required",
						Required:    true,
					},
				},
				Runs: Runs{
					Using: "composite",
				},
			},
		},
		"Outputs": {
			yaml: `
name: Action output example
description: An example of Action outputs
outputs:
    described:
        description: Hello world!
        value: ""
    described-value:
        description: A random number
        value: ${{ steps.random-number-generator.outputs.random-id-1 }}
    valued:
        description: ""
        value: ${{ steps.random-number-generator.outputs.random-id-2 }}
runs:
    using: composite
`,
			model: Manifest{
				Name:        "Action output example",
				Description: "An example of Action outputs",
				Outputs: map[string]Output{
					"described": {
						Description: "Hello world!",
					},
					"described-value": {
						Description: "A random number",
						Value:       "${{ steps.random-number-generator.outputs.random-id-1 }}",
					},
					"valued": {
						Value: "${{ steps.random-number-generator.outputs.random-id-2 }}",
					},
				},
				Runs: Runs{
					Using: "composite",
				},
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
			got, err := ParseManifest([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkManifest(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"yaml: invalid 'name' value": {
			yaml: `
name: ['foo', 'bar']
`,
		},
		"yaml: invalid 'author' value": {
			yaml: `
author: ['foo', 'bar']
`,
		},
		"yaml: invalid 'description' value": {
			yaml: `
description: ['foo', 'bar']
`,
		},
		"yaml: invalid 'branding' value": {
			yaml: `
branding: 3.14
`,
		},
		"yaml: invalid 'color' value for branding": {
			yaml: `
branding:
  color: ['foo', 'bar']
`,
		},
		"yaml: invalid 'icon' value for branding": {
			yaml: `
branding:
  icon: ['foo', 'bar']
`,
		},
		"yaml: invalid 'inputs' value": {
			yaml: `
inputs: 3.14
`,
		},
		"yaml: invalid 'default' value for input": {
			yaml: `
inputs:
  foo:
    default: ['foo', 'bar']
`,
		},
		"yaml: invalid 'deprecationMessage' value for input": {
			yaml: `
inputs:
  foo:
    deprecationMessage: ['foo', 'bar']
`,
		},
		"yaml: invalid 'description' value for input": {
			yaml: `
inputs:
  foo:
    description: ['foo', 'bar']
`,
		},
		"yaml: invalid 'required' value for input": {
			yaml: `
inputs:
  foo:
    required: bar
`,
		},
		"yaml: invalid 'outputs' value": {
			yaml: `
outputs: 3.14
`,
		},
		"yaml: invalid 'description' value for output": {
			yaml: `
outputs:
  foo:
    description: ['foo', 'bar']
`,
		},
		"yaml: invalid 'value' value for output": {
			yaml: `
outputs:
  foo:
    value: ['foo', 'bar']
`,
		},
		"yaml: invalid 'runs' value": {
			yaml: `
runs: 3.14
`,
		},
		"yaml: invalid 'using' value in runs": {
			yaml: `
runs:
  using: ['foo', 'bar']
`,
		},
		"yaml: invalid 'steps' value in runs": {
			yaml: `
runs:
  using: composite
  steps: 3.14
`,
		},
		"yaml: invalid 'image' value in runs": {
			yaml: `
runs:
  using: docker
  image: ['foo', 'bar']
`,
		},
		"yaml: invalid 'args' value in runs": {
			yaml: `
runs:
  using: docker
  args: foobar
`,
		},
		"yaml: invalid 'env' value in runs": {
			yaml: `
runs:
  using: docker
  env: foobar
`,
		},
		"yaml: invalid 'pre-entrypoint' value in runs": {
			yaml: `
runs:
  using: docker
  pre-entrypoint: ['foo', 'bar']
`,
		},
		"yaml: invalid 'entrypoint' value in runs": {
			yaml: `
runs:
  using: docker
  entrypoint: ['foo', 'bar']
`,
		},
		"yaml: invalid 'post-entrypoint' value in runs": {
			yaml: `
runs:
  using: docker
  post-entrypoint: ['foo', 'bar']
`,
		},
		"yaml: invalid 'pre' value in runs": {
			yaml: `
runs:
  using: node
  pre: ['foo', 'bar']
`,
		},
		"yaml: invalid 'pre-if' value in runs": {
			yaml: `
runs:
  using: node
  pre-if: ['foo', 'bar']
`,
		},
		"yaml: invalid 'main' value in runs": {
			yaml: `
runs:
  using: node
  main: ['foo', 'bar']
`,
		},
		"yaml: invalid 'post' value in runs": {
			yaml: `
runs:
  using: node
  post: ['foo', 'bar']
`,
		},
		"yaml: invalid 'post-if' value in runs": {
			yaml: `
runs:
  using: node
  post-if: ['foo', 'bar']
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

	checkSteps(t, got.Steps, want.Steps)

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
