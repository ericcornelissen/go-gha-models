// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"
	"reflect"
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
        environment: foo-env
        continue-on-error: true
        timeout-minutes: 60
        if: foo == 'bar'
        needs:
            - job2
        concurrency:
            cancel-in-progress: true
            group: group B
        defaults:
            run:
                shell: bash
                working-directory: ./scripts
        strategy:
            fail-fast: true
            max-parallel: 10
        services:
            nginx:
                image: nginx
                credentials:
                    username: foo
                    password: bar
                env:
                    FOO: bar
                ports:
                    - 80
                volumes:
                    - my_docker_volume:/volume_mount
                options: --cpus 1
        outputs:
            output1: ${{ steps.step1.outputs.test }}
            output2: ${{ steps.step2.outputs.test }}
        permissions:
            packages: write
            statuses: read
        env:
            FOO: baz
    job2:
        environment:
            name: bar-env
            url: https://example.com
        defaults:
            run:
                shell: bash
        permissions:
            attestations: write
            models: read
`,
			model: Workflow{
				Jobs: map[string]Job{
					"job1": {
						Name: "Example",
						Environment: Environment{
							Name: "foo-env",
						},
						ContinueOnError: true,
						TimeoutMinutes:  60,
						If:              "foo == 'bar'",
						Needs: []string{
							"job2",
						},
						Concurrency: Concurrency{
							CancelInProgress: true,
							Group:            "group B",
						},
						Defaults: Defaults{
							Run: DefaultsRun{
								Shell:            "bash",
								WorkingDirectory: "./scripts",
							},
						},
						Strategy: Strategy{
							FailFast:    true,
							MaxParallel: 10,
						},
						Services: map[string]Service{
							"nginx": {
								Image: "nginx",
								Credentials: ServiceCredentials{
									Username: "foo",
									Password: "bar",
								},
								Env: map[string]string{
									"FOO": "bar",
								},
								Ports: []int{
									80,
								},
								Volumes: []string{
									"my_docker_volume:/volume_mount",
								},
								Options: "--cpus 1",
							},
						},
						Outputs: map[string]string{
							"output1": "${{ steps.step1.outputs.test }}",
							"output2": "${{ steps.step2.outputs.test }}",
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
						Defaults: Defaults{
							Run: DefaultsRun{
								Shell: "bash",
							},
						},
						Permissions: Permissions{
							Actions:        "none",
							Attestations:   "write",
							Checks:         "none",
							Contents:       "none",
							Deployments:    "none",
							Discussions:    "none",
							IdToken:        "none",
							Issues:         "none",
							Models:         "read",
							Packages:       "none",
							Pages:          "none",
							PullRequests:   "none",
							SecurityEvents: "none",
							Statuses:       "none",
						},
					},
				},
			},
		},
		"Job with steps": {
			yaml: `
jobs:
    example:
        name: Example job
        steps:
            - name: Checkout repository
              uses: actions/checkout@v3
              with:
                persist-credentials: "false"
            - uses: actions/setup-node@v3
              with:
                node-version: "20"
            - name: Echo value (run)
              run: echo '${{ inputs.value }}'
            - name: Echo value (uses)
              uses: actions/github-script@v6
              with:
                script: console.log('${{ inputs.value }}')

`,
			model: Workflow{
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
									"persist-credentials": "false",
								},
							},
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
								Name: "Echo value (run)",
								Run:  "echo '${{ inputs.value }}'",
							},
							{
								Name: "Echo value (uses)",
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
		"Job with 'uses:'": {
			yaml: `
jobs:
    example:
        uses: octo-org/example-repo/.github/workflows/called-workflow.yml@main
        with:
            foo: bar
`,
			model: Workflow{
				Jobs: map[string]Job{
					"example": {
						Uses: "octo-org/example-repo/.github/workflows/called-workflow.yml@main",
						With: map[string]string{"foo": "bar"},
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
		"Job matrix": {
			yaml: `
