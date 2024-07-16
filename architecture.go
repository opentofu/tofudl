// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"regexp"
	"runtime"
)

// Architecture describes the architecture to download OpenTofu for. It defaults to the current system architecture.
type Architecture string

const (
	// ArchitectureAuto is the default value and defaults to downloading OpenTofu for the current architecture.
	ArchitectureAuto Architecture = ""
	// Architecture386 describes the 32-bit Intel CPU architecture.
	Architecture386 Architecture = "386"
	// ArchitectureAMD64 describes the 64-bit Intel/AMD CPU architecture.
	ArchitectureAMD64 Architecture = "amd64"
	// ArchitectureARM describes the 32-bit ARM (v7) architecture.
	ArchitectureARM Architecture = "arm"
	// ArchitectureARM64 describes the 64-bit ARM (v8) architecture.
	ArchitectureARM64 Architecture = "arm64"
)

var architectureRe = regexp.MustCompile("^[a-z0-9]*$")

// Validate returns an error if the platform is not a valid platform descriptor.
func (a Architecture) Validate() error {
	if !architectureRe.MatchString(string(a)) {
		return &InvalidArchitectureError{a}
	}
	return nil
}

// ResolveAuto resolves the value of ArchitectureAuto if needed based on the current runtime.GOARCH.
func (a Architecture) ResolveAuto() (Architecture, error) {
	if a != ArchitectureAuto {
		return a, nil
	}
	switch runtime.GOARCH {
	case "386":
		return Architecture386, nil
	case "amd64":
		return ArchitectureAMD64, nil
	case "arm":
		return ArchitectureARM, nil
	case "arm64":
		return ArchitectureARM64, nil
	default:
		return ArchitectureAuto, UnsupportedArchitectureError{
			Architecture(runtime.GOARCH),
		}
	}
}

// ArchitectureValues returns all supported values for Architecture excluding ArchitectureAuto.
func ArchitectureValues() []Architecture {
	return []Architecture{
		Architecture386,
		ArchitectureAMD64,
		ArchitectureARM,
		ArchitectureARM64,
	}
}
