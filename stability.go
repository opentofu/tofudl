// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"fmt"
)

// Stability describes the minimum stability to download.
type Stability string

const (
	// StabilityAlpha accepts any stability.
	StabilityAlpha Stability = "alpha"
	// StabilityBeta accepts beta, release candidate and stable versions.
	StabilityBeta Stability = "beta"
	// StabilityRC accepts release candidate and stable versions.
	StabilityRC Stability = "rc"
	// StabilityStable accepts only stable versions.
	StabilityStable Stability = ""
)

// AsInt returns a numeric representation of the stability for easier comparison.
func (s Stability) AsInt() int {
	switch s {
	case StabilityStable:
		return 0
	case StabilityRC:
		return -1
	case StabilityBeta:
		return -2
	case StabilityAlpha:
		return -3
	default:
		panic(s.Validate())
	}
}

// Matches returns true if the provided version matches the current stability or higher.
func (s Stability) Matches(version Version) bool {
	return version.Stability().AsInt() >= s.AsInt()
}

// Validate returns an error if the stability is not one of the listed stabilities.
func (s Stability) Validate() error {
	switch s {
	case StabilityStable:
		return nil
	case StabilityRC:
		return nil
	case StabilityBeta:
		return nil
	case StabilityAlpha:
		return nil
	default:
		return fmt.Errorf("invalid stability value: %s", s)
	}
}

// StabilityValues returns all supported values for Stability excluding StabilityStable.
func StabilityValues() []Stability {
	return []Stability{
		StabilityRC,
		StabilityBeta,
		StabilityAlpha,
	}
}
