// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Workflow is a model of a GitHub Actions workflow.
type Workflow struct {
	Name        string            `yaml:"name,omitempty"`
	RunName     string            `yaml:"run-name,omitempty"`
	Concurrency Concurrency       `yaml:"concurrency,omitempty"`
	Defaults    Defaults          `yaml:"defaults,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Jobs        map[string]Job    `yaml:"jobs"`
}

// Concurrency is a model of a GitHub Actions `concurrency:` object.
type Concurrency struct {
	CancelInProgress bool   `yaml:"cancel-in-progress,omitempty"`
	Group            string `yaml:"group,omitempty"`
}

// Defaults is a model of a GitHub Actions `defaults:` object.
type Defaults struct {
	Run DefaultsRun `yaml:"run,omitempty"`
}

type DefaultsRun struct {
	Shell            string `yaml:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

// Job is a model of a workflow job.
type Job struct {
	Name  string `yaml:"name,omitempty"`
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
