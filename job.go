// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// Job is a model of a workflow job.
type Job struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

// Step is a model of a workflow/manifest job step.
type Step struct {
	With           map[string]string
	Env            map[string]string
	Name           string
	Run            string
	Shell          string
	Uses           string
	UsesAnnotation string
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

// ParseUses parses a Github Actions workflow job step's `uses:` value.
func ParseUses(step *Step) (Uses, error) {
	var uses Uses

	i := strings.LastIndex(step.Uses, "@")
	if i <= 0 || i >= len(step.Uses)-1 {
		return uses, errors.New("step has no or invalid `uses`")
	}

	uses.Name = step.Uses[:i]
	uses.Ref = step.Uses[i+1:]
	uses.Annotation = step.UsesAnnotation
	return uses, nil
}

func (step *Step) UnmarshalYAML(node *yaml.Node) error {
	for i := range node.Content {
		if i%2 == 1 {
			continue
		}

		k, v := node.Content[i].Value, node.Content[i+1]

		var err error
		switch k {
		case "env":
			err = v.Decode(&step.Env)
		case "name":
			step.Name = v.Value
		case "run":
			step.Run = v.Value
		case "shell":
			step.Shell = v.Value
		case "uses":
			step.Uses = v.Value
			step.UsesAnnotation = strings.TrimLeft(v.LineComment, "# ")
		case "with":
			err = v.Decode(&step.With)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
