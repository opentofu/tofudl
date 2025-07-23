# TofuDL: OpenTofu downloader library for Go with minimal dependencies
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopentofu%2Ftofudl.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopentofu%2Ftofudl?ref=badge_shield)


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

## Caching

This library also supports caching using the mirror tool:

```go
package main

import (
    "context"
    "os"
    "os/exec"
    "runtime"
    "time"

    "github.com/opentofu/tofudl"
)

func main() {
    // Initialize the downloader:
    dl, err := tofudl.New()
    if err != nil {
        panic(err)
    }

    // Set up the caching layer:
    storage, err := tofudl.NewFilesystemStorage("/tmp")
    if err != nil {
        panic(err)
    }
    mirror, err := tofudl.NewMirror(
        tofudl.MirrorConfig{
            AllowStale: false,
            APICacheTimeout: time.Minute * 10,
            ArtifactCacheTimeout: time.Hour * 24,
        },
        storage,
        dl,
    )
    if err != nil {
        panic(err)
    }

    // Download the latest stable version
    // for the current architecture and platform:
    binary, err := mirror.Download(context.TODO())
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

You can also use the `mirror` variable as an `http.Handler`. Additionally, you can also call `PreWarm` on the caching layer in order to pre-warm your local caches. (Be careful, this may take a long time!)

## Standalone mirror

The example above showed a cache/mirror that acts as a pull-through cache to upstream. You can alternatively also use the mirror as a stand-alone mirror and publish your own binaries. The mirror has functions to facilitate uploading basic artifacts, but you can also use the `ReleaseBuilder` to make building releases easier. (Note: the `ReleaseBuilder` only builds artifacts needed for TofuDL, not all artifacts OpenTofu typically publishes.)

## Advanced usage

Both `New()` and `Download()` accept a number of options. You can find the detailed documentation [here](https://pkg.go.dev/github.com/opentofu/tofudl).


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopentofu%2Ftofudl.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopentofu%2Ftofudl?ref=badge_large)