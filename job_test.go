// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"testing"
)

func TestUses(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		type TestCase struct {
			step Step
			want Uses
		}

		testCases := map[string]TestCase{
			"Full version tag": {
				step: Step{
					Uses: "foobar@v1.2.3",
				},
				want: Uses{
					Name: "foobar",
					Ref:  "v1.2.3",
				},
			},
			"Major version tag": {
				step: Step{
					Uses: "hello-world@v2",
				},
				want: Uses{
					Name: "hello-world",
					Ref:  "v2",
				},
			},
			"Full SHA": {
				step: Step{
					Uses: "actions/checkout@2a08af6587712680d7d485082f61ed6cdb72280a",
				},
				want: Uses{
					Name: "actions/checkout",
					Ref:  "2a08af6587712680d7d485082f61ed6cdb72280a",
				},
			},
			"Unconventional tag (no 'v' prefix)": {
				step: Step{
					Uses: "actions/upload-artifact@3.1.4",
				},
				want: Uses{
					Name: "actions/upload-artifact",
					Ref:  "3.1.4",
				},
			},
			"short name": {
				step: Step{
					Uses: "a@3.1.4",
				},
				want: Uses{
					Name: "a",
					Ref:  "3.1.4",
				},
			},
			"1 character version": {
				step: Step{
					Uses: "actions/download-artifact@7",
				},
				want: Uses{
					Name: "actions/download-artifact",
					Ref:  "7",
				},
			},
			"with comment": {
				step: Step{
					Uses:           "actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b",
					UsesAnnotation: "v4",
				},
				want: Uses{
					Name:       "actions/checkout",
					Ref:        "0ad4b8fadaa221de15dcec353f45205ec38ea70b",
					Annotation: "v4",
				},
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				uses, err := ParseUses(&tt.step)
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}

				if got, want := uses.Name, tt.want.Name; got != want {
					t.Fatalf("Unexpected name (got %q, want %q)", got, want)
				}

				if got, want := uses.Ref, tt.want.Ref; got != want {
					t.Fatalf("Unexpected ref (got %q, want %q)", got, want)
				}

				if got, want := uses.Annotation, tt.want.Annotation; got != want {
					t.Fatalf("Unexpected annotation (got %q, want %q)", got, want)
				}
			})
		}
	})

	t.Run("Error", func(t *testing.T) {
		type TestCase struct {
			step Step
		}

		testCases := map[string]TestCase{
			"No 'uses' value": {
				step: Step{},
			},
			"Invalid 'uses' value": {
				step: Step{
					Uses: "foobar",
				},
			},
			"Missing version": {
				step: Step{
					Uses: "foobar@",
				},
			},
			"Missing name": {
				step: Step{
					Uses: "@v1.2.3",
				},
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := ParseUses(&tt.step)
				if err == nil {
					t.Fatal("Expected an error, got none")
				}
			})
		}
	})
}

func CheckJob(t *testing.T, n string, got, want Job) {
	t.Helper()

	if got, want := got.Name, want.Name; got != want {
		t.Errorf("Unexpected name for job %q (got %q, want %q)", n, got, want)
	}

	if got, want := len(got.Steps), len(want.Steps); got != want {
		t.Errorf("Unexpected number of steps for job %q (got %d, want %d)", n, got, want)
		return
	}

	for i, step := range got.Steps {
		want := want.Steps[i]

		if got, want := step.Name, want.Name; got != want {
			t.Errorf("Unexpected name for job %q step %d (got %q, want %q)", n, i, got, want)
		}

		if got, want := step.Run, want.Run; got != want {
			t.Errorf("Unexpected run for job %q step %d (got %q, want %q)", n, i, got, want)
		}

		if got, want := step.Shell, want.Shell; got != want {
			t.Errorf("Unexpected shell for job %q step %d (got %q, want %q)", n, i, got, want)
		}

		if got, want := step.Uses, want.Uses; got != want {
			t.Errorf("Unexpected uses for job %q step %d (got %q, want %q)", n, i, got, want)
		}

		if got, want := step.UsesAnnotation, want.UsesAnnotation; got != want {
			t.Errorf("Unexpected uses comment for job %q step %d (got %q, want %q)", n, i, got, want)
		}

		if got, want := step.With["script"], want.With["script"]; got != want {
			t.Errorf("Unexpected with for job %q step %d (got %q, want %q)", n, i, got, want)
		}
	}
}
