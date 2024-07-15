// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"

	"github.com/opentofu/tofudl/cli"
)

func main() {
	c := cli.New()
	os.Exit(c.Run(os.Args, os.Environ(), os.Stdout, os.Stderr))
}
