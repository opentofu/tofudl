// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"regexp"
	"runtime"
)

// Platform describes the operating system to download OpenTofu for. Defaults to the current operating system.
type Platform string

const (
	// PlatformAuto is the default value and describes the current operating system.
	PlatformAuto Platform = ""
	// PlatformWindows describes the Windows platform.
	PlatformWindows Platform = "windows"
	// PlatformLinux describes the Linux platform.
	PlatformLinux Platform = "linux"
	// PlatformMacOS describes the macOS (Darwin) platform.
	PlatformMacOS Platform = "darwin"
	// PlatformSolaris describes the Solaris platform. (Note: this is currently only supported on AMD64.)
	PlatformSolaris Platform = "solaris"
	// PlatformOpenBSD describes the OpenBSD platform. (Note: this is currently only supported on 386 and AMD64.)
	PlatformOpenBSD Platform = "openbsd"
	// PlatformFreeBSD describes the FreeBSD platform. (Note: this is currently not supported on ARM64)
	PlatformFreeBSD Platform = "freebsd"
)

var platformRe = regexp.MustCompile("^[a-z]*$")

// Validate returns an error if the platform is not a valid platform descriptor.
func (p Platform) Validate() error {
	if !platformRe.MatchString(string(p)) {
		return &InvalidPlatformError{p}
	}
	return nil
}

// ResolveAuto resolves the value of PlatformAuto if needed based on the current runtime.GOOS.
func (p Platform) ResolveAuto() (Platform, error) {
	if p != PlatformAuto {
		return p, nil
	}
	switch runtime.GOOS {
	case "windows":
		return PlatformWindows, nil
	case "linux":
		return PlatformLinux, nil
	case "darwin":
		return PlatformMacOS, nil
	case "solaris":
		return PlatformSolaris, nil
	case "openbsd":
		return PlatformOpenBSD, nil
	case "freebsd":
		return PlatformFreeBSD, nil
	default:
		return PlatformAuto, UnsupportedPlatformError{
			Platform(runtime.GOOS),
		}
	}
}

// PlatformValues returns all supported values for Platform excluding PlatformAuto.
func PlatformValues() []Platform {
	return []Platform{
		PlatformWindows,
		PlatformLinux,
		PlatformMacOS,
		PlatformSolaris,
		PlatformOpenBSD,
		PlatformFreeBSD,
	}
}
