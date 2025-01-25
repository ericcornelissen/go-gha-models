// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestWorkflow(t *testing.T) {
	type TestCase struct {
		yaml string
		want Workflow
	}

	okCases := map[string]TestCase{
		"Workflow with 'run:'": {
			yaml: `
jobs:
  example:
    name: Example
    steps:
    - name: Checkout repository
      uses: actions/checkout@v3
      with:
        fetch-depth: 1
    - name: Echo value
      run: echo '${{ inputs.value }}'
`,
			want: Workflow{
				Jobs: map[string]Job{
					"example": {
						Name: "Example",
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
								Name: "Echo value",
								Run:  "echo '${{ inputs.value }}'",
							},
						},
					},
				},
			},
		},
		"Workflow with 'actions/github-script'": {
			yaml: `
jobs:
  example:
    name: Example
    steps:
    - name: Checkout repository
      uses: actions/checkout@v3
      with:
        fetch-depth: 1
    - name: Echo value
      uses: actions/github-script@v6
      with:
        script: console.log('${{ inputs.value }}')
`,
			want: Workflow{
				Jobs: map[string]Job{
					"example": {
						Name: "Example",
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
								Name: "Echo value",
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
		},
		"No names": {
			yaml: `
jobs:
  example:
    steps:
    - uses: actions/setup-node@v3
      with:
        node-version: 20
    - run: echo ${{ inputs.value }}
`,
			want: Workflow{
				Jobs: map[string]Job{
					"example": {
						Name: "",
						Steps: []Step{
							{
								Uses: Uses{
									Name: "actions/setup-node",
									Ref:  "v3",
								},
								With: map[string]string{
									"node-version": "20",
								},
							},
							{
								Run: "echo ${{ inputs.value }}",
							},
						},
					},
				},
			},
		},
		"Version annotation": {
			yaml: `
jobs:
  example:
    steps:
    - uses: actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b # v4
`,
			want: Workflow{
				Jobs: map[string]Job{
					"example": {
						Name: "",
						Steps: []Step{
							{
								Uses: Uses{
									Name:       "actions/checkout",
									Ref:        "0ad4b8fadaa221de15dcec353f45205ec38ea70b",
									Annotation: "v4",
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tt := range okCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseWorkflow([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			checkWorkflow(t, &got, &tt.want)
		})
	}

	errCases := map[string]TestCase{
		"Invalid 'jobs' value": {
			yaml: `
jobs: 3.14
`,
		},
		"Invalid 'steps' value": {
			yaml: `
jobs:
  example:
    steps: 42
`,
		},
		"Invalid 'env' value": {
			yaml: `
jobs:
  example:
    steps:
    - env: 1.618
`,
		},
		"Invalid 'with' value": {
			yaml: `
jobs:
  example:
    steps:
    - with: 1.618
`,
		},
		"Invalid 'uses' value, no ref": {
			yaml: `
jobs:
  example:
    steps:
    - uses: foobar
`,
		},
		"Invalid 'uses' value, empty ref": {
			yaml: `
jobs:
  example:
    steps:
    - uses: foobar@
`,
		},
		"Invalid 'uses' value, empty name": {
			yaml: `
jobs:
  example:
    steps:
    - uses: @v1.2.3
`,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseWorkflow([]byte(tt.yaml)); err == nil {
				t.Fatal("Want an error, got none")
			}
		})
	}

	roundtrip := func(w Workflow) bool {
		b, err := yaml.Marshal(w)
		if err != nil {
			return true
		}

		if err := yaml.Unmarshal(b, &w); err != nil {
			return false
		}

		return true
	}

	if err := quick.Check(roundtrip, nil); err != nil {
		t.Error(err)
	}
}

func checkWorkflow(t *testing.T, got, want *Workflow) {
	t.Helper()

	if got, want := got.Jobs, want.Jobs; len(got) != len(want) {
		t.Errorf("Unexpected number of jobs (got %d, want %d)", len(got), len(want))
	} else {
		for name := range want {
			if _, ok := got[name]; !ok {
				t.Errorf("Want job named %q but it is not present", name)
			}
		}

		for name, got := range got {
			if want, ok := want[name]; !ok {
				t.Errorf("Got job named %q but it is not wanted", name)
			} else {
				checkJob(t, name, &got, &want)
			}
		}
	}
}

func checkJob(t *testing.T, id string, got, want *Job) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected name for job %q (got %q, want %q)", id, got, want)
	} else if got != "" {
		id = got
	}

	checkSteps(t, id, got.Steps, want.Steps)
}
