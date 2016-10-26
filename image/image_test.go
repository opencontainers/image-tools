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
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	refTag = "latest"

	layoutStr = `{"imageLayoutVersion": "1.0.0"}`

	configStr = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Alyssa P. Hacker <alyspdev@example.com>",
    "architecture": "amd64",
    "os": "linux",
    "config": {
        "User": "alice",
        "Memory": 2048,
        "MemorySwap": 4096,
        "CpuShares": 8,
        "ExposedPorts": {
            "8080/tcp": {}
        },
        "Env": [
            "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
            "FOO=oci_is_a",
            "BAR=well_written_spec"
        ],
        "Entrypoint": [
            "/bin/my-app-binary"
        ],
        "Cmd": [
            "--foreground",
            "--config",
            "/etc/my-app.d/default.cfg"
        ],
        "Volumes": {
            "/var/job-result-data": {},
            "/var/log/my-app-logs": {}
        },
        "WorkingDir": "/home/alice"
    },
    "rootfs": {
      "diff_ids": [
        "sha256:c6f988f4874bb0add23a778f753c65efe992244e148a1d2ec2a8b664fb66bbd1",
        "sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef"
      ],
      "type": "layers"
    },
    "history": [
      {
        "created": "2015-10-31T22:22:54.690851953Z",
        "created_by": "/bin/sh -c #(nop) ADD file:a3bc1e842b69636f9df5256c49c5374fb4eef1e281fe3f282c65fb853ee171c5 in /"
      },
      {
        "created": "2015-10-31T22:22:55.613815829Z",
        "created_by": "/bin/sh -c #(nop) CMD [\"sh\"]",
        "empty_layer": true
      }
    ]
}
`
)

var (
	refStr = `{"digest":"<manifest_digest>","mediaType":"application/vnd.oci.image.manifest.v1+json","size":<manifest_size>}`

	manifestStr = `{
    "annotations": null,
    "config": {
        "digest": "<config_digest>",
        "mediaType": "application/vnd.oci.image.config.v1+json",
        "size": <config_size>
    },
    "layers": [
        {
            "digest": "<layer_digest>",
            "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
            "size": <layer_size>
        }
    ],
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "schemaVersion": 2
}
 `
)

type tarContent struct {
	header *tar.Header
	b      []byte
}

type imageLayout struct {
	rootDir  string
	layout   string
	ref      string
	manifest string
	config   string
	tarList  []tarContent
}

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

func createImageLayoutBundle(il imageLayout) error {
	err := os.MkdirAll(filepath.Join(il.rootDir, "blobs", "sha256"), 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(il.rootDir, "refs"), 0700)
	if err != nil {
		return err
	}

	// create image layout file
	err = createLayoutFile(il.rootDir)
	if err != nil {
		return err
	}

	// create image layer blob file.
	desc, err := createImageLayerFile(il.rootDir, il.tarList)
	if err != nil {
		return err
	}
	il.manifest = strings.Replace(il.manifest, "<layer_digest>", desc.Digest, 1)
	il.manifest = strings.Replace(il.manifest, "<layer_size>", strconv.FormatInt(desc.Size, 10), 1)

	desc, err = createConfigFile(il.rootDir, il.config)
	if err != nil {
		return err
	}
	il.manifest = strings.Replace(il.manifest, "<config_digest>", desc.Digest, 1)
	il.manifest = strings.Replace(il.manifest, "<config_size>", strconv.FormatInt(desc.Size, 10), 1)

	// create manifest blob file
	desc, err = createManifestFile(il.rootDir, il.manifest)
	if err != nil {
		return err
	}

	return createRefFile(il.rootDir, il.ref, desc)
}

func createLayoutFile(root string) error {
	layoutPath := filepath.Join(root, "oci-layout")
	f, err := os.Create(layoutPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewBuffer([]byte(layoutStr)))
	return err
}

func createRefFile(root, ref string, mft descriptor) error {
	refpath := filepath.Join(root, "refs", ref)
	f, err := os.Create(refpath)
	if err != nil {
		return err
	}
	defer f.Close()
	refStr = strings.Replace(refStr, "<manifest_digest>", mft.Digest, -1)
	refStr = strings.Replace(refStr, "<manifest_size>", strconv.FormatInt(mft.Size, 10), -1)
	_, err = io.Copy(f, bytes.NewBuffer([]byte(refStr)))
	return err
}

func createManifestFile(root, str string) (descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-manifest")
	f, err := os.Create(name)
	if err != nil {
		return descriptor{}, err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer([]byte(str)))
	if err != nil {
		return descriptor{}, err
	}

	return createHashedBlob(name)
}

func createConfigFile(root, config string) (descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-config")
	f, err := os.Create(name)
	if err != nil {
		return descriptor{}, err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer([]byte(config)))
	if err != nil {
		return descriptor{}, err
	}

	return createHashedBlob(name)
}

func createImageLayerFile(root string, list []tarContent) (descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-layer")
	err := createTarBlob(name, list)
	if err != nil {
		return descriptor{}, err
	}

	desc, err := createHashedBlob(name)
	if err != nil {
		return descriptor{}, err
	}

	desc.MediaType = v1.MediaTypeImageLayer
	return desc, nil
}

func createTarBlob(name string, list []tarContent) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, content := range list {
		if err = tarWriter.WriteHeader(content.header); err != nil {
			return err
		}
		if _, err = io.Copy(tarWriter, bytes.NewReader(content.b)); err != nil {
			return err
		}
	}
	return nil
}

func createHashedBlob(name string) (descriptor, error) {
	desc, err := newDescriptor(name)
	if err != nil {
		return descriptor{}, err
	}

	// Rename the file to hashed-digest name.
	err = os.Rename(name, filepath.Join(filepath.Dir(name), desc.Digest))
	if err != nil {
		return descriptor{}, err
	}

	//Normalize the hashed digest.
	desc.Digest = "sha256:" + desc.Digest

	return desc, nil
}

func newDescriptor(name string) (descriptor, error) {
	file, err := os.Open(name)
	if err != nil {
		return descriptor{}, err
	}
	defer file.Close()

	// generate sha256 hash
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return descriptor{}, err
	}

	return descriptor{
		Digest: fmt.Sprintf("%x", hash.Sum(nil)),
		Size:   size,
	}, nil
}
