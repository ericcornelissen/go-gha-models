// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest is a model of a GitHub Actions Action manifest.
type Manifest struct {
	Name        string   `yaml:"name"`
	Author      string   `yaml:"author,omitempty"`
	Description string   `yaml:"description"`
	Branding    Branding `yaml:"branding,omitempty"`

	Inputs  map[string]Input  `yaml:"inputs,omitempty"`
	Outputs map[string]Output `yaml:"outputs,omitempty"`

	Runs Runs `yaml:"runs"`
}

// Branding is a model of an Action [Manifest]'s `branding:` object.
type Branding struct {
	Color string `yaml:"color"`
	Icon  string `yaml:"icon"`
}

// Input is a model of an Action [Manifest]'s `inputs:`.
type Input struct {
	Description        string `yaml:"description"`
	Default            string `yaml:"default,omitempty"`
	Required           bool   `yaml:"required,omitempty"`
	DeprecationMessage string `yaml:"deprecationMessage,omitempty"`
}

// Output is a model of an Action [Manifest]'s `outputs:`.
type Output struct {
	Description string `yaml:"description"`
	Value       string `yaml:"value"`
}

// Runs is a model of an Action [Manifest]'s `runs:` object.
type Runs struct {
	Using string `yaml:"using"`

	/* using: composite */

	Steps []Step `yaml:"steps,omitempty"`

	/* using: docker */

	Image          string            `yaml:"image,omitempty"`
	PreEntrypoint  string            `yaml:"pre-entrypoint,omitempty"`
	Entrypoint     string            `yaml:"entrypoint,omitempty"`
	PostEntrypoint string            `yaml:"post-entrypoint,omitempty"`
	Args           []string          `yaml:"args,omitempty"`
	Env            map[string]string `yaml:"env,omitempty"`

	/* using: node */

	Pre    string `yaml:"pre,omitempty"`
	PreIf  string `yaml:"pre-if,omitempty"`
	Main   string `yaml:"main,omitempty"`
	Post   string `yaml:"post,omitempty"`
	PostIf string `yaml:"post-if,omitempty"`
}

// ParseManifest parses a GitHub Actions Action manifest into a [Manifest].
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("could not parse manifest: %v", err)
	}

	return manifest, nil
}
