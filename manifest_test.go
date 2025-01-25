// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"testing"
)

func TestManifest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		type TestCase struct {
			yaml string
			want Manifest
		}

		testCases := map[string]TestCase{
			"Non-composite manifest": {
				yaml: `
runs:
  using: node16
  main: index.js
`,
				want: Manifest{
					Runs: ManifestRuns{
						Using: "node16",
					},
				},
			},
			"Manifest with 'run:'": {
				yaml: `
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
				want: Manifest{
					Runs: ManifestRuns{
						Using: "composite",
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
			"Manifest with 'actions/github-script'": {
				yaml: `
runs:
  using: composite
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
				want: Manifest{
					Runs: ManifestRuns{
						Using: "composite",
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
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				manifest, err := ParseManifest([]byte(tt.yaml))
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}

				if got, want := len(manifest.Runs.Using), len(tt.want.Runs.Using); got != want {
					t.Fatalf("Unexpected using value (got %d, want %d)", got, want)
				}

				if got, want := len(manifest.Runs.Steps), len(tt.want.Runs.Steps); got != want {
					t.Fatalf("Unexpected number of steps (got %d, want %d)", got, want)
				}

				for i, step := range manifest.Runs.Steps {
					want := tt.want.Runs.Steps[i]

					if got, want := step.Name, want.Name; got != want {
						t.Errorf("Unexpected name for step %d (got %q, want %q)", i, got, want)
					}

					if got, want := step.Run, want.Run; got != want {
						t.Errorf("Unexpected run for step %d (got %q, want %q)", i, got, want)
					}

					if got, want := step.Shell, want.Shell; got != want {
						t.Errorf("Unexpected shell for step %d (got %q, want %q)", i, got, want)
					}

					if got, want := step.Uses, want.Uses; got != want {
						t.Errorf("Unexpected uses for step %d (got %q, want %q)", i, got, want)
					}

					if got, want := step.With["script"], want.With["script"]; got != want {
						t.Errorf("Unexpected with for step %d (got %q, want %q)", i, got, want)
					}
				}
			})
		}
	})

	t.Run("Error", func(t *testing.T) {
		type TestCase struct {
			yaml string
		}

		testCases := map[string]TestCase{
			"Invalid 'runs' value": {
				yaml: `runs: 3.14`,
			},
			"Invalid 'steps' value": {
				yaml: `
runs:
  steps: 3.14
`,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := ParseManifest([]byte(tt.yaml))
				if err == nil {
					t.Fatal("Expected an error, got none")
				}
			})
		}
	})
}
