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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type descriptor v1.Descriptor

func (d *descriptor) algo() string {
	pts := strings.SplitN(string(d.Digest), ":", 2)
	if len(pts) != 2 {
		return ""
	}
	return pts[0]
}

func (d *descriptor) hash() string {
	pts := strings.SplitN(string(d.Digest), ":", 2)
	if len(pts) != 2 {
		return ""
	}
	return pts[1]
}

func listReferences(w walker) (map[string]*descriptor, error) {
	refs := make(map[string]*descriptor)
	var index v1.ImageIndex

	if err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
		if info.IsDir() || filepath.Clean(path) != "index.json" {
			return nil
		}

		if err := json.NewDecoder(r).Decode(&index); err != nil {
			return err
		}

		for i := 0; i < len(index.Manifests); i++ {
			if index.Manifests[i].Descriptor.Annotations["org.opencontainers.ref.name"] != "" {
				refs[index.Manifests[i].Descriptor.Annotations["org.opencontainers.ref.name"]] = (*descriptor)(&index.Manifests[i].Descriptor)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return refs, nil
}

func findDescriptor(w walker, name string) (*descriptor, error) {
	var d descriptor
	var index v1.ImageIndex

	switch err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
		if info.IsDir() || filepath.Clean(path) != "index.json" {
			return nil
		}

		if err := json.NewDecoder(r).Decode(&index); err != nil {
			return err
		}

		for i := 0; i < len(index.Manifests); i++ {
			if index.Manifests[i].Descriptor.Annotations["org.opencontainers.ref.name"] == name {
				d = (descriptor)(index.Manifests[i].Descriptor)
				return errEOW
			}
		}

		return nil
	}); err {
	case nil:
		return nil, fmt.Errorf("index.json: descriptor not found")
	case errEOW:
		return &d, nil
	default:
		return nil, err
	}
}

func (d *descriptor) validate(w walker, mts []string) error {
	var found bool
	for _, mt := range mts {
		if d.MediaType == mt {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid descriptor MediaType %q", d.MediaType)
	}

	parsed, err := digest.Parse(string(d.Digest))
	if err != nil {
		return err
	}

	// Copy the contents of the layer in to the verifier
	verifier := parsed.Verifier()
	numBytes, err := w.get(*d, verifier)
	if err != nil {
		return err
	}

	if err != nil {
		return errors.Wrap(err, "error generating hash")
	}

	if numBytes != d.Size {
		return errors.New("size mismatch")
	}

	if !verifier.Verified() {
		return errors.New("digest mismatch")
	}

	return nil
}
