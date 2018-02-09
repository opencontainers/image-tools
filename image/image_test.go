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
	"strconv"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	layoutStr = `{"imageLayoutVersion": "1.0.0"}`

	configStr = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Alyssa P. Hacker <alyspdev@example.com>",
    "architecture": "amd64",
    "os": "linux",
    "config": {
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
        "WorkingDir": "/home/alice",
        "Labels": {
            "com.example.project.git.url": "https://example.com/project.git",
            "com.example.project.git.commit": "45a939b2999782a3f005621a8d0f29aa387e1d6b"
        }
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
	select1 = []string{
		"org.opencontainers.ref.name=latest",
		"platform.os=linux",
	}

	select2 = []string{
		"org.opencontainers.ref.name=v1.0",
		"platform.os=linux",
	}

	indexJSON = `{
    "schemaVersion": 2,
    "manifests": [
      {
        "mediaType": "application/vnd.oci.image.index.v1+json",
        "size": <index_size>,
        "digest": "<index_digest>",
        "annotations": {
          "org.opencontainers.image.ref.name": "v1.0"
        }
      },
      {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "size": <manifest_size>,
        "digest": "<manifest_digest>",
        "platform": {
          "architecture": "ppc64le",
          "os": "linux"
        },
        "annotations": {
          "org.opencontainers.image.ref.name": "latest"
        }
      }
    ],
    "annotations": {
      "com.example.index.revision": "r124356"
    }
}
 `
	indexStr = `{
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": <manifest_size>,
      "digest": "<manifest_digest>",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": <manifest_size>,
      "digest": "<manifest_digest>",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    }
  ],
  "annotations": {
    "com.example.index.revision": "r124356"
  }
}
 `
	manifestStr = `{
    "annotations": {
        "org.freedesktop.specifications.metainfo.version": "1.0",
        "org.freedesktop.specifications.metainfo.type": "AppStream"
    },
    "config": {
        "digest": "<config_digest>",
        "mediaType": "application/vnd.oci.image.config.v1+json",
        "size": <config_size>
    },
    "layers": [
        {
            "digest": "<layer_digest>",
            "mediaType": "application/vnd.oci.image.layer.v1.tar",
            "size": <layer_size>
        }
    ],
    "schemaVersion": 2
}
 `
)

type tarContent struct {
	header *tar.Header
	b      []byte
}

type imageLayout struct {
	rootDir   string
	layout    string
	selects   []string
	manifest  string
	index     string
	config    string
	indexjson string
	tarList   []tarContent
}

