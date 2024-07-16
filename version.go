// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"fmt"
	"regexp"
	"strconv"
)

// VersionWithArtifacts is a version and the list of artifacts belonging to that version.
type VersionWithArtifacts struct {
	ID    Version  `json:"id"`
	Files []string `json:"files"`
}

// Version describes a version number with this project's version and stability understanding.
type Version string

var versionRe = regexp.MustCompile(`^(?P<major>[0-9]+)\.(?P<minor>[0-9]+)\.(?P<patch>[0-9]+)(|-(?P<stability>alpha|beta|rc)(?P<stabilityver>[0-9]+))$`)

// Validate checks if the version is valid
func (v Version) Validate() error {
	if !versionRe.MatchString(string(v)) {
		return &InvalidVersionError{v}
	}
	return nil
}

// Major returns the major version. The version must be valid or this function will panic.
func (v Version) Major() int {
	return v.parse().major
}

// Minor returns the minor version. The version must be valid or this function will panic.
func (v Version) Minor() int {
	return v.parse().minor
}

// Patch returns the patch version. The version must be valid or this function will panic.
func (v Version) Patch() int {
	return v.parse().patch
}

// Stability returns the stability string for the version. The version must be valid or this function will panic.
func (v Version) Stability() Stability {
	return v.parse().stability
}

// StabilityVer returns the stability version number for the version. The version must be valid or this function will
// panic.
func (v Version) StabilityVer() int {
	return v.parse().stabilityVer
}

// Compare returns 1 if the current version is larger than the other, -1 if it is smaller, 0 otherwise.
func (v Version) Compare(other Version) int {
	parsedThis := v.parse()
	parsedOther := other.parse()
	thisStabilityInt := parsedThis.stability.AsInt()
	otherStabilityInt := parsedOther.stability.AsInt()

	if parsedThis.major > parsedOther.major {
		return 1
	} else if parsedThis.major < parsedOther.major {
		return -1
	}
	if parsedThis.minor > parsedOther.minor {
		return 1
	} else if parsedThis.minor < parsedOther.minor {
		return -1
	}
	if parsedThis.patch > parsedOther.patch {
		return 1
	} else if parsedThis.patch < parsedOther.patch {
		return -1
	}
	if thisStabilityInt > otherStabilityInt {
		return 1
	} else if thisStabilityInt < otherStabilityInt {
		return -1
	}
	if parsedThis.stabilityVer > parsedOther.stabilityVer {
		return 1
	} else if parsedThis.stabilityVer < parsedOther.stabilityVer {
		return -1
	}
	return 0
}

func (v Version) parse() parsedVersion {
	subMatch := versionRe.FindStringSubmatch(string(v))
	if len(subMatch) == 0 {
		panic(fmt.Errorf("invalid version: %v", v))
	}
	result := map[string]any{}
	for i, name := range versionRe.SubexpNames() {
		result[name] = subMatch[i]
	}
	var err error
	for _, name := range []string{"major", "minor", "patch"} {
		result[name], err = strconv.Atoi(result[name].(string))
		if err != nil {
			panic(fmt.Errorf("invalid version: %w", err))
		}
	}

	stabilityVer := -1
	if result["stabilityver"] != "" {
		stabilityVer, err = strconv.Atoi(result["stabilityver"].(string))
		if err != nil {
			panic(fmt.Errorf("invalid version: %w", err))
		}
	}

	return parsedVersion{
		major:        result["major"].(int),
		minor:        result["minor"].(int),
		patch:        result["patch"].(int),
		stability:    Stability(result["stability"].(string)),
		stabilityVer: stabilityVer,
	}
}

type parsedVersion struct {
	major        int
	minor        int
	patch        int
	stability    Stability
	stabilityVer int
}
