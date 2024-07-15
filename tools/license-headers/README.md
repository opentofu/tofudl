# License headers checker

This tool checks all `.go` files to contain the right copyright headers and adds them if needed. You can run it to update the headers by running `go generate` or `go run github.com/opentofu/tofudl/tools/license-headers` in the root directory. To run the tool in check-only mode for CI, use the `-check` option.
