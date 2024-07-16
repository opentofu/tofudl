// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"fmt"

	"github.com/opentofu/tofudl/branding"
)

// InvalidPlatformError describes an error where a platform name was found to be invalid.
type InvalidPlatformError struct {
	Platform Platform
}

// Error returns the error message.
func (e InvalidPlatformError) Error() string {
	return fmt.Sprintf("Invalid platform: %s", e.Platform)
}

// UnsupportedPlatformError indicates that the given runtime.GOOS platform is not supported and cannot automatically
// resolve to a build artifact.
type UnsupportedPlatformError struct {
	Platform Platform
}

// Error returns the error message.
func (e UnsupportedPlatformError) Error() string {
	return fmt.Sprintf("Unsupported platform: %s", e.Platform)
}

// UnsupportedArchitectureError indicates that the given runtime.GOARCH architecture is not supported and cannot
// automatically resolve to a build artifact.
type UnsupportedArchitectureError struct {
	Architecture Architecture
}

// Error returns the error message.
func (e UnsupportedArchitectureError) Error() string {
	return fmt.Sprintf("Unsupported architecture: %s", e.Architecture)
}

// InvalidArchitectureError describes an error where an architecture name was found to be invalid.
type InvalidArchitectureError struct {
	Architecture Architecture
}

// Error returns the error message.
func (e InvalidArchitectureError) Error() string {
	return fmt.Sprintf("Invalid architecture: %s", e.Architecture)
}

// InvalidVersionError describes an error where the version string is invalid.
type InvalidVersionError struct {
	Version Version
}

// Error returns the error message.
func (e InvalidVersionError) Error() string {
	return fmt.Sprintf("Invalid version: %s", e.Version)
}

// NoSuchVersionError indicates that the given version does not exist on the API endpoint.
type NoSuchVersionError struct {
	Version Version
}

// Error returns the error message.
func (e NoSuchVersionError) Error() string {
	return fmt.Sprintf("No such version: %s", e.Version)
}

// UnsupportedPlatformOrArchitectureError describes an error where the platform name and architecture are syntactically
// valid, but no release artifact was found matching that name.
type UnsupportedPlatformOrArchitectureError struct {
	Platform     Platform
	Architecture Architecture
	Version      Version
}

func (e UnsupportedPlatformOrArchitectureError) Error() string {
	return fmt.Sprintf(
		"Unsupported platform (%s) or architecture (%s) for %s version %s.",
		e.Platform,
		e.Architecture,
		branding.ProductName,
		e.Version,
	)
}

// InvalidConfigurationError indicates that the base configuration for the downloader is invalid.
type InvalidConfigurationError struct {
	Message string
	Cause   error
}

// Error returns the error message.
func (e InvalidConfigurationError) Error() string {
	if e.Cause != nil {
		return "Invalid configuration: " + e.Message + " (" + e.Cause.Error() + ")"
	}
	return "Invalid configuration: " + e.Message
}

func (e InvalidConfigurationError) Unwrap() error {
	return e.Cause
}

// SignatureError indicates that the signature verification failed.
type SignatureError struct {
	Message string
	Cause   error
}

// Error returns the error message.
func (e SignatureError) Error() string {
	if e.Cause != nil {
		return "Invalid signature: " + e.Message + " (" + e.Cause.Error() + ")"
	}
	return "Invalid signature: " + e.Message
}

func (e SignatureError) Unwrap() error {
	return e.Cause
}

// InvalidOptionsError indicates that the request options are invalid.
type InvalidOptionsError struct {
	Cause error
}

// Error returns the error message.
func (e InvalidOptionsError) Error() string {
	return "Invalid options: " + e.Cause.Error()
}

// Unwrap returns the original error.
func (e InvalidOptionsError) Unwrap() error {
	return e.Cause
}

// NoSuchArtifactError indicates that there is no artifact for the given version with the given name.
type NoSuchArtifactError struct {
	ArtifactName string
}

// Error returns the error message.
func (e NoSuchArtifactError) Error() string {
	return "No such artifact: " + e.ArtifactName
}

// RequestFailedError indicates that a request to an API or the download mirror failed.
type RequestFailedError struct {
	Cause error
}

// Error returns the error message.
func (e RequestFailedError) Error() string {
	return fmt.Sprintf("Request failed (%v)", e.Cause)
}

// Unwrap returns the original error.
func (e RequestFailedError) Unwrap() error {
	return e.Cause
}

// ArtifactCorruptedError indicates that the downloaded artifact is corrupt.
type ArtifactCorruptedError struct {
	Artifact string
	Cause    error
}

// Error returns the error message.
func (e ArtifactCorruptedError) Error() string {
	return fmt.Sprintf("Corrupted artifact %s (%v)", e.Artifact, e.Cause)
}

// Unwrap returns the original error.
func (e ArtifactCorruptedError) Unwrap() error {
	return e.Cause
}