jobs:
    0-one-dimensional-matrix:
        strategy:
            matrix:
                version:
                    - 10
                    - 12
                    - 14
    1-two-dimensional-matrix:
        strategy:
            matrix:
                os:
                    - ubuntu-22.04
                    - ubuntu 24.04
                version:
                    - 10
                    - 12
                    - 14
    2-nested-values-matrix:
        strategy:
            matrix:
                node:
                    - version: 14
                    - env: NODE_OPTIONS=--openssl-legacy-provider
                      version: 20
                os:
                    - ubuntu-latest
                    - macos-latest
    3-context-matrix:
        strategy:
            matrix:
                version: ${{ github.event.client_payload.versions }}
    4-matrix-include:
        strategy:
            matrix:
                animal:
                    - cat
                    - dog
                fruit:
                    - apple
                    - pear
                include:
                    - color: green
                    - animal: cat
                      color: pink
                    - fruit: apple
                      shape: circle
                    - fruit: banana
                    - animal: cat
                      fruit: banana
    5-expanding-configuration:
        strategy:
            matrix:
                include:
                    - node: 16
                      npm: 6
                      os: windows-latest
                node:
                    - 14
                    - 16
                os:
                    - windows-latest
                    - ubuntu-latest
    6-include-only:
        strategy:
            matrix:
                include:
                    - datacenter: site-a
                      site: production
                    - datacenter: site-b
                      site: staging
    7-exclude:
        strategy:
            matrix:
                environment:
                    - staging
                    - production
                exclude:
                    - environment: production
                      os: macos-latest
                      version: 12
                    - os: windows-latest
                      version: 16
                os:
                    - macos-latest
                    - windows-latest
                version:
                    - 12
                    - 14
                    - 16
`,
			model: Workflow{
				Jobs: map[string]Job{
					"0-one-dimensional-matrix": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"version": []any{
										10,
										12,
										14,
									},
								},
							},
						},
					},
					"1-two-dimensional-matrix": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"os": []any{
										"ubuntu-22.04",
										"ubuntu 24.04",
									},
									"version": []any{
										10,
										12,
										14,
									},
								},
							},
						},
					},
					"2-nested-values-matrix": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"os": []any{
										"ubuntu-latest",
										"macos-latest",
									},
									"node": []any{
										map[string]any{
											"version": 14,
										},
										map[string]any{
											"version": 20,
											"env":     "NODE_OPTIONS=--openssl-legacy-provider",
										},
									},
								},
							},
						},
					},
					"3-context-matrix": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"version": "${{ github.event.client_payload.versions }}",
								},
							},
						},
					},
					"4-matrix-include": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"animal": []any{"cat", "dog"},
									"fruit":  []any{"apple", "pear"},
								},
								Include: []map[string]any{
									{"color": "green"},
									{"animal": "cat", "color": "pink"},
									{"fruit": "apple", "shape": "circle"},
									{"fruit": "banana"},
									{"animal": "cat", "fruit": "banana"},
								},
							},
						},
					},
					"5-expanding-configuration": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"os":   []any{"windows-latest", "ubuntu-latest"},
									"node": []any{14, 16},
								},
								Include: []map[string]any{
									{"node": 16, "npm": 6, "os": "windows-latest"},
								},
							},
						},
					},
					"6-include-only": {
						Strategy: Strategy{
							Matrix: Matrix{
								Include: []map[string]any{
									{"datacenter": "site-a", "site": "production"},
									{"datacenter": "site-b", "site": "staging"},
								},
							},
						},
					},
					"7-exclude": {
						Strategy: Strategy{
							Matrix: Matrix{
								Matrix: map[string]any{
									"environment": []any{"staging", "production"},
									"os":          []any{"macos-latest", "windows-latest"},
									"version":     []any{12, 14, 16},
								},
								Exclude: []map[string]any{
									{"environment": "production", "os": "macos-latest", "version": 12},
									{"os": "windows-latest", "version": 16},
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

	edgeCases := map[string]TestCase{
		"non-object concurrency": {
			yaml: `
