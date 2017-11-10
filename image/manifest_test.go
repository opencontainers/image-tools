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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	bz2 "github.com/dsnet/compress/bzip2"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestUnpackLayerDuplicateEntries(t *testing.T) {
	tmp1, err := ioutil.TempDir("", "test-dup")
	if err != nil {
		t.Fatal(err)
	}
	tarfile := filepath.Join(tmp1, "test.tar")
	f, err := os.Create(tarfile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.RemoveAll(tmp1)
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	tw.WriteHeader(&tar.Header{Name: "test", Size: 4, Mode: 0600})
	io.Copy(tw, bytes.NewReader([]byte("test")))
	tw.WriteHeader(&tar.Header{Name: "test", Size: 5, Mode: 0600})
	io.Copy(tw, bytes.NewReader([]byte("test1")))
	tw.Close()
	gw.Close()

	r, err := os.Open(tarfile)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	tmp2, err := ioutil.TempDir("", "test-dest-unpack")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp2)
	if err := unpackLayer("application/vnd.oci.image.layer.v1.tar+gzip", f.Name(), tmp2, r); err != nil && !strings.Contains(err.Error(), "duplicate entry for") {
		t.Fatalf("Expected to fail with duplicate entry, got %v", err)
	}
}

func testUnpackLayer(t *testing.T, compression string, invalid bool) {
	tmp1, err := ioutil.TempDir("", "test-layer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp1)
	err = os.MkdirAll(filepath.Join(tmp1, "blobs", "sha256"), 0700)
	if err != nil {
		t.Fatal(err)
	}
	tarfile := filepath.Join(tmp1, "blobs", "sha256", "test.tar")
	f, err := os.Create(tarfile)
	if err != nil {
		t.Fatal(err)
	}

	var writer io.WriteCloser = f

	if !invalid {
		switch compression {
		case "gzip":
			writer = gzip.NewWriter(f)
		case "bzip2":
			writer, err = bz2.NewWriter(f, nil)
			if err != nil {
				t.Fatal(errors.Wrap(err, "compiling bzip compressor"))
			}
		}
	} else if invalid && compression == "" {
		writer = gzip.NewWriter(f)
	}

	tw := tar.NewWriter(writer)

	if headerErr := tw.WriteHeader(&tar.Header{Name: "test", Size: 4, Mode: 0600}); headerErr != nil {
		t.Fatal(headerErr)
	}

	if _, copyErr := io.Copy(tw, bytes.NewReader([]byte("test"))); copyErr != nil {
		t.Fatal(copyErr)
	}

	tw.Close()
	writer.Close()

	digester := digest.SHA256.Digester()
	file, err := os.Open(tarfile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	_, err = io.Copy(digester.Hash(), file)
	if err != nil {
		t.Fatal(err)
	}

	blobPath := filepath.Join(tmp1, "blobs", "sha256", digester.Digest().Hex())

	if renameErr := os.Rename(tarfile, blobPath); renameErr != nil {
		t.Fatal(errors.Wrap(renameErr, blobPath))
	}

	mediatype := "application/vnd.oci.image.layer.v1.tar"
	if compression != "" {
		mediatype += "+" + compression
	}

	testManifest := manifest{
		Layers: []v1.Descriptor{
			{
				MediaType: mediatype,
				Digest:    digester.Digest(),
			},
		},
	}
	err = testManifest.unpack(newPathWalker(tmp1), filepath.Join(tmp1, "rootfs"))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "%q / %s", blobPath, compression))
	}

	_, err = os.Stat(filepath.Join(tmp1, "rootfs", "test"))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "%q / %s", blobPath, compression))
	}
}

func TestUnpackLayer(t *testing.T) {
	testUnpackLayer(t, "gzip", true)
	testUnpackLayer(t, "gzip", false)
	testUnpackLayer(t, "", true)
	testUnpackLayer(t, "", false)
	testUnpackLayer(t, "bzip2", true)
	testUnpackLayer(t, "bzip2", false)
}

func TestUnpackLayerRemovePartiallyUnpackedFile(t *testing.T) {
	// generate a tar file has duplicate entry which will failed on unpacking
	tmp1, err := ioutil.TempDir("", "test-layer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp1)
	err = os.MkdirAll(filepath.Join(tmp1, "blobs", "sha256"), 0700)
	if err != nil {
		t.Fatal(err)
	}
	tarfile := filepath.Join(tmp1, "blobs", "sha256", "test.tar")
	f, err := os.Create(tarfile)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	tw.WriteHeader(&tar.Header{Name: "test", Size: 4, Mode: 0600})
	io.Copy(tw, bytes.NewReader([]byte("test")))
	tw.WriteHeader(&tar.Header{Name: "test", Size: 5, Mode: 0600})
	io.Copy(tw, bytes.NewReader([]byte("test1")))
	tw.Close()
	gw.Close()
	f.Close()

	digester := digest.SHA256.Digester()
	file, err := os.Open(tarfile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	_, err = io.Copy(digester.Hash(), file)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Rename(tarfile, filepath.Join(tmp1, "blobs", "sha256", digester.Digest().Hex()))
	if err != nil {
		t.Fatal(err)
	}

	testManifest := manifest{
		Layers: []v1.Descriptor{
			{
				MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
				Digest:    digester.Digest(),
			},
		},
	}
	err = testManifest.unpack(newPathWalker(tmp1), filepath.Join(tmp1, "rootfs"))
	if err != nil && !strings.Contains(err.Error(), "duplicate entry for") {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(tmp1, "rootfs"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err == nil {
		t.Fatal("Except partially unpacked file has been removed")
	}
}
