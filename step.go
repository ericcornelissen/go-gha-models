// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Step is a model of a workflow/manifest job step.
type Step struct {
	With  map[string]string `yaml:"with"`
	Env   map[string]string `yaml:"env"`
	Name  string            `yaml:"name"`
	Run   string            `yaml:"run"`
	Shell string            `yaml:"shell"`
	Uses  Uses              `yaml:"uses"`
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

func (u Uses) MarshalYAML() (interface{}, error) {
	if u.Name == "" && u.Ref == "" {
		return "", nil
	}

	if u.Name == "" && u.Ref != "" {
		return nil, errors.New("missing 'name' value")
	}

	n := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: u.Name,
		LineComment: strings.Map(func(r rune) rune {
			if !unicode.IsPrint(r) {
				return -1
			}

			return r
		}, u.Annotation),
	}

	if u.Ref != "" {
		n.Value = n.Value + "@" + u.Ref
	}

	return n, nil
}
