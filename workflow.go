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
	Permissions Permissions       `yaml:"permissions,omitempty"`
	Concurrency Concurrency       `yaml:"concurrency,omitempty"`
	Defaults    Defaults          `yaml:"defaults,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Jobs        map[string]Job    `yaml:"jobs"`
}

// Job is a model of a GitHub Actions workflow job.
type Job struct {
	Name            string             `yaml:"name,omitempty"`
	Environment     Environment        `yaml:"environment,omitempty"`
	ContinueOnError bool               `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes  int                `yaml:"timeout-minutes,omitempty"`
	If              string             `yaml:"if,omitempty"`
	Needs           []string           `yaml:"needs,omitempty"`
	Concurrency     Concurrency        `yaml:"concurrency,omitempty"`
	Defaults        Defaults           `yaml:"defaults,omitempty"`
	Strategy        Strategy           `yaml:"strategy,omitempty"`
	Services        map[string]Service `yaml:"services,omitempty"`
	Outputs         map[string]string  `yaml:"outputs,omitempty"`
	Permissions     Permissions        `yaml:"permissions,omitempty"`
	Env             map[string]string  `yaml:"env,omitempty"`

	/* step-based job */

	Steps []Step `yaml:"steps,omitempty"`

	/* uses-based job */

	Uses string            `yaml:"uses,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
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

// DefaultsRun is a model of a GitHub Actions `defaults.run:` object.
type DefaultsRun struct {
	Shell            string `yaml:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

// Environment is a model of a GitHub Actions `environment:` object.
type Environment struct {
	Name string `yaml:"name,omitempty"`
	Url  string `yaml:"url,omitempty"`
}

func (e *Environment) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		e.Name = n.Value
	case yaml.MappingNode:
		var env map[string]string
		_ = n.Decode(&env)

		if v, ok := env["name"]; ok {
			e.Name = v
		}
		if v, ok := env["url"]; ok {
			e.Url = v
		}
	default:
		return fmt.Errorf("invalid environment %q", n.Value)
	}

	return nil
}

func (e Environment) MarshalYAML() (any, error) {
	n := yaml.Node{}
	if e.Url == "" {
		n.Kind = yaml.ScalarNode
		n.Tag = "!!str"
		n.Value = e.Name
	} else {
		env := make(map[string]string, 2)
		env["name"] = e.Name
		env["url"] = e.Url
		_ = n.Encode(env)
	}

	return n, nil
}

// Permissions is a model of a GitHub Actions `permissions:` object.
type Permissions struct {
	Actions        string `yaml:"actions,omitempty"`
	Attestations   string `yaml:"attestations,omitempty"`
	Checks         string `yaml:"checks,omitempty"`
	Contents       string `yaml:"contents,omitempty"`
	Deployments    string `yaml:"deployments,omitempty"`
	Discussions    string `yaml:"discussions,omitempty"`
	IdToken        string `yaml:"id-token,omitempty"`
	Issues         string `yaml:"issues,omitempty"`
	Models         string `yaml:"models,omitempty"`
	Packages       string `yaml:"packages,omitempty"`
	Pages          string `yaml:"pages,omitempty"`
	PullRequests   string `yaml:"pull-requests,omitempty"`
	SecurityEvents string `yaml:"security-events,omitempty"`
	Statuses       string `yaml:"statuses,omitempty"`
}

func (p *Permissions) UnmarshalYAML(n *yaml.Node) error {
	all := func(s string) {
		p.Actions = s
		p.Attestations = s
		p.Checks = s
		p.Contents = s
		p.Deployments = s
		p.Discussions = s
		p.IdToken = s
		p.Issues = s
		p.Models = s
		p.Packages = s
		p.Pages = s
		p.PullRequests = s
		p.SecurityEvents = s
		p.Statuses = s
	}

	switch n.Kind {
	case yaml.ScalarNode:
		switch n.Value {
		case "read-all":
			all("read")
		case "write-all":
			all("write")
		default:
			return fmt.Errorf("invalid permissions value %q", n.Value)
		}
	case yaml.MappingNode:
		var perms map[string]string
		if err := n.Decode(&perms); err != nil {
			return err
		}

		all("none")
		if v, ok := perms["actions"]; ok {
			p.Actions = v
		}
		if v, ok := perms["attestations"]; ok {
			p.Attestations = v
		}
		if v, ok := perms["checks"]; ok {
			p.Checks = v
		}
		if v, ok := perms["contents"]; ok {
			p.Contents = v
		}
		if v, ok := perms["deployments"]; ok {
			p.Deployments = v
		}
		if v, ok := perms["discussions"]; ok {
			p.Discussions = v
		}
		if v, ok := perms["id-token"]; ok {
			p.IdToken = v
		}
		if v, ok := perms["issues"]; ok {
			p.Issues = v
		}
		if v, ok := perms["models"]; ok {
			p.Models = v
		}
		if v, ok := perms["issues"]; ok {
			p.Issues = v
		}
		if v, ok := perms["models"]; ok {
			p.Models = v
		}
		if v, ok := perms["issues"]; ok {
			p.Issues = v
		}
		if v, ok := perms["packages"]; ok {
			p.Packages = v
		}
		if v, ok := perms["pages"]; ok {
			p.Pages = v
		}
		if v, ok := perms["pull-requests"]; ok {
			p.PullRequests = v
		}
		if v, ok := perms["security-events"]; ok {
			p.SecurityEvents = v
		}
		if v, ok := perms["statuses"]; ok {
			p.Statuses = v
		}
	default:
		return fmt.Errorf("invalid permissions %q", n.Value)
	}

	return nil
}

