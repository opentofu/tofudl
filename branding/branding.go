// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package branding

// ProductName describes the name of the product being downloaded.
const ProductName = "OpenTofu"

// DefaultDownloadAPIURL describes the API serving the version and file information.
const DefaultDownloadAPIURL = "https://get.opentofu.org/tofu/api.json"

// DefaultMirrorURLTemplate is a Go template that describes the download URL with the {{ .Version }} and {{ .Artifact }}
// embedded into the URL.
const DefaultMirrorURLTemplate = "https://github.com/opentofu/opentofu/releases/download/v{{ .Version }}/{{ .Artifact }}"

// BinaryName holds the name of the binary in the artifact. This may be suffixed .exe on Windows.
const BinaryName = "tofu"

// GPGKeyURL describes the URL to download the bundled GPG key from. The GPG key bundler uses this to download the
// GPG key for verification.
const GPGKeyURL = "https://get.opentofu.org/opentofu.asc"

// GPGKeyFingerprint is the GPG key fingerprint the bundler should expect to find when downloading the key.
const GPGKeyFingerprint = "E3E6E43D84CB852EADB0051D0C0AF313E5FD9F80"

// SPDXAuthorsName describes the name of the authors to be attributed in copyright notices in this repository.
const SPDXAuthorsName = "The OpenTofu Authors"

// SPDXLicense describes the license for copyright attribution in this repository.
const SPDXLicense = "MPL-2.0"

// MaximumUncompressedFileSize indicates the maximum file size when uncompressed.
const MaximumUncompressedFileSize = 1024 * 1024 * 1024 * 1024
