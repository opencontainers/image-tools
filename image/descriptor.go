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
	"github.com/pkg/errors"
)

type descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

func (d *descriptor) algo() string {
	pts := strings.SplitN(d.Digest, ":", 2)
	if len(pts) != 2 {
		return ""
	}
	return pts[0]
}

func (d *descriptor) hash() string {
	pts := strings.SplitN(d.Digest, ":", 2)
	if len(pts) != 2 {
		return ""
	}
	return pts[1]
}

func listReferences(w walker) (map[string]*descriptor, error) {
	refs := make(map[string]*descriptor)

	if err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
		if info.IsDir() || !strings.HasPrefix(path, "refs") {
			return nil
		}

		var d descriptor
		if err := json.NewDecoder(r).Decode(&d); err != nil {
			return err
		}
		refs[info.Name()] = &d

		return nil
	}); err != nil {
		return nil, err
	}
	return refs, nil
}

func findDescriptor(w walker, name string) (*descriptor, error) {
	var d descriptor
	dpath := filepath.Join("refs", name)

	switch err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
		if info.IsDir() || filepath.Clean(path) != dpath {
			return nil
		}

		if err := json.NewDecoder(r).Decode(&d); err != nil {
			return err
		}

		return errEOW
	}); err {
	case nil:
		return nil, fmt.Errorf("%s: descriptor not found", dpath)
	case errEOW:
		return &d, nil
	default:
		return nil, err
	}
}

func (d *descriptor) validate(get reader, mts []string) error {
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

	rc, err := get.Get(*d)
	if err != nil {
		return err
	}
	defer rc.Close()

	return d.validateContent(rc)
}

func (d *descriptor) validateContent(r io.Reader) error {
	parsed, err := digest.Parse(d.Digest)
	if err != nil {
		return err
	}

	verifier := parsed.Verifier()
	n, err := io.Copy(verifier, r)
	if err != nil {
		return errors.Wrap(err, "error generating hash")
	}

	if n != d.Size {
		return errors.New("size mismatch")
	}

	if !verifier.Verified() {
		return errors.New("digest mismatch")
	}

	return nil
}