concurrency: foobar
`,
			model: Workflow{
				Concurrency: Concurrency{
					Group: "foobar",
				},
			},
		},
		"non-array job.needs": {
			yaml: `
jobs:
  example:
    needs: foobar
`,
			model: Workflow{
				Jobs: map[string]Job{
					"example": {
						Needs: []string{"foobar"},
					},
				},
			},
		},
	}

	for name, tt := range edgeCases {
		t.Run(name, func(t *testing.T) {
			var got Workflow
			if err := yaml.Unmarshal([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("Want no error, got %#v", err)
			}

			want := tt.model
			checkWorkflow(t, &got, &want)
		})
	}

	errCases := map[string]TestCase{
		"yaml: invalid 'name' value": {
			yaml: `
name:
- foo
- bar
`,
		},
		"yaml: invalid 'run-name' value": {
			yaml: `
run-name:
- foo
- bar
`,
		},
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
		"yaml: invalid 'permissions.[*]' value": {
			yaml: `
permissions:
    issues: [3, 14]
`,
		},
		"yaml: invalid 'concurrency' value": {
			yaml: `
concurrency: [3, 14]
`,
		},
		"yaml: invalid 'concurrency.cancel-in-progress' value": {
			yaml: `
concurrency:
    cancel-in-progress: foobar
`,
		},
		"yaml: invalid 'concurrency.group' value": {
			yaml: `
concurrency:
    group: [42]
`,
		},
		"yaml: invalid 'defaults' value": {
			yaml: `
defaults: 3
`,
		},
		"yaml: invalid 'defaults.run' value": {
			yaml: `
defaults:
  run: 14
`,
		},
		"yaml: invalid 'defaults.run.shell' value": {
			yaml: `
defaults:
  run:
    shell: [3, 14]
`,
		},
		"yaml: invalid 'defaults.run.working-directory' value": {
			yaml: `
defaults:
  run:
    working-directory: [42]
`,
		},
		"yaml: invalid 'env' value": {
			yaml: `
env: [3, 14]
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
		"yaml: invalid job 'continue-on-error' value": {
			yaml: `
jobs:
  example:
    continue-on-error: foobar
`,
		},
		"yaml: invalid job 'timeout-minutes' value": {
			yaml: `
jobs:
  example:
    timeout-minutes: foobar
`,
		},
		"yaml: invalid job 'if' value": {
			yaml: `
jobs:
  example:
    if: [3, 14]
`,
		},
		"yaml: invalid job 'needs' value": {
			yaml: `
jobs:
  example:
    needs:
      foo: bar
`,
		},
		"yaml: invalid job 'concurrency' value": {
			yaml: `
jobs:
  example:
    concurrency: [3, 14]
`,
		},
		"yaml: invalid job 'concurrency.cancel-in-progress' value": {
			yaml: `
jobs:
  example:
    concurrency:
      cancel-in-progress: foobar
`,
		},
		"yaml: invalid job 'concurrency.group' value": {
			yaml: `
jobs:
  example:
    concurrency:
      group: [42]
`,
		},
		"yaml: invalid job 'defaults' value": {
			yaml: `
jobs:
  example:
    defaults: foobar
`,
		},
		"yaml: invalid job 'defaults.run' value": {
			yaml: `
jobs:
  example:
    defaults:
      run: 14
`,
		},
		"yaml: invalid job 'defaults.run.shell' value": {
			yaml: `
jobs:
  example:
     defaults:
       run:
         shell: [3, 14]
`,
		},
		"yaml: invalid job 'defaults.run.working-directory' value": {
			yaml: `
jobs:
  example:
    defaults:
      run:
        working-directory: [42]
`,
		},
		"yaml: invalid job 'strategy.matrix' value": {
			yaml: `
jobs:
  example:
    strategy:
      matrix: 42
`,
		},
		"yaml: invalid job 'strategy.matrix.include' value": {
			yaml: `
jobs:
  example:
    strategy:
      matrix:
        include: 42
`,
		},
		"yaml: invalid job 'strategy.matrix.include[*]' value": {
			yaml: `
jobs:
  example:
    strategy:
      matrix:
        include:
          - 42
`,
		},
		"yaml: invalid job 'strategy.matrix.exclude' value": {
			yaml: `
jobs:
  example:
    strategy:
      matrix:
        exclude: 42
`,
		},
		"yaml: invalid job 'strategy.matrix.exclude[*]' value": {
			yaml: `
jobs:
  example:
    strategy:
      matrix:
        exclude:
          - 42
`,
		},
		"yaml: invalid job 'strategy.fail-fast' value": {
			yaml: `
jobs:
  example:
    strategy:
      fail-fast: [42]
`,
		},
		"yaml: invalid job 'strategy.max-parallel' value": {
			yaml: `
jobs:
  example:
    strategy:
      max-parallel: [42]
`,
		},
		"yaml: invalid job 'services' value": {
			yaml: `
jobs:
  example:
    services: 42
`,
		},
		"yaml: invalid job 'services.[*]' value": {
			yaml: `
jobs:
  example:
    services:
      example: 3.14
`,
		},
		"yaml: invalid job 'services.[*].image' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        image: [3, 14]
