// SPDX-License-Identifier: BSD-2-Clause-Patent

package gha

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest is a model of a GitHub Actions Action manifest.
type Manifest struct {
	Runs ManifestRuns `yaml:"runs"`
}

// ManifestRuns is a model of an Action manifest's `runs:` object.
type ManifestRuns struct {
	Using string `yaml:"using"`
	Steps []Step `yaml:"steps"`
}

// ParseManifest parses a GitHub Actions Action manifest into a [Manifest] struct.
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("could not parse manifest: %v", err)
	}

	return manifest, nil
}
