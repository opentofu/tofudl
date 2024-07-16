# TofuDL: OpenTofu downloader library for Go with minimal dependencies

This library provides an easy way to download, verify, and unpack OpenTofu binaries for local use in Go. It has a minimal set of dependencies and is easy to integrate.

**Note:** This is not a standalone tool to download OpenTofu, it is purely meant to be used as Go library in support of other tools that need to run `tofu`. Please check the [installation instructions](https://opentofu.org/docs/intro/install/) for supported ways to perform an OpenTofu installation.

## Basic usage

The downloader will work without any extra configuration out of the box:

```go
package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/opentofu/tofudl"
)

func main() {
	// Initialize the downloader:
	dl, err := tofudl.New()
	if err != nil {
		panic(err)
	}

	// Download the latest stable version
	// for the current architecture and platform:
	binary, err := dl.Download(context.TODO())
	if err != nil {
		panic(err)
	}

	// Write out the tofu binary to the disk:
	file := "tofu"
	if runtime.GOOS == "windows" {
		file += ".exe"
	}
	if err := os.WriteFile(file, binary, 0755); err != nil {
		panic(err)
	}

	// Run tofu:
	cmd := exec.Command("./"+file, "init")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
```

## Advanced usage

Both `New()` and `Download()` accept a number of options. You can find the detailed documentation [here](https://pkg.go.dev/github.com/opentofu/tofudl).