func (p Permissions) MarshalYAML() (any, error) {
	all := func(s string) bool {
		return p.Actions == s && p.Attestations == s && p.Checks == s && p.Contents == s && p.Deployments == s && p.Discussions == s && p.IdToken == s && p.Issues == s && p.Models == s && p.Packages == s && p.Pages == s && p.PullRequests == s && p.SecurityEvents == s && p.Statuses == s
	}

	n := yaml.Node{}
	switch {
	case all("read"):
		n.Kind = yaml.ScalarNode
		n.Tag = "!!str"
		n.Value = "read-all"
	case all("write"):
		n.Kind = yaml.ScalarNode
		n.Tag = "!!str"
		n.Value = "write-all"
	default:
		perms := make(map[string]string, 14)
		if v := p.Actions; v != "none" {
			perms["actions"] = v
		}
		if v := p.Attestations; v != "none" {
			perms["attestations"] = v
		}
		if v := p.Checks; v != "none" {
			perms["checks"] = v
		}
		if v := p.Contents; v != "none" {
			perms["contents"] = v
		}
		if v := p.Deployments; v != "none" {
			perms["deployments"] = v
		}
		if v := p.Discussions; v != "none" {
			perms["discussions"] = v
		}
		if v := p.IdToken; v != "none" {
			perms["id-token"] = v
		}
		if v := p.Issues; v != "none" {
			perms["issues"] = v
		}
		if v := p.Models; v != "none" {
			perms["models"] = v
		}
		if v := p.Packages; v != "none" {
			perms["packages"] = v
		}
		if v := p.Pages; v != "none" {
			perms["pages"] = v
		}
		if v := p.PullRequests; v != "none" {
			perms["pull-requests"] = v
		}
		if v := p.SecurityEvents; v != "none" {
			perms["security-events"] = v
		}
		if v := p.Statuses; v != "none" {
			perms["statuses"] = v
		}
		_ = n.Encode(perms)
	}

	return n, nil
}

// Service is a model of a GitHub Actions `services:` object.
type Service struct {
	Image       string             `yaml:"image"`
	Credentials ServiceCredentials `yaml:"credentials,omitempty"`
	Env         map[string]string  `yaml:"env,omitempty"`
	Ports       []int              `yaml:"ports,omitempty"`
	Volumes     []string           `yaml:"volumes,omitempty"`
	Options     string             `yaml:"options,omitempty"`
}

// ServiceCredentials is a model of a GitHub Actions `services.credentials:` object.
type ServiceCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Strategy is a model of a GitHub Actions `strategy:` object.
type Strategy struct {
	Matrix      Matrix `yaml:"matrix,omitempty"`
	FailFast    bool   `yaml:"fail-fast,omitempty"`
	MaxParallel int    `yaml:"max-parallel,omitempty"`
}

// Matrix is a model of a GitHub Actions `strategy.matrix:` object.
type Matrix struct {
	Matrix  map[string]any
	Include []map[string]any
	Exclude []map[string]any
}

func (m *Matrix) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("invalid matrix %q", n.Value)
	}

	var matrix map[string]any
	_ = n.Decode(&matrix)

	if include, ok := matrix["include"]; ok {
		tmp, ok := include.([]any)
		if !ok {
			return fmt.Errorf("invalid matrix.include %q", n.Value)
		}

		m.Include = make([]map[string]any, len(tmp))
		for i, tmp := range tmp {
			include, ok := tmp.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid matrix.include %q", n.Value)
			}

			m.Include[i] = include
		}

		delete(matrix, "include")
	}

	if exclude, ok := matrix["exclude"]; ok {
		tmp, ok := exclude.([]any)
		if !ok {
			return fmt.Errorf("invalid matrix.exclude %q", n.Value)
		}

		m.Exclude = make([]map[string]any, len(tmp))
		for i, tmp := range tmp {
			exclude, ok := tmp.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid matrix.exclude %q", n.Value)
			}

			m.Exclude[i] = exclude
		}

		delete(matrix, "exclude")
	}

	if len(matrix) != 0 {
		m.Matrix = matrix
	}

	return nil
}

func (m Matrix) MarshalYAML() (any, error) {
	matrix := make(map[string]any, len(m.Matrix))
	for k, v := range m.Matrix {
		matrix[k] = v
	}

	if include := m.Include; len(include) != 0 {
		matrix["include"] = include
	}
	if exclude := m.Exclude; len(exclude) != 0 {
		matrix["exclude"] = exclude
	}

	n := yaml.Node{}
	err := n.Encode(matrix)
	return n, err
}

// ParseWorkflow parses a GitHub Actions workflow into a [Workflow].
func ParseWorkflow(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return workflow, fmt.Errorf("could not parse workflow: %v", err)
	}

	return workflow, nil
}
