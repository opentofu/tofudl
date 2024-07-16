// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"crypto/tls"
	"net/http"

	"github.com/opentofu/tofudl/branding"
)

// Config describes the base configuration for the downloader.
type Config struct {
	// GPGKey holds the ASCII-armored GPG key to verify the binaries against. Defaults to the bundled
	// signing key.
	GPGKey string
	// APIURL describes the URL to the JSON API listing the versions and artifacts. Defaults to branding.DownloadAPIURL.
	APIURL string
	// APIURLAuthorization is an optional Authorization header to add to all request to the API URL. For requests
	// to the default API URL leave this empty.
	APIURLAuthorization string
	// DownloadMirrorAuthorization is an optional Authorization header to add to all requests to the download mirror.
	// Typically, you'll want to set this to "Bearer YOUR-GITHUB-TOKEN".
	DownloadMirrorAuthorization string
	// DownloadMirrorURLTemplate is a Go text template containing a URL with MirrorURLTemplateParameters embedded to
	// generate the download URL. Defaults to branding.DefaultMirrorURLTemplate.
	DownloadMirrorURLTemplate string
	// HTTPClient holds an HTTP client to use for requests. Defaults to the standard HTTP client with hardened TLS
	// settings.
	HTTPClient *http.Client
}

// ApplyDefaults applies defaults for all fields that are not set.
func (c *Config) ApplyDefaults() {
	if c.GPGKey == "" {
		c.GPGKey = branding.DefaultGPGKey
	}
	if c.APIURL == "" {
		c.APIURL = branding.DefaultDownloadAPIURL
	}
	if c.DownloadMirrorURLTemplate == "" {
		c.DownloadMirrorURLTemplate = branding.DefaultMirrorURLTemplate
	}
	if c.HTTPClient == nil {
		client := &http.Client{}
		client.Transport = http.DefaultTransport
		client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		c.HTTPClient = client
	}
}

// MirrorURLTemplateParameters describes the parameters to a URL template for mirrors.
type MirrorURLTemplateParameters struct {
	Version  Version
	Artifact string
}

// ConfigOpt is a function that modifies the config.
type ConfigOpt func(config *Config) error

// ConfigGPGKey is a config option to set an ASCII-armored GPG key.
func ConfigGPGKey(gpgKey string) ConfigOpt {
	return func(config *Config) error {
		if config.GPGKey != "" {
			return &InvalidConfigurationError{Message: "Duplicate options for GPG key."}
		}
		config.GPGKey = gpgKey
		return nil
	}
}

// ConfigAPIURL adds an API URL for the version listing. Defaults to branding.DownloadAPIURL.
func ConfigAPIURL(url string) ConfigOpt {
	return func(config *Config) error {
		if config.APIURL != "" {
			return &InvalidConfigurationError{Message: "Duplicate options for API URL."}
		}
		config.APIURL = url
		return nil
	}
}

// ConfigAPIAuthorization adds an authorization header to any request sent to the API server. This is not needed
// for the default API, but may be needed for private mirrors.
func ConfigAPIAuthorization(authorization string) ConfigOpt {
	return func(config *Config) error {
		if config.APIURLAuthorization != "" {
			return &InvalidConfigurationError{Message: "Duplicate options for API authorization.."}
		}
		config.APIURLAuthorization = authorization
		return nil
	}
}

// ConfigDownloadMirrorAuthorization adds the specified value to any request when connecting the download mirror.
// For example, you can add your GitHub token by specifying "Bearer YOUR-TOKEN-HERE".
func ConfigDownloadMirrorAuthorization(authorization string) ConfigOpt {
	return func(config *Config) error {
		if config.DownloadMirrorAuthorization != "" {
			return &InvalidConfigurationError{Message: "Duplicate options for download mirror authorization."}
		}
		config.DownloadMirrorAuthorization = authorization
		return nil
	}
}

// ConfigDownloadMirrorURLTemplate adds a Go text template containing a URL with MirrorURLTemplateParameters embedded to
// generate the download URL. Defaults to branding.DefaultMirrorURLTemplate.
func ConfigDownloadMirrorURLTemplate(urlTemplate string) ConfigOpt {
	return func(config *Config) error {
		if config.DownloadMirrorURLTemplate != "" {
			return &InvalidConfigurationError{Message: "Duplicate options for download mirror URL template."}
		}
		config.DownloadMirrorURLTemplate = urlTemplate
		return nil
	}
}

// ConfigHTTPClient adds a customized HTTP client to the downloader.
func ConfigHTTPClient(client *http.Client) ConfigOpt {
	return func(config *Config) error {
		if config.HTTPClient != nil {
			return &InvalidConfigurationError{Message: "Duplicate options for the HTTP client."}
		}
		config.HTTPClient = client
		return nil
	}
}
