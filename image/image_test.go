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
	"archive/tar"
	"io/ioutil"
	"os"
	"testing"
)

func TestValidateLayout(t *testing.T) {
	root, err := ioutil.TempDir("", "oci-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	il := imageLayout{
		rootDir:  root,
		layout:   layoutStr,
		ref:      refTag,
		manifest: manifestStr,
		config:   configStr,
		tarList: []tarContent{
			tarContent{&tar.Header{Name: "test", Size: 4, Mode: 0600}, []byte("test")},
		},
	}

	// create image layout bundle
	err = createImageLayoutBundle(il)
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateLayout(root, []string{refTag}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
