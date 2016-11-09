// Copyright 2016 The Linux Foundation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package layout defines utility code shared by refs/layout and cas/layout.
package layout

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"
)

// EntryPath is a template for helpers that convert from ref or blob
// names to image-layout paths.
type EntryPath func(entry string, separator string) (path string, err error)

// ImageLayoutVersion represents the oci-version content for the image
// layout format.
type ImageLayoutVersion struct {
	Version string `json:"imageLayoutVersion"`
}

// CheckVersion checks an oci-layout reader and returns an error if it
// has unrecognized content.
func CheckVersion(ctx context.Context, reader io.Reader) (err error) {
	decoder := json.NewDecoder(reader)
	var version ImageLayoutVersion
	err = decoder.Decode(&version)
	if err != nil {
		return err
	}
	if version.Version != "1.0.0" {
		return fmt.Errorf("unrecognized imageLayoutVersion: %q", version.Version)
	}

	return nil
}