`,
		},
		"yaml: invalid job 'services.[*].credentials' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        credentials: foobar
`,
		},
		"yaml: invalid job 'services.[*].credentials.username' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        credentials:
          username: ["foo", "bar"]
`,
		},
		"yaml: invalid job 'services.[*].credentials.password' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        credentials:
          password:
            foo: bar
`,
		},
		"yaml: invalid job 'services.[*].env' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        env: foobar
`,
		},
		"yaml: invalid job 'services.[*].ports' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        ports: foobar
`,
		},
		"yaml: invalid job 'services.[*].volumes' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        volumes: foobar
`,
		},
		"yaml: invalid job 'services.[*].options' value": {
			yaml: `
jobs:
  example:
    services:
      example:
        options: [3, 14]
`,
		},
		"yaml: invalid job 'outputs' value": {
			yaml: `
jobs:
  example:
    outputs: [3, 14]
`,
		},
		"yaml: invalid job 'permissions' value": {
			yaml: `
jobs:
  example:
    permissions: [3, 14]
`,
		},
		"yaml: invalid job 'env' value": {
			yaml: `
jobs:
  example:
    env: foobar
`,
		},
		"yaml: invalid job 'steps' value": {
			yaml: `
jobs:
  example:
    steps: 42
`,
		},
		"yaml: invalid job 'uses' value": {
			yaml: `
jobs:
  example:
    uses:
      - name: actions/checkout@v4
`,
		},
		"yaml: invalid job 'with' value": {
			yaml: `
jobs:
  example:
    with: foobar
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
    uses: octo-org/example-repo/.github/workflows/called-workflow.yml@main
    with:
      foo: bar
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
		t.Errorf("Unexpected workflow name (got %q, want %q)", got, want)
	}

	if got, want := got.RunName, want.RunName; got != want {
		t.Errorf("Unexpected workflow run-name (got %q, want %q)", got, want)
	}

	checkConcurrency(t, &got.Concurrency, &want.Concurrency)
	checkDefaults(t, &got.Defaults, &want.Defaults)
	checkMap(t, got.Env, want.Env)
	checkJobs(t, got.Jobs, want.Jobs)
	checkPermissions(t, &got.Permissions, &want.Permissions)
}

