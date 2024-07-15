// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

// Package cli is a demonstration how a CLI downloader can be implemented with this library. In order to not include
// additional dependencies, it implements CLI argument parsing and environment variable handling.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/branding"
)

// New creates a new CLI interface to run the downloader on. For API usage please refer to tofudl.New.
func New() CLI {
	return &cli{
		configOptions: []option{
			optionAPIURL,
			optionDownloadMirrorURLTemplate,
			optionGPGKeyFile,
			optionAPIAuthorization,
			optionDownloadMirrorAuthorization,
			optionPlatform,
			optionArchitecture,
			optionVersion,
			optionStability,
			optionTimeout,
			optionOutput,
		},
		outputFileWriter: os.WriteFile,
	}
}

// CLI is a command-line downloader. This is for CLI use only. For API usage please refer to tofudl.Downloader.
type CLI interface {
	// Run executes the downloader non-interactively with the given options and returns the exit code.
	Run(argv []string, env []string, stdout io.Writer, stderr io.Writer) int
}

const isWindows = runtime.GOOS == "windows"

type cli struct {
	configOptions    []option
	outputFileWriter func(fileName string, bytes []byte, mode os.FileMode) error
}

func (c cli) Run(
	argv []string,
	env []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	for _, arg := range argv {
		if arg == "-h" || arg == "--help" {
			c.Usage(stdout)
			return 0
		}
	}

	args, err := argvToMap(argv[1:])
	if err != nil {
		_, _ = stderr.Write([]byte(fmt.Sprintf("Failed to parse command line arguments: %s", err.Error())))
		c.Usage(stdout)
		return 1
	}

	envVars, err := envToMap(env)
	if err != nil {
		_, _ = stderr.Write([]byte(fmt.Sprintf("Failed to parse environment variables: %s", err.Error())))
		c.Usage(stdout)
		return 1
	}

	var configOpts []tofudl.ConfigOpt
	var downloadOpts []tofudl.DownloadOpt
	storedConfigs := map[string]string{}
	for _, cliOpt := range c.configOptions {
		value := ""
		optName := ""
		if cliOpt.envVarName != "" {
			if envValue, ok := envVars[cliOpt.envVarName]; ok {
				value = envValue
				optName = "environment variable " + cliOpt.envVarName
			}
		}
		if cliOpt.cliFlagName != "" {
			if cliValue, ok := args[cliOpt.cliFlagName]; ok {
				value = cliValue
				delete(args, cliOpt.cliFlagName)
				optName = "command line option " + cliOpt.cliFlagName
			}
		}
		if value == "" {
			value = cliOpt.defaultValue
			var parts []string
			if cliOpt.cliFlagName != "" {
				parts = append(parts, "--"+cliOpt.cliFlagName)
			}
			if cliOpt.envVarName != "" {
				parts = append(parts, cliOpt.envVarName)
			}
			if len(parts) == 0 {
				parts = append(parts, "unnamed option")
			}
			optName = "default value for " + strings.Join(parts, "/")
		}
		if value != "" { //nolint:nestif // Slightly complex, but easier to keep in one function.
			if cliOpt.validate != nil {
				if err := cliOpt.validate(value); err != nil {
					_, _ = stderr.Write([]byte(fmt.Sprintf("Failed to parse %s (%v)", optName, err)))
					c.Usage(stdout)
					return 1
				}
			}

			if cliOpt.applyConfig != nil {
				opt, err := cliOpt.applyConfig(value)
				if err != nil {
					_, _ = stderr.Write([]byte(fmt.Sprintf("Failed to parse %s (%v)", optName, err)))
					c.Usage(stdout)
					return 1
				}
				configOpts = append(configOpts, opt)
			}

			if cliOpt.applyDownloadOption != nil {
				opt, err := cliOpt.applyDownloadOption(value)
				if err != nil {
					_, _ = stderr.Write([]byte(fmt.Sprintf("Failed to parse %s (%v)", optName, err)))
					c.Usage(stdout)
					return 1
				}
				downloadOpts = append(downloadOpts, opt)
			}

			if cliOpt.cliFlagName != "" {
				storedConfigs[cliOpt.cliFlagName] = value
			}
		}
	}

	if len(args) != 0 {
		for arg := range args {
			_, _ = stderr.Write([]byte(fmt.Sprintf("Invalid command line option: %s", arg)))
		}
		c.Usage(stdout)
		return 1
	}

	dl, err := tofudl.New(configOpts...)
	if err != nil {
		_, _ = stderr.Write([]byte(err.Error()))
		c.Usage(stdout)
		return 1
	}

	timeout, _ := strconv.Atoi(storedConfigs[optionTimeout.cliFlagName])
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	binaryContents, err := dl.Download(ctx, downloadOpts...)
	if err != nil {
		_, _ = stderr.Write([]byte(err.Error()))
		return 1
	}
	if err := c.outputFileWriter(storedConfigs[optionOutput.cliFlagName], binaryContents, 0755); err != nil {
		_, _ = stderr.Write([]byte(
			fmt.Sprintf("Failed to write output file: %s", storedConfigs[optionOutput.cliFlagName]),
		))
		return 1
	}
	return 0
}

func (c cli) Usage(stdout io.Writer) {
	binaryName := branding.CLIBinaryName
	if isWindows {
		binaryName += ".exe"
	}

	_, _ = stdout.Write([]byte("Usage: " + binaryName + " [OPTIONS]\n"))
	_, _ = stdout.Write([]byte("\nOPTIONS:\n\n"))

	for _, opt := range c.configOptions {
		var parts []string
		if opt.cliFlagName != "" {
			parts = append(parts, "--"+opt.cliFlagName)
		}
		if opt.envVarName != "" {
			if isWindows {
				parts = append(parts, "$Env:"+opt.envVarName)
			} else {
				parts = append(parts, "$"+opt.envVarName)
			}
		}
		firstLine := strings.Join(parts, " / ")
		if opt.defaultValue != "" {
			firstLine += " (Default: " + opt.defaultValue + ")"
		} else if opt.defaultDescription != "" {
			firstLine += " (Default: " + opt.defaultDescription + ")"
		}
		_, _ = stdout.Write([]byte(firstLine))
		_, _ = stdout.Write([]byte("\n\n"))
		_, _ = stdout.Write([]byte("  " + opt.description))
		_, _ = stdout.Write([]byte("\n\n"))
	}
}

func argvToMap(argv []string) (map[string]string, error) {
	result := map[string]string{}
	for {
		if len(argv) == 0 {
			return result, nil
		}
		if len(argv) == 1 {
			return result, fmt.Errorf("unexpected argument or value missing: %s", argv[0])
		}
		if !strings.HasPrefix(argv[0], "--") {
			return nil, fmt.Errorf("unexpected argument: %s", argv[0])
		}
		result[argv[0][2:]] = argv[1]
		argv = argv[2:]
	}
}
func envToMap(env []string) (map[string]string, error) {
	result := map[string]string{}
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid environment variable: %s", e)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}
