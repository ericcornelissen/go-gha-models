// SPDX-License-Identifier: BSD-2-Clause-Patent

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

func checkManifest(t *testing.T, got, want *Manifest) {
	t.Helper()

	checkRuns(t, &got.Runs, &want.Runs)
}

func checkRuns(t *testing.T, got, want *ManifestRuns) {
	t.Helper()

	if got, want := len(got.Using), len(want.Using); got != want {
		t.Errorf("Unexpected using value (got %d, want %d)", got, want)
	}

	if got, want := len(got.Steps), len(want.Steps); got != want {
		t.Errorf("Unexpected number of steps (got %d, want %d)", got, want)
	}

	checkSteps(t, "runs", got.Steps, want.Steps)
}