func checkJobs(t *testing.T, got, want map[string]Job) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of jobs (got %d, want %d)", got, want)
		return
	}

	for name, got := range got {
		want, ok := want[name]
		if !ok {
			t.Errorf("Got job named %q but it is not wanted", name)
			continue
		}

		checkJob(t, &got, &want)
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
		t.Errorf("Unexpected job.name (got %q, want %q)", got, want)
	}

	if got, want := got.ContinueOnError, want.ContinueOnError; got != want {
		t.Errorf("Unexpected job.continue-on-error (got %t, want %t)", got, want)
	}

	if got, want := got.TimeoutMinutes, want.TimeoutMinutes; got != want {
		t.Errorf("Unexpected job.timeout-minutes (got %d, want %d)", got, want)
	}

	if got, want := got.If, want.If; got != want {
		t.Errorf("Unexpected job.if (got %q, want %q)", got, want)
	}

	if got, want := got.Needs, want.Needs; !slices.Equal(got, want) {
		t.Errorf("Unexpected job.needs (got %v, want %v)", got, want)
	}

	checkConcurrency(t, &got.Concurrency, &want.Concurrency)
	checkDefaults(t, &got.Defaults, &want.Defaults)
	checkMap(t, got.Env, want.Env)
	checkEnvironment(t, &got.Environment, &want.Environment)
	checkMap(t, got.Outputs, want.Outputs)
	checkPermissions(t, &got.Permissions, &want.Permissions)
	checkServices(t, got.Services, want.Services)
	checkStrategy(t, &got.Strategy, &want.Strategy)

	/* step-based job */

	checkSteps(t, got.Steps, want.Steps)

	/* uses-based job */

	if got, want := got.Uses, want.Uses; got != want {
		t.Errorf("Unexpected job.uses (got %q, want %q)", got, want)
	}

	checkMap(t, got.With, want.With)
}

func checkConcurrency(t *testing.T, got, want *Concurrency) {
	t.Helper()

	if got, want := got.CancelInProgress, want.CancelInProgress; got != want {
		t.Errorf("Unexpected concurrency.cancel-in-progress (got %t, want %t)", got, want)
	}

	if got, want := got.Group, want.Group; got != want {
		t.Errorf("Unexpected concurrency.group (got %q, want %q)", got, want)
	}
}

func checkDefaults(t *testing.T, got, want *Defaults) {
	t.Helper()

	if got, want := got.Run.Shell, want.Run.Shell; got != want {
		t.Errorf("Unexpected defaults.run.shell (got %q, want %q)", got, want)
	}

	if got, want := got.Run.WorkingDirectory, want.Run.WorkingDirectory; got != want {
		t.Errorf("Unexpected defaults.run.working-directory (got %q, want %q)", got, want)
	}
}

func checkEnvironment(t *testing.T, got, want *Environment) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected environment.name (got %q, want %q)", got, want)
	}

	if got, want := got.Url, want.Url; got != want {
		t.Errorf("Unexpected environment.url (got %q, want %q)", got, want)
	}
}

