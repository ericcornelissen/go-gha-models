// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"strings"
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func TestParseWorkflow(t *testing.T) {
	type TestCase struct {
		yaml  string
		model Workflow
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
                fetch-depth: "1"
            - name: Echo value
              run: echo '${{ inputs.value }}'
`,
			model: Workflow{
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
                fetch-depth: "1"
            - name: Echo value
              uses: actions/github-script@v6
              with:
                script: console.log('${{ inputs.value }}')
`,
			model: Workflow{
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
                node-version: "20"
            - run: echo ${{ inputs.value }}
`,
			model: Workflow{
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
			model: Workflow{
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
			got, err := ParseWorkflow([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkWorkflow(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"yaml: invalid 'jobs' value": {
			yaml: `
jobs: 3.14
`,
		},
		"yaml: invalid 'name' value": {
			yaml: `
jobs:
  example:
    name:
    - uses: actions/checkout@v4
`,
		},
		"yaml: invalid 'steps' value": {
			yaml: `
jobs:
  example:
    steps: 42
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

func FuzzParseWorkflow(f *testing.F) {
	seeds := []string{
		`
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
		`
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
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ParseManifest(data)
	})
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
				checkJob(t, &got, &want)
			}
		}
	}
}

func checkJob(t *testing.T, got, want *Job) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected name (got %q, want %q)", got, want)
	}

	checkSteps(t, got.Steps, want.Steps)
}
