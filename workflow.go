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
			return fmt.Errorf("invalid permissions value %s", n.Value)
		}
	case yaml.MappingNode:
		var perms map[string]string
		_ = n.Decode(&perms)

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
		return fmt.Errorf("invalid permissions %s", n.Value)
	}

	return nil
}

func (p Permissions) MarshalYAML() (interface{}, error) {
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

// ParseWorkflow parses a GitHub Actions workflow into a [Workflow].
func ParseWorkflow(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return workflow, fmt.Errorf("could not parse workflow: %v", err)
	}

	return workflow, nil
}
