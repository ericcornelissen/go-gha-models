// SPDX-License-Identifier: BSD-2-Clause

package gha

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest is a model of a GitHub Actions Action manifest.
type Manifest struct {
	Runs Runs `yaml:"runs"`
}

// Runs is a model of an Action manifest's `runs:` object.
type Runs struct {
	Using string `yaml:"using"`
	Steps []Step `yaml:"steps"`
}

// ParseManifest parses a GitHub Actions Action manifest into a [Manifest].
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("could not parse manifest: %v", err)
	}

	return manifest, nil
}
