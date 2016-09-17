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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
)

// CheckDirVersion checks the oci-layout entry in an image-layout
// directory and returns an error if oci-layout is missing or has
// unrecognized content.
func CheckDirVersion(ctx context.Context, path string) (err error) {
	file, err := os.Open(filepath.Join(path, "oci-layout"))
	if os.IsNotExist(err) {
		return errors.New("oci-layout not found")
	}
	if err != nil {
		return err
	}
	defer file.Close()

	return CheckVersion(ctx, file)
}

// CreateDir creates a new image-layout directory at the given path.
func CreateDir(ctx context.Context, path string) (err error) {
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(
		filepath.Join(path, "oci-layout"),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0666,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	imageLayoutVersion := ImageLayoutVersion{
		Version: "1.0.0",
	}
	imageLayoutVersionBytes, err := json.Marshal(imageLayoutVersion)
	if err != nil {
		return err
	}
	n, err := file.Write(imageLayoutVersionBytes)
	if err != nil {
		return err
	}
	if n < len(imageLayoutVersionBytes) {
		return fmt.Errorf("wrote %d of %d bytes", n, len(imageLayoutVersionBytes))
	}

	err = os.MkdirAll(filepath.Join(path, "blobs"), 0777)
	if err != nil {
		return err
	}

	return os.MkdirAll(filepath.Join(path, "refs"), 0777)
}

// DirDelete removes an entry from a directory, wrapping a
// path-constructing call to entryPath and a removing call to
// os.Remove.  Deletion is idempotent (unlike os.Remove, where
// attempting to delete a nonexistent path results in an error).
func DirDelete(ctx context.Context, path string, entry string, entryPath EntryPath) (err error) {
	targetName, err := entryPath(entry, string(os.PathSeparator))
	if err != nil {
		return err
	}

	err = os.Remove(filepath.Join(path, targetName))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
