// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Workflow is a model of a GitHub Actions workflow.
type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}

// Job is a model of a workflow job.
type Job struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

// ParseWorkflow parses a GitHub Actions workflow into a [Workflow].
func ParseWorkflow(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return workflow, fmt.Errorf("could not parse workflow: %v", err)
	}

	return workflow, nil
}
