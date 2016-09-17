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

	"github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-spec/specs-go/v1"
	caslayout "github.com/opencontainers/image-tools/image/cas/layout"
	imagelayout "github.com/opencontainers/image-tools/image/layout"
	"golang.org/x/net/context"
)

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
	if err := unpackLayer(tmp2, r); err != nil && !strings.Contains(err.Error(), "duplicate entry for") {
		t.Fatalf("Expected to fail with duplicate entry, got %v", err)
	}
}

func TestUnpackLayer(t *testing.T) {
	ctx := context.Background()

	tmp1, err := ioutil.TempDir("", "test-layer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp1)

	path := filepath.Join(tmp1, "image.tar")
	err = imagelayout.CreateDir(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := caslayout.NewEngine(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	layer, err := createTarBlob(ctx, engine, []tarContent{
		tarContent{&tar.Header{Name: "test", Size: 4, Mode: 0600}, []byte("test")},
	})
	if err != nil {
		t.Fatal(err)
	}

	testManifest := v1.Manifest{
		Layers: []specs.Descriptor{*layer},
	}

	err = unpackManifest(ctx, &testManifest, engine, filepath.Join(tmp1, "rootfs"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(tmp1, "rootfs", "test"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnpackLayerRemovePartialyUnpackedFile(t *testing.T) {
	ctx := context.Background()

	// generate a tar file has duplicate entry which will failed on unpacking
	tmp1, err := ioutil.TempDir("", "test-layer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp1)

	err = imagelayout.CreateDir(ctx, tmp1)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := caslayout.NewEngine(ctx, tmp1)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	layer, err := createTarBlob(ctx, engine, []tarContent{
		tarContent{&tar.Header{Name: "test", Size: 4, Mode: 0600}, []byte("test")},
		tarContent{&tar.Header{Name: "test", Size: 5, Mode: 0600}, []byte("test1")},
	})
	if err != nil {
		t.Fatal(err)
	}

	testManifest := v1.Manifest{
		Layers: []specs.Descriptor{*layer},
	}
	err = unpackManifest(ctx, &testManifest, engine, filepath.Join(tmp1, "rootfs"))
	if err != nil && !strings.Contains(err.Error(), "duplicate entry for") {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(tmp1, "rootfs"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err == nil {
		t.Fatal("Execpt partialy unpacked file has been removed")
	}
}