func TestImageLayout(t *testing.T) {
	root, err := ioutil.TempDir("", "oci-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	dest1, err := ioutil.TempDir("", "dest1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dest1)

	dest2, err := ioutil.TempDir("", "dest2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dest2)

	dest3, err := ioutil.TempDir("", "dest3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dest3)

	dest4, err := ioutil.TempDir("", "dest4")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dest4)

	il := imageLayout{
		rootDir:   root,
		layout:    layoutStr,
		selects:   select1,
		manifest:  manifestStr,
		index:     indexStr,
		indexjson: indexJSON,
		config:    configStr,
		tarList: []tarContent{
			{&tar.Header{Name: "test", Size: 4, Mode: 0600}, []byte("test")},
		},
	}

	// create image layout bundle
	err = createImageLayoutBundle(il)
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateLayout(root, select1, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = UnpackLayout(root, dest1, "", select1)
	if err != nil {
		t.Fatal(err)
	}
	err = UnpackLayout(root, dest2, "linux:amd64", select2)
	if err != nil {
		t.Fatal(err)
	}
	err = CreateRuntimeBundleLayout(root, dest3, "rootfs", "", select1)
	if err != nil {
		t.Fatal(err)
	}
	err = CreateRuntimeBundleLayout(root, dest4, "rootfs", "linux:amd64", select2)
	if err != nil {
		t.Fatal(err)
	}
}

func createImageLayoutBundle(il imageLayout) error {
	err := os.MkdirAll(filepath.Join(il.rootDir, "blobs", "sha256"), 0700)
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
	il.manifest = strings.Replace(il.manifest, "<layer_digest>", string(desc.Digest), 1)
	il.manifest = strings.Replace(il.manifest, "<layer_size>", strconv.FormatInt(desc.Size, 10), 1)

	desc, err = createConfigFile(il.rootDir, il.config)
	if err != nil {
		return err
	}
	il.manifest = strings.Replace(il.manifest, "<config_digest>", string(desc.Digest), 1)
	il.manifest = strings.Replace(il.manifest, "<config_size>", strconv.FormatInt(desc.Size, 10), 1)

	// create manifest blob file
	desc, err = createManifestFile(il.rootDir, il.manifest)
	if err != nil {
		return err
	}
	il.index = strings.Replace(il.index, "<manifest_digest>", string(desc.Digest), -1)
	il.index = strings.Replace(il.index, "<manifest_size>", strconv.FormatInt(desc.Size, 10), -1)

	il.indexjson = strings.Replace(il.indexjson, "<manifest_digest>", string(desc.Digest), -1)
	il.indexjson = strings.Replace(il.indexjson, "<manifest_size>", strconv.FormatInt(desc.Size, 10), -1)

	// create index blob file
	desc, err = createIndexFile(il.rootDir, il.index)
	if err != nil {
		return err
	}
	il.indexjson = strings.Replace(il.indexjson, "<index_digest>", string(desc.Digest), -1)
	il.indexjson = strings.Replace(il.indexjson, "<index_size>", strconv.FormatInt(desc.Size, 10), -1)

	// create index.json file
	return createIndexJSON(il.rootDir, il.indexjson)
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

func createIndexJSON(root string, str string) error {
	indexpath := filepath.Join(root, "index.json")
	f, err := os.Create(indexpath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewBuffer([]byte(str)))

	return err
}

func createIndexFile(root, str string) (v1.Descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-index")
	f, err := os.Create(name)
	if err != nil {
		return v1.Descriptor{}, err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer([]byte(str)))
	if err != nil {
		return v1.Descriptor{}, err
	}

	return createHashedBlob(name)
}

func createManifestFile(root, str string) (v1.Descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-manifest")
	f, err := os.Create(name)
	if err != nil {
		return v1.Descriptor{}, err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer([]byte(str)))
	if err != nil {
		return v1.Descriptor{}, err
	}

	return createHashedBlob(name)
}

func createConfigFile(root, config string) (v1.Descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-config")
	f, err := os.Create(name)
	if err != nil {
		return v1.Descriptor{}, err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer([]byte(config)))
	if err != nil {
		return v1.Descriptor{}, err
	}

	return createHashedBlob(name)
}

func createImageLayerFile(root string, list []tarContent) (v1.Descriptor, error) {
	name := filepath.Join(root, "blobs", "sha256", "test-layer")
	err := createTarBlob(name, list)
	if err != nil {
		return v1.Descriptor{}, err
	}

	desc, err := createHashedBlob(name)
	if err != nil {
		return v1.Descriptor{}, err
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

func createHashedBlob(name string) (v1.Descriptor, error) {
	desc, err := newDescriptor(name)
	if err != nil {
		return v1.Descriptor{}, err
	}

	if err := desc.Digest.Validate(); err != nil {
		return v1.Descriptor{}, err
	}

	// Rename the file to hashed-digest name.
	err = os.Rename(name, filepath.Join(filepath.Dir(name), desc.Digest.Hex()))
	if err != nil {
		return v1.Descriptor{}, err
	}

	return desc, nil
}

func newDescriptor(name string) (v1.Descriptor, error) {
	file, err := os.Open(name)
	if err != nil {
		return v1.Descriptor{}, err
	}
	defer file.Close()

	digester := digest.SHA256.Digester()
	size, err := io.Copy(digester.Hash(), file)
	if err != nil {
		return v1.Descriptor{}, err
	}

	return v1.Descriptor{
		Digest: digester.Digest(),
		Size:   size,
	}, nil
}
