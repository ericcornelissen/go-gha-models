// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"testing"
)

func TestWorkflow(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		type TestCase struct {
			yaml string
			want Workflow
		}

		testCases := map[string]TestCase{
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
									Uses: "actions/checkout@v3",
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
									Uses: "actions/checkout@v3",
								},
								{
									Name: "Echo value",
									Uses: "actions/github-script@v6",
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
									Uses: "actions/setup-node@v3",
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
									Uses:           "actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b",
									UsesAnnotation: "v4",
								},
							},
						},
					},
				},
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				workflow, err := ParseWorkflow([]byte(tt.yaml))
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}

				if got, want := len(workflow.Jobs), len(tt.want.Jobs); got != want {
					t.Fatalf("Unexpected number of jobs (got %d, want %d)", got, want)
				}

				for name, job := range workflow.Jobs {
					want := tt.want.Jobs[name]
					CheckJob(t, name, job, want)
				}
			})
		}
	})

	t.Run("Error", func(t *testing.T) {
		type TestCase struct {
			yaml string
		}

		testCases := map[string]TestCase{
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
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := ParseWorkflow([]byte(tt.yaml))
				if err == nil {
					t.Fatal("Expected an error, got none")
				}
			})
		}
	})
}
