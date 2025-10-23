// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0

package tofudl

import (
	"fmt"
	"regexp"
	"strings"
)

type NightlyID string

var nightlyIDRegex = regexp.MustCompile(`^\d{8}-[a-fA-F0-9]{10}$`)

func newNightlyID(buildDate, hash string) (NightlyID, error) {
	nightlyID := NightlyID(fmt.Sprintf("%s-%s", buildDate, hash))
	if err := nightlyID.Validate(); err != nil {
		return "", err
	}
	return nightlyID, nil
}

func (id NightlyID) Validate() error {
	if !nightlyIDRegex.MatchString(string(id)) {
		return &InvalidOptionsError{
			fmt.Errorf("nightly build id %q does not match required format YYYYMMDD-XXXXXXXXXX", id),
		}
	}
	return nil
}

// GetDate returns date part of the nightly ID
// It is assumed that the id is in correct format and validate is already called
func (id NightlyID) GetDate() string {
	splits := strings.Split(string(id), "-")
	return splits[0]
}
