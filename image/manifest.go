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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func (m *manifest) unpack(w walker, dest string) (retErr error) {
	// error out if the dest directory is not empty
	s, err := ioutil.ReadDir(dest)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to open file") // err contains dest
	}
	if len(s) > 0 {
		return fmt.Errorf("%s is not empty", dest)
	}
	defer func() {
		// if we encounter error during unpacking
		// clean up the partially-unpacked destination
		if retErr != nil {
			if err := os.RemoveAll(dest); err != nil {
				fmt.Printf("Error: failed to remove partially-unpacked destination %v", err)
			}
		}
	}()
	for _, d := range m.Layers {
		if d.MediaType != string(schema.MediaTypeImageLayer) {
			continue
		}

		switch err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
			if info.IsDir() {
				return nil
			}

			dd, err := filepath.Rel(filepath.Join("blobs", d.algo()), filepath.Clean(path))
			if err != nil || d.hash() != dd {
				return nil
			}

			if err := unpackLayer(dest, r); err != nil {
				return errors.Wrap(err, "error extracting layer")
			}

			return errEOW
		}); err {
		case nil:
			return fmt.Errorf("%s: layer not found", dest)
		case errEOW:
		default:
			return err
		}
	}
	return nil
}

func unpackLayer(dest string, r io.Reader) error {
	entries := make(map[string]bool)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return errors.Wrap(err, "error creating gzip reader")
	}
	defer gz.Close()

	var dirs []*tar.Header
	tr := tar.NewReader(gz)

loop:
	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			break loop
		case nil:
			// success, continue below
		default:
			return errors.Wrapf(err, "error advancing tar stream")
		}

		var whiteout bool
		whiteout, err = unpackLayerEntry(dest, hdr, tr, &entries)
		if err != nil {
			return err
		}
		if whiteout {
			continue loop
		}

		// Directory mtimes must be handled at the end to avoid further
		// file creation in them to modify the directory mtime
		if hdr.Typeflag == tar.TypeDir {
			dirs = append(dirs, hdr)
		}
	}
	for _, hdr := range dirs {
		path := filepath.Join(dest, hdr.Name)

		finfo := hdr.FileInfo()
		// I believe the old version was using time.Now().UTC() to overcome an
		// invalid error from chtimes.....but here we lose hdr.AccessTime like this...
		if err := os.Chtimes(path, time.Now().UTC(), finfo.ModTime()); err != nil {
			return errors.Wrap(err, "error changing time")
		}
	}
	return nil
}

// unpackLayerEntry unpacks a single entry from a layer.
func unpackLayerEntry(dest string, header *tar.Header, reader io.Reader, entries *map[string]bool) (whiteout bool, err error) {
	header.Name = filepath.Clean(header.Name)
	if !strings.HasSuffix(header.Name, string(os.PathSeparator)) {
		// Not the root directory, ensure that the parent directory exists
		parent := filepath.Dir(header.Name)
		parentPath := filepath.Join(dest, parent)
		if _, err2 := os.Lstat(parentPath); err2 != nil && os.IsNotExist(err2) {
			if err3 := os.MkdirAll(parentPath, 0755); err3 != nil {
				return false, err3
			}
		}
	}
	path := filepath.Join(dest, header.Name)
	if (*entries)[path] {
		return false, fmt.Errorf("duplicate entry for %s", path)
	}
	(*entries)[path] = true
	rel, err := filepath.Rel(dest, path)
	if err != nil {
		return false, err
	}
	info := header.FileInfo()
	if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false, fmt.Errorf("%q is outside of %q", header.Name, dest)
	}

	if strings.HasPrefix(info.Name(), ".wh.") {
		path = strings.Replace(path, ".wh.", "", 1)

		if err = os.RemoveAll(path); err != nil {
			return true, errors.Wrap(err, "unable to delete whiteout path")
		}

		return true, nil
	}

	if header.Typeflag != tar.TypeDir {
		err = os.RemoveAll(path)
		if err != nil && !os.IsNotExist(err) {
			return false, err
		}
	}

	switch header.Typeflag {
	case tar.TypeDir:
		fi, err := os.Lstat(path)
		if err != nil && !os.IsNotExist(err) {
			return false, err
		}
		if os.IsNotExist(err) || !fi.IsDir() {
			err = os.RemoveAll(path)
			if err != nil && !os.IsNotExist(err) {
				return false, err
			}
			err = os.MkdirAll(path, info.Mode())
			if err != nil {
				return false, err
			}
		}

	case tar.TypeReg, tar.TypeRegA:
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, info.Mode())
		if err != nil {
			return false, errors.Wrap(err, "unable to open file")
		}

		if _, err := io.Copy(f, reader); err != nil {
			f.Close()
			return false, errors.Wrap(err, "unable to copy")
		}
		f.Close()

	case tar.TypeLink:
		target := filepath.Join(dest, header.Linkname)

		if !strings.HasPrefix(target, dest) {
			return false, fmt.Errorf("invalid hardlink %q -> %q", target, header.Linkname)
		}

		if err := os.Link(target, path); err != nil {
			return false, err
		}

	case tar.TypeSymlink:
		target := filepath.Join(filepath.Dir(path), header.Linkname)

		if !strings.HasPrefix(target, dest) {
			return false, fmt.Errorf("invalid symlink %q -> %q", path, header.Linkname)
		}

		if err := os.Symlink(header.Linkname, path); err != nil {
			return false, err
		}
	case tar.TypeXGlobalHeader:
		return false, nil
	}

	return false, nil
}
