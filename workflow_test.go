// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"
	"slices"
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
		"Workflow metadata": {
			yaml: `
name: Example workflow
run-name: Example run by @${{ github.actor }}
concurrency:
    cancel-in-progress: true
    group: group A
defaults:
    run:
        shell: bash
        working-directory: ./scripts
env:
    FOO: bar
jobs: {}
`,
			model: Workflow{
				Name:    "Example workflow",
				RunName: "Example run by @${{ github.actor }}",
				Concurrency: Concurrency{
					CancelInProgress: true,
					Group:            "group A",
				},
				Env: map[string]string{"FOO": "bar"},
				Defaults: Defaults{
					Run: DefaultsRun{
						Shell:            "bash",
						WorkingDirectory: "./scripts",
					},
				},
			},
		},
		"Job metadata": {
			yaml: `
jobs:
    job1:
        name: Example
        runs-on: ubuntu-latest
        environment: foo-env
        continue-on-error: true
        timeout-minutes: 60
        if: foo == 'bar'
        needs:
            - job2
        defaults:
            run:
                shell: bash
                working-directory: ./scripts
        outputs:
            output1: ${{ steps.step1.outputs.test }}
            output2: ${{ steps.step2.outputs.test }}
        concurrency:
            cancel-in-progress: true
            group: group B
        permissions:
            packages: write
            statuses: read
        env:
            FOO: baz
    job2:
        environment:
            name: bar-env
            url: https://example.com
`,
			model: Workflow{
				Jobs: map[string]Job{
					"job1": {
						Name:   "Example",
						RunsOn: "ubuntu-latest",
						Environment: Environment{
							Name: "foo-env",
						},
						ContinueOnError: true,
						TimeoutMinutes:  60,
						If:              "foo == 'bar'",
						Needs:           []string{"job2"},
						Defaults: Defaults{
							Run: DefaultsRun{
								Shell:            "bash",
								WorkingDirectory: "./scripts",
							},
						},
						Outputs: map[string]string{
							"output1": "${{ steps.step1.outputs.test }}",
							"output2": "${{ steps.step2.outputs.test }}",
						},
						Concurrency: Concurrency{
							CancelInProgress: true,
							Group:            "group B",
						},
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
							Packages:       "write",
							Pages:          "none",
							PullRequests:   "none",
							SecurityEvents: "none",
							Statuses:       "read",
						},
						Env: map[string]string{"FOO": "baz"},
					},
					"job2": {
						Environment: Environment{
							Name: "bar-env",
							Url:  "https://example.com",
						},
					},
				},
			},
		},
		"Job with 'uses:'": {
			yaml: `
jobs:
    example:
        uses: octo-org/example-repo/.github/workflows/called-workflow.yml@main
        with:
            foo: bar
        secrets:
            foo: baz
`,
			model: Workflow{
				Jobs: map[string]Job{
					"example": {
						Uses:    "octo-org/example-repo/.github/workflows/called-workflow.yml@main",
						With:    map[string]string{"foo": "bar"},
						Secrets: map[string]string{"foo": "baz"},
					},
				},
			},
		},
		"Workflow with only some permissions": {
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
		"Workflow with `permissions: read-all`": {
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
		"Workflow with `permissions: write-all`": {
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
		"Workflow with `permissions: {}`": {
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
		"yaml: invalid job 'name' value": {
			yaml: `
jobs:
  example:
    name:
    - uses: actions/checkout@v4
`,
		},
		"yaml: invalid job 'environment' value": {
			yaml: `
jobs:
  example:
    environment:
    - foo
    - bar
`,
		},
		"yaml: invalid job 'steps' value": {
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
			switch {
			case strings.HasPrefix(name, "model:"):
				_, err = yaml.Marshal(tt.model)
			case strings.HasPrefix(name, "yaml:"):
				err = yaml.Unmarshal([]byte(tt.yaml), &tt.model)
			default:
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

	checkConcurrency(t, got.Concurrency, want.Concurrency)
	checkDefaults(t, got.Defaults, want.Defaults)
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

	if got, want := got.RunsOn, want.RunsOn; got != want {
		t.Errorf("Unexpected runs-on (got %q, want %q)", got, want)
	}

	if got, want := got.TimeoutMinutes, want.TimeoutMinutes; got != want {
		t.Errorf("Unexpected timeout-minutes (got %d, want %d)", got, want)
	}

	if got, want := got.If, want.If; got != want {
		t.Errorf("Unexpected if (got %q, want %q)", got, want)
	}

	if got, want := got.Needs, want.Needs; !slices.Equal(got, want) {
		t.Errorf("Unexpected needs (got %q, want %q)", got, want)
	}

	if got, want := got.Uses, want.Uses; got != want {
		t.Errorf("Unexpected uses (got %q, want %q)", got, want)
	}

	checkConcurrency(t, got.Concurrency, want.Concurrency)
	checkDefaults(t, got.Defaults, want.Defaults)
	checkMap(t, got.Env, want.Env)
	checkEnvironment(t, got.Environment, want.Environment)
	checkMap(t, got.Outputs, want.Outputs)
	checkPermissions(t, got.Permissions, want.Permissions)
	checkSteps(t, got.Steps, want.Steps)
	checkMap(t, got.With, want.With)
	checkMap(t, got.Secrets, want.Secrets)
}

func checkConcurrency(t *testing.T, got, want Concurrency) {
	if got, want := got.CancelInProgress, want.CancelInProgress; got != want {
		t.Errorf("Unexpected concurrency.cancel-in-progress (got %t, want %t)", got, want)
	}

	if got, want := got.Group, want.Group; got != want {
		t.Errorf("Unexpected concurrency.group (got %q, want %q)", got, want)
	}
}

func checkDefaults(t *testing.T, got, want Defaults) {
	if got, want := got.Run.Shell, want.Run.Shell; got != want {
		t.Errorf("Unexpected defaults.run.shell (got %q, want %q)", got, want)
	}

	if got, want := got.Run.WorkingDirectory, want.Run.WorkingDirectory; got != want {
		t.Errorf("Unexpected defaults.run.working-directory (got %q, want %q)", got, want)
	}
}

func checkEnvironment(t *testing.T, got, want Environment) {
	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected environment.name (got %q, want %q)", got, want)
	}

	if got, want := got.Url, want.Url; got != want {
		t.Errorf("Unexpected environment.url (got %q, want %q)", got, want)
	}
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