func checkPermissions(t *testing.T, got, want *Permissions) {
	t.Helper()

	if got, want := got.Actions, want.Actions; got != want {
		t.Errorf("Unexpected permission for 'actions' (got %q, want %q)", got, want)
	}

	if got, want := got.Attestations, want.Attestations; got != want {
		t.Errorf("Unexpected permissions.attestations (got %q, want %q)", got, want)
	}

	if got, want := got.Checks, want.Checks; got != want {
		t.Errorf("Unexpected permissions.checks (got %q, want %q)", got, want)
	}

	if got, want := got.Contents, want.Contents; got != want {
		t.Errorf("Unexpected permissions.contents (got %q, want %q)", got, want)
	}

	if got, want := got.Deployments, want.Deployments; got != want {
		t.Errorf("Unexpected permissions.deployments (got %q, want %q)", got, want)
	}

	if got, want := got.Discussions, want.Discussions; got != want {
		t.Errorf("Unexpected permissions.discussions (got %q, want %q)", got, want)
	}

	if got, want := got.IdToken, want.IdToken; got != want {
		t.Errorf("Unexpected permissions.id-token (got %q, want %q)", got, want)
	}

	if got, want := got.Issues, want.Issues; got != want {
		t.Errorf("Unexpected permissions.issues (got %q, want %q)", got, want)
	}

	if got, want := got.Models, want.Models; got != want {
		t.Errorf("Unexpected permissions.models (got %q, want %q)", got, want)
	}

	if got, want := got.Packages, want.Packages; got != want {
		t.Errorf("Unexpected permissions.packages (got %q, want %q)", got, want)
	}

	if got, want := got.Pages, want.Pages; got != want {
		t.Errorf("Unexpected permissions.pages (got %q, want %q)", got, want)
	}

	if got, want := got.PullRequests, want.PullRequests; got != want {
		t.Errorf("Unexpected permissions.pull-requests (got %q, want %q)", got, want)
	}

	if got, want := got.SecurityEvents, want.SecurityEvents; got != want {
		t.Errorf("Unexpected permissions.security-events (got %q, want %q)", got, want)
	}

	if got, want := got.Statuses, want.Statuses; got != want {
		t.Errorf("Unexpected permissions.statuses (got %q, want %q)", got, want)
	}
}

func checkServices(t *testing.T, got, want map[string]Service) {
	t.Helper()

	if got, want := len(got), len(want); got != want {
		t.Errorf("Unexpected number of services (got %d, want %d)", got, want)
		return
	}

	for name, got := range got {
		want, ok := want[name]
		if !ok {
			t.Errorf("Got service named %q but it is not wanted", name)
			continue
		}

		if got, want := got.Image, want.Image; got != want {
			t.Errorf("Unexpected image for service %q (got %q, want %q)", name, got, want)
		}

		if got, want := got.Credentials.Username, want.Credentials.Username; got != want {
			t.Errorf("Unexpected credentials.username for service %q (got %q, want %q)", name, got, want)
		}

		if got, want := got.Credentials.Password, want.Credentials.Password; got != want {
			t.Errorf("Unexpected credentials.password for service %q (got %q, want %q)", name, got, want)
		}

		if got, want := got.Ports, want.Ports; !slices.Equal(got, want) {
			t.Errorf("Unexpected ports for service %q (got %v, want %v)", name, got, want)
		}

		if got, want := got.Volumes, want.Volumes; !slices.Equal(got, want) {
			t.Errorf("Unexpected volumes for service %q (got %v, want %v)", name, got, want)
		}

		if got, want := got.Options, want.Options; got != want {
			t.Errorf("Unexpected options for service %q (got %q, want %q)", name, got, want)
		}

		checkMap(t, got.Env, want.Env)
	}

	for name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("Want service named %q but it is not present", name)
		}
	}
}

func checkStrategy(t *testing.T, got, want *Strategy) {
	t.Helper()

	if got, want := got.Matrix.Matrix, want.Matrix.Matrix; !reflect.DeepEqual(got, want) {
		t.Errorf("Strategy matrix are not equal (got %+v, want %+v)", got, want)
	}

	if got, want := got.Matrix.Include, want.Matrix.Include; !reflect.DeepEqual(got, want) {
		t.Errorf("Strategy matrix.include are not equal (got %+v, want %+v)", got, want)
	}

	if got, want := got.Matrix.Exclude, want.Matrix.Exclude; !reflect.DeepEqual(got, want) {
		t.Errorf("Strategy matrix.exclude are not equal (got %+v, want %+v)", got, want)
	}

	if got, want := got.FailFast, want.FailFast; got != want {
		t.Errorf("Unexpected fail-fast for strategy (got %t, want %t)", got, want)
	}

	if got, want := got.MaxParallel, want.MaxParallel; got != want {
		t.Errorf("Unexpected max-parallel for strategy (got %d, want %d)", got, want)
	}
}
