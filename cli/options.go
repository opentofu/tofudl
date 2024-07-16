// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/branding"
)

type option struct {
	cliFlagName        string
	envVarName         string
	description        string
	defaultValue       string
	defaultDescription string

	validate            func(value string) error
	applyConfig         func(value string) (tofudl.ConfigOpt, error)
	applyDownloadOption func(value string) (tofudl.DownloadOpt, error)
}

var optionAPIURL = option{
	cliFlagName:  "api-url",
	envVarName:   branding.CLIEnvPrefix + "API_URL",
	description:  "URL to fetch the version information from.",
	defaultValue: branding.DefaultDownloadAPIURL,
	applyConfig: func(value string) (tofudl.ConfigOpt, error) {
		return tofudl.ConfigAPIURL(value), nil
	},
}

var optionDownloadMirrorURLTemplate = option{
	cliFlagName:  "download-mirror-url-template",
	envVarName:   branding.CLIEnvPrefix + "DOWNLOAD_MIRROR_URL_TEMPLATE",
	description:  "URL template for the artifact mirror. May contain {{ .Version }} for the " + branding.ProductName + " version and {{ .Artifact }} for the artifact name.",
	defaultValue: branding.DefaultMirrorURLTemplate,
	applyConfig: func(value string) (tofudl.ConfigOpt, error) {
		return tofudl.ConfigDownloadMirrorURLTemplate(value), nil
	},
}

var optionGPGKeyFile = option{
	cliFlagName:        "gpg-key-file",
	envVarName:         branding.CLIEnvPrefix + "GPG_KEY_FILE",
	description:        "GPG key file to verify downloaded artifacts against.",
	defaultDescription: fmt.Sprintf("bundled key, fingerprint %s", branding.GPGKeyFingerprint),
	applyConfig: func(value string) (tofudl.ConfigOpt, error) {
		gpgKey, err := os.ReadFile(value)
		if err != nil {
			return nil, fmt.Errorf("failed to read GPG key file %s (%w)", value, err)
		}
		return tofudl.ConfigGPGKey(string(gpgKey)), nil
	},
}

var optionAPIAuthorization = option{
	cliFlagName: "api-authorization",
	envVarName:  branding.CLIEnvPrefix + "API_AUTHORIZATION",
	description: "Use the provided value as an 'Authorization' header when requesting data from the API server. This is not needed for the default settings, but may be needed for private mirrors.",
	applyConfig: func(value string) (tofudl.ConfigOpt, error) {
		return tofudl.ConfigAPIAuthorization(value), nil
	},
}

var optionDownloadMirrorAuthorization = option{
	cliFlagName: "download-mirror-authorization",
	envVarName:  branding.CLIEnvPrefix + "DOWNLOAD_MIRROR_AUTHORIZATION",
	description: "Use the provided value as an 'Authorization' header when requesting data from the downloar mirror. You can set your GitHub token by specifying 'Bearer GITHUB_TOKEN' here to work around rate limits.",
	applyConfig: func(value string) (tofudl.ConfigOpt, error) {
		return tofudl.ConfigDownloadMirrorAuthorization(value), nil
	},
}

var optionPlatform = option{
	cliFlagName:        "platform",
	envVarName:         branding.CLIEnvPrefix + "PLATFORM",
	description:        "Platform to download the binary for. Possible values are: " + getPlatformValues() + ", or a custom value.",
	defaultDescription: "current platform",
	applyDownloadOption: func(value string) (tofudl.DownloadOpt, error) {
		platform := tofudl.Platform(value)
		if err := platform.Validate(); err != nil {
			return nil, err
		}
		return tofudl.DownloadOptPlatform(platform), nil
	},
}

func getPlatformValues() string {
	values := tofudl.PlatformValues()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = string(value)
	}
	return strings.Join(result, ", ")
}

var optionArchitecture = option{
	cliFlagName:        "architecture",
	envVarName:         branding.CLIEnvPrefix + "ARCHITECTURE",
	description:        "Architecture to download the binary for. Possible values are: " + getArchitectureValues() + ", or a custom value.",
	defaultDescription: "current platform",
	applyDownloadOption: func(value string) (tofudl.DownloadOpt, error) {
		architecture := tofudl.Architecture(value)
		if err := architecture.Validate(); err != nil {
			return nil, err
		}
		return tofudl.DownloadOptArchitecture(architecture), nil
	},
}

func getArchitectureValues() string {
	values := tofudl.ArchitectureValues()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = string(value)
	}
	return strings.Join(result, ", ")
}

var optionVersion = option{
	cliFlagName:        "version",
	envVarName:         branding.CLIEnvPrefix + "VERSION",
	description:        "Exact version to download.",
	defaultDescription: "latest version matching the minimum stability",
	applyDownloadOption: func(value string) (tofudl.DownloadOpt, error) {
		version := tofudl.Version(value)
		if err := version.Validate(); err != nil {
			return nil, err
		}
		return tofudl.DownloadOptVersion(version), nil
	},
}

var optionStability = option{
	cliFlagName:        "minimum-stability",
	envVarName:         branding.CLIEnvPrefix + "MINIMUM_STABILITY",
	description:        "Minimum stability to download for. Possible values are: " + getStabilityValues() + "",
	defaultDescription: "stable",
	applyDownloadOption: func(value string) (tofudl.DownloadOpt, error) {
		platform := tofudl.Platform(value)
		if err := platform.Validate(); err != nil {
			return nil, err
		}
		return tofudl.DownloadOptPlatform(platform), nil
	},
}

func getStabilityValues() string {
	values := tofudl.StabilityValues()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = string(value)
	}
	return strings.Join(result, ", ")
}

var optionTimeout = option{
	cliFlagName:  "timeout",
	envVarName:   branding.CLIEnvPrefix + "TIMEOUT",
	description:  "Download timeout in seconds.",
	defaultValue: "300",
	validate: func(value string) error {
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer: %s", value)
		}
		if v < 1 {
			return fmt.Errorf("timeout must be positive: %s", value)
		}
		return nil
	},
}

var optionOutput = option{
	cliFlagName:  "output",
	envVarName:   branding.CLIEnvPrefix + "OUTPUT",
	description:  "Write the " + branding.BinaryName + " to this file.",
	defaultValue: getDefaultFile(),
}

func getDefaultFile() string {
	defaultFile := branding.BinaryName
	if isWindows {
		defaultFile += ".exe"
	}
	return defaultFile
}
