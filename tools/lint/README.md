# golangci-lint wrapper

In order to make build tools runnable on any platform, this directory contains a thin wrapper that calls `golangci-lint run` without adding it to the package dependencies. You can run it by typing `go run github.com/opentofu/tofudl/tools/lint` in the root directory.