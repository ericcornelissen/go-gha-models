// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Step is a model of a workflow/manifest job step.
type Step struct {
	Name             string            `yaml:"name,omitempty"`
	Uses             Uses              `yaml:"uses,omitempty"`
	Id               string            `yaml:"id,omitempty"`
	If               string            `yaml:"if,omitempty"`
	ContinueOnError  bool              `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes   uint              `yaml:"timeout-minutes,omitempty"`
	WorkingDirectory string            `yaml:"working-directory,omitempty"`
	Shell            string            `yaml:"shell,omitempty"`
	Run              string            `yaml:"run,omitempty"`
	With             map[string]string `yaml:"with,omitempty"`
	Env              map[string]string `yaml:"env,omitempty"`
}

// Uses is a model of a step `uses:` value.
type Uses struct {
	// Name is the name of the Action that is used. Typically <owner>/<repository>.
	Name string

	// Ref is the git reference used for the Action. Typically a tag ref, branch ref, or commit SHA.
	Ref string

	// Annotation is the comment after the `uses:` value, if any.
	Annotation string
}

// IsLocal reports whether the uses value is for a local or remote Action.
func (u *Uses) IsLocal() bool {
	name := u.Name
	return len(name) > 0 && name[0] == '.'
}

func (u *Uses) String() string {
	if len(u.Ref) == 0 {
		return u.Name
	}

	return u.Name + "@" + u.Ref
}

func (u *Uses) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("cannot unmarshal %s into a gha.Uses struct", n.Tag)
	}

	if n.Value == "" {
		return nil
	}

	i := strings.LastIndex(n.Value, "@")
	if i == 0 || i == len(n.Value)-1 {
		return fmt.Errorf("invalid `uses` value (%q)", n.Value)
	}

	if i > 0 {
		u.Name = n.Value[:i]
		u.Ref = n.Value[i+1:]
	} else {
		u.Name = n.Value
	}

	u.Annotation = strings.TrimLeft(n.LineComment, "# ")

	return nil
}
