// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentofu/tofudl/branding"
)

func main() {
	header := `// Copyright (c) ` + branding.SPDXAuthorsName + `
// SPDX-License-Identifier: ` + branding.SPDXLicense + `

`
	checkOnly := false
	flag.BoolVar(&checkOnly, "check-only", checkOnly, "Only check if the license headers are correct.")
	flag.Parse()

	var files []string
	if err := filepath.Walk(".", func(filePath string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(info.Name(), ".go") {
			files = append(files, filePath)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	hasError := false
	checkFailed := false
	for _, file := range files {
		fileContents, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read file %s (%v)", file, err)
			hasError = true
			continue
		}
		if strings.HasPrefix(string(fileContents), header) {
			continue
		}
		if checkOnly {
			log.Printf("%s does not have the correct license headers.", file)
			checkFailed = true
			continue
		}
		log.Printf("Updating license headers in %s...", file)
		tempFile := file + "~"
		if err := os.WriteFile(tempFile, []byte(header+string(fileContents)), 0644); err != nil { //nolint:gosec //The permissions are ok here.
			log.Printf("Failed to write file %s (%v)", tempFile, err)
			hasError = true
			continue
		}
		if err := os.Rename(tempFile, file); err != nil {
			log.Printf("Failed to move temporary file %s to %s (%v)", tempFile, file, err)
			hasError = true
		}
	}
	if hasError {
		log.Fatal("One or more files have failed processing.")
	}
	if checkFailed {
		log.Fatalf("One or more files don't contain the correct license headers, please run go generate.")
	}
}
