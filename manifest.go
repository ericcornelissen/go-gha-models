// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest is a model of a GitHub Actions Action manifest.
type Manifest struct {
	Name        string   `yaml:"name"`
	Author      string   `yaml:"author"`
	Description string   `yaml:"description"`
	Branding    Branding `yaml:"branding"`

	Inputs  map[string]Input  `yaml:"inputs"`
	Outputs map[string]Output `yaml:"outputs"`

	Runs Runs `yaml:"runs"`
}

// Branding is a model of an Action [Manifest]'s `branding:` object.
type Branding struct {
	Color string `yaml:"color"`
	Icon  string `yaml:"icon"`
}

// Input is a model of an Action [Manifest]'s `inputs:`.
type Input struct {
	Default            string `yaml:"default"`
	DeprecationMessage string `yaml:"deprecationMessage"`
	Description        string `yaml:"description"`
	Required           bool   `yaml:"required"`
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

	Steps []Step `yaml:"steps"`

	/* using: docker */

	Image          string            `yaml:"image"`
	Args           []string          `yaml:"args"`
	Env            map[string]string `yaml:"env"`
	PreEntrypoint  string            `yaml:"pre-entrypoint"`
	Entrypoint     string            `yaml:"entrypoint"`
	PostEntrypoint string            `yaml:"post-entrypoint"`

	/* using: node */

	Pre    string `yaml:"pre"`
	PreIf  string `yaml:"pre-if"`
	Main   string `yaml:"main"`
	Post   string `yaml:"post"`
	PostIf string `yaml:"post-if"`
}

// ParseManifest parses a GitHub Actions Action manifest into a [Manifest].
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("could not parse manifest: %v", err)
	}

	return manifest, nil
}
