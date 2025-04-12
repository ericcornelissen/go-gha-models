// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"
	"strings"
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func ExampleParseWorkflow() {
	yaml := `
name: learn-github-actions
run-name: ${{ github.actor }} is learning GitHub Actions
on: [push]
jobs:
  check-bats-version:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: '20'
    - run: npm install -g bats
    - run: bats -v`

	workflow, _ := ParseWorkflow([]byte(yaml))
	fmt.Println(workflow.Name)
	// Output: learn-github-actions
}

func TestParseWorkflow(t *testing.T) {
	type TestCase struct {
		yaml  string
		model Workflow
	}

	okCases := map[string]TestCase{
		"Workflow with 'run:'": {
			yaml: `
name: Example workflow
jobs:
    example:
        name: Example job
        steps:
            - name: Checkout repository
              uses: actions/checkout@v3
              with:
                fetch-depth: "1"
            - name: Echo value
              run: echo '${{ inputs.value }}'
`,
			model: Workflow{
				Name: "Example workflow",
				Jobs: map[string]Job{
					"example": {
						Name: "Example job",
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
		"some permissions": {
			yaml: `
permissions:
    actions: read
    checks: write
    packages: write
    statuses: read
jobs:
    example:
        steps:
            - uses: actions/checkout@v4
`,
			model: Workflow{
				Permissions: Permissions{
					Actions:        "read",
					Attestations:   "none",
					Checks:         "write",
					Contents:       "none",
					Deployments:    "none",
					Discussions:    "none",
					IdToken:        "none",
					Issues:         "none",
					Models:         "none",
					Packages:       "write",
					Pages:          "none",
					PullRequests:   "none",
					SecurityEvents: "none",
					Statuses:       "read",
				},
				Jobs: map[string]Job{
					"example": {
						Steps: []Step{
							{
								Uses: Uses{
									Name: "actions/checkout",
									Ref:  "v4",
								},
							},
						},
					},
				},
			},
		},
		"permissions: read-all": {
			yaml: `
permissions: read-all
jobs:
    example:
        steps:
            - uses: actions/checkout@v4
`,
			model: Workflow{
				Permissions: Permissions{
					Actions:        "read",
					Attestations:   "read",
					Checks:         "read",
					Contents:       "read",
					Deployments:    "read",
					Discussions:    "read",
					IdToken:        "read",
					Issues:         "read",
					Models:         "read",
					Packages:       "read",
					Pages:          "read",
					PullRequests:   "read",
					SecurityEvents: "read",
					Statuses:       "read",
				},
				Jobs: map[string]Job{
					"example": {
						Steps: []Step{
							{
								Uses: Uses{
									Name: "actions/checkout",
									Ref:  "v4",
								},
							},
						},
					},
				},
			},
		},
		"permissions: write-all": {
			yaml: `
permissions: write-all
jobs:
    example:
        steps:
            - uses: actions/checkout@v4
`,
			model: Workflow{
				Permissions: Permissions{
					Actions:        "write",
					Attestations:   "write",
					Checks:         "write",
					Contents:       "write",
					Deployments:    "write",
					Discussions:    "write",
					IdToken:        "write",
					Issues:         "write",
					Models:         "write",
					Packages:       "write",
					Pages:          "write",
					PullRequests:   "write",
					SecurityEvents: "write",
					Statuses:       "write",
				},
				Jobs: map[string]Job{
					"example": {
						Steps: []Step{
							{
								Uses: Uses{
									Name: "actions/checkout",
									Ref:  "v4",
								},
							},
						},
					},
				},
			},
		},
		"permissions: {}": {
			yaml: `
permissions: {}
jobs:
    example:
        steps:
            - uses: actions/checkout@v4
`,
			model: Workflow{
				Permissions: Permissions{
					Actions:        "none",
					Attestations:   "none",
					Checks:         "none",
					Contents:       "none",
					Deployments:    "none",
					Discussions:    "none",
					IdToken:        "none",
					Issues:         "none",
					Models:         "none",
					Packages:       "none",
					Pages:          "none",
					PullRequests:   "none",
					SecurityEvents: "none",
					Statuses:       "none",
				},
				Jobs: map[string]Job{
					"example": {
						Steps: []Step{
							{
								Uses: Uses{
									Name: "actions/checkout",
									Ref:  "v4",
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
		"yaml: invalid 'permissions' value, scalar": {
			yaml: `
permissions: 3.14
`,
		},
		"yaml: invalid 'permissions' value, non-scalar": {
			yaml: `
permissions:
- foo
- bar
`,
		},
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

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected name (got %q, want %q)", got, want)
	}

	if got, want := got.RunName, want.RunName; got != want {
		t.Errorf("Unexpected run-name (got %q, want %q)", got, want)
	}

	if got, want := got.Concurrency.CancelInProgress, want.Concurrency.CancelInProgress; got != want {
		t.Errorf("Unexpected concurrency.cancel-in-progress (got %t, want %t)", got, want)
	}

	if got, want := got.Concurrency.Group, want.Concurrency.Group; got != want {
		t.Errorf("Unexpected concurrency.group (got %q, want %q)", got, want)
	}

	if got, want := got.Defaults.Run.Shell, want.Defaults.Run.Shell; got != want {
		t.Errorf("Unexpected defaults.run.shell (got %q, want %q)", got, want)
	}

	if got, want := got.Defaults.Run.WorkingDirectory, want.Defaults.Run.WorkingDirectory; got != want {
		t.Errorf("Unexpected defaults.run.working-directory (got %q, want %q)", got, want)
	}

	checkPermissions(t, got.Permissions, want.Permissions)
	checkMap(t, got.Env, want.Env)
	checkJobs(t, got.Jobs, want.Jobs)
}

func checkJobs(t *testing.T, got, want map[string]Job) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of jobs (got %d, want %d)", got, want)
		return
	}

	for name, got := range got {
		if want, ok := want[name]; !ok {
			t.Errorf("Got job named %q but it is not wanted", name)
		} else {
			checkJob(t, &got, &want)
		}
	}

	for name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("Want job named %q but it is not present", name)
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

func checkPermissions(t *testing.T, got, want Permissions) {
	if got, want := got.Actions, want.Actions; got != want {
		t.Errorf("Unexpected permission for 'actions' (got %q, want %q)", got, want)
	}

	if got, want := got.Attestations, want.Attestations; got != want {
		t.Errorf("Unexpected permission for 'attestations' (got %q, want %q)", got, want)
	}

	if got, want := got.Checks, want.Checks; got != want {
		t.Errorf("Unexpected permission for 'checks' (got %q, want %q)", got, want)
	}

	if got, want := got.Contents, want.Contents; got != want {
		t.Errorf("Unexpected permission for 'contents' (got %q, want %q)", got, want)
	}

	if got, want := got.Deployments, want.Deployments; got != want {
		t.Errorf("Unexpected permission for 'deployments' (got %q, want %q)", got, want)
	}

	if got, want := got.Discussions, want.Discussions; got != want {
		t.Errorf("Unexpected permission for 'discussions' (got %q, want %q)", got, want)
	}

	if got, want := got.IdToken, want.IdToken; got != want {
		t.Errorf("Unexpected permission for 'id-token' (got %q, want %q)", got, want)
	}

	if got, want := got.Issues, want.Issues; got != want {
		t.Errorf("Unexpected permission for 'issues' (got %q, want %q)", got, want)
	}

	if got, want := got.Models, want.Models; got != want {
		t.Errorf("Unexpected permission for 'models' (got %q, want %q)", got, want)
	}

	if got, want := got.Issues, want.Issues; got != want {
		t.Errorf("Unexpected permission for 'issues' (got %q, want %q)", got, want)
	}

	if got, want := got.Packages, want.Packages; got != want {
		t.Errorf("Unexpected permission for 'packages' (got %q, want %q)", got, want)
	}

	if got, want := got.Pages, want.Pages; got != want {
		t.Errorf("Unexpected permission for 'pages' (got %q, want %q)", got, want)
	}

	if got, want := got.PullRequests, want.PullRequests; got != want {
		t.Errorf("Unexpected permission for 'pull-requests' (got %q, want %q)", got, want)
	}

	if got, want := got.SecurityEvents, want.SecurityEvents; got != want {
		t.Errorf("Unexpected permission for 'security-events' (got %q, want %q)", got, want)
	}

	if got, want := got.Statuses, want.Statuses; got != want {
		t.Errorf("Unexpected permission for 'statuses' (got %q, want %q)", got, want)
	}
}
