// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"
	"maps"
	"sort"

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
	Needs           Needs              `yaml:"needs,omitempty"`
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
	CancelInProgress string `yaml:"cancel-in-progress,omitempty"`
	Group            string `yaml:"group,omitempty"`
}

func (c *Concurrency) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		c.Group = n.Value
	case yaml.MappingNode:
		var concurrency map[string]any
		_ = n.Decode(&concurrency)

		if v, ok := concurrency["cancel-in-progress"]; ok {
			switch v := v.(type) {
			case bool:
				c.CancelInProgress = fmt.Sprintf("%t", v)
			case string:
				c.CancelInProgress = v
			default:
				return fmt.Errorf("invalid concurrency.cancel-in-progress %v", n.Kind)
			}
		}

		if v, ok := concurrency["group"]; ok {
			v, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid concurrency.group %v", n.Kind)
			}

			c.Group = v
		}
	default:
		return fmt.Errorf("invalid concurrency %v", n.Kind)
	}

	return nil
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

type Needs []string

func (l *Needs) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		*l = []string{n.Value}
	case yaml.SequenceNode:
		var list []string
		if err := n.Decode(&list); err != nil {
			return err
		}

		*l = list
	default:
		return fmt.Errorf("invalid job.needs %v", n.Kind)
	}

	return nil
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

// Service is a model of a GitHub Actions `services:` object.
type Service struct {
	Image       string             `yaml:"image"`
	Credentials ServiceCredentials `yaml:"credentials,omitempty"`
	Env         map[string]string  `yaml:"env,omitempty"`
	Ports       Ports              `yaml:"ports,omitempty"`
	Volumes     []string           `yaml:"volumes,omitempty"`
	Options     string             `yaml:"options,omitempty"`
}

type Ports []string

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
type Matrix []map[string]any

func (m *Matrix) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("invalid matrix %q", n.Value)
	}

	var raw map[string]any
	_ = n.Decode(&raw)

	var include []map[string]any
	if v, ok := raw["include"]; ok {
		delete(raw, "include")

		tmp, ok := v.([]any)
		if !ok {
			return fmt.Errorf("invalid matrix.include %v", v)
		}

		include = make([]map[string]any, len(tmp))
		for k, v := range tmp {
			if v, ok := v.(map[string]any); !ok {
				return fmt.Errorf("invalid matrix.include entry %v", v)
			} else {
				include[k] = v
			}
		}
	}

	var exclude []map[string]any
	if v, ok := raw["exclude"]; ok {
		delete(raw, "exclude")

		tmp, ok := v.([]any)
		if !ok {
			return fmt.Errorf("invalid matrix.exclude %v", v)
		}

		exclude = make([]map[string]any, len(tmp))
		for k, v := range tmp {
			if v, ok := v.(map[string]any); !ok {
				return fmt.Errorf("invalid matrix.exclude entry %v", v)
			} else {
				exclude[k] = v
			}
		}
	}

	result := []map[string]any{}
	if len(raw) != 0 {
		result = append(result, map[string]any{})
	}
	for k, tmp := range raw {
		var vs []any
		switch tmp := tmp.(type) {
		case []any:
			vs = tmp
		case string:
			vs = []any{tmp}
		default:
			return fmt.Errorf("invalid matrix entry %q", k)
		}

		matrix := []map[string]any{}
		for _, v := range vs {
			for _, src := range result {
				dest := map[string]any{}
				matrix = append(matrix, dest)

				maps.Copy(dest, src)
				dest[k] = v
			}
		}

		result = matrix
	}

	extend := []map[string]any{}
Loop_include:
	for _, include := range include {
		for _, entry := range result {
			found := entry
			for k, want := range entry {
				if got, ok := include[k]; !ok || got != want {
					found = nil
				}
			}

			if found != nil {
				for k, v := range include {
					found[k] = v
				}

				continue Loop_include
			}
		}

		extend = append(extend, include)
	}
	result = append(result, extend...)

	for _, exclude := range exclude {
		omit := []int{}
		for i, entry := range result {
			found := true
			for k, want := range exclude {
				if got, ok := entry[k]; !ok || got != want {
					found = false
				}
			}

			if found {
				omit = append(omit, i)
			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(omit)))
		for _, i := range omit {
			result = append(result[0:i], result[i+1:]...)
		}
	}

	*m = result

	return nil
}

// ParseWorkflow parses a GitHub Actions workflow into a [Workflow].
func ParseWorkflow(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return workflow, fmt.Errorf("could not parse workflow: %v", err)
	}

	return workflow, nil
}
