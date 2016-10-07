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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/image-spec/schema"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type manifest struct {
	Config descriptor   `json:"config"`
	Layers []descriptor `json:"layers"`
}

func findManifest(w walker, d *descriptor) (*manifest, error) {
	var m manifest
	mpath := filepath.Join("blobs", d.algo(), d.hash())

	switch err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
		if info.IsDir() || filepath.Clean(path) != mpath {
			return nil
		}

		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrapf(err, "%s: error reading manifest", path)
		}

		if err := schema.MediaTypeManifest.Validate(bytes.NewReader(buf)); err != nil {
			return errors.Wrapf(err, "%s: manifest validation failed", path)
		}

		if err := json.Unmarshal(buf, &m); err != nil {
			return err
		}

		if len(m.Layers) == 0 {
			return fmt.Errorf("%s: no layers found", path)
		}

		return errEOW
	}); err {
	case nil:
		return nil, fmt.Errorf("%s: manifest not found", mpath)
	case errEOW:
		return &m, nil
	default:
		return nil, err
	}
}

func (m *manifest) validate(w walker) error {
	if err := m.Config.validate(w, []string{v1.MediaTypeImageConfig}); err != nil {
		return errors.Wrap(err, "config validation failed")
	}

	for _, d := range m.Layers {
		if err := d.validate(w, []string{v1.MediaTypeImageLayer}); err != nil {
			return errors.Wrap(err, "layer validation failed")
		}
	}

	return nil
}
