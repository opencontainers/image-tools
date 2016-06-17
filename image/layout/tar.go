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

package layout

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"
)

// CheckTarVersion walks a tarball pointed to by reader and returns an
// error if oci-layout is missing or has unrecognized content.
func CheckTarVersion(ctx context.Context, reader io.ReadSeeker) (err error) {
	_, err = reader.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(reader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			return errors.New("oci-layout not found")
		}
		if err != nil {
			return err
		}

		if header.Name == "./oci-layout" {
			decoder := json.NewDecoder(tarReader)
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
	}
}