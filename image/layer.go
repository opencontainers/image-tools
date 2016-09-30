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

package image

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateLayer(child, parent, dest string) error {
	arch, err := Diff(child, parent)
	if err != nil {
		return err
	}
	defer arch.Close()
	filename := fmt.Sprintf("%s.tar", filepath.Clean(child))
	if dest != "" {
		filename = filepath.Clean(dest)
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, arch)
	return err
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be "".
func Diff(child, parent string) (arch io.ReadCloser, err error) {
	changes, err := ChangesDirs(child, parent)
	if err != nil {
		return nil, err
	}
	archive, err := exportChanges(child, changes)
	if err != nil {
		return nil, err
	}
	return archive, nil
}
