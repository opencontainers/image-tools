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
	"strconv"
	"strings"
	"testing"

	"github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-spec/specs-go/v1"
	cas "github.com/opencontainers/image-tools/image/cas"
	caslayout "github.com/opencontainers/image-tools/image/cas/layout"
	imagelayout "github.com/opencontainers/image-tools/image/layout"
	refslayout "github.com/opencontainers/image-tools/image/refs/layout"
	"golang.org/x/net/context"
)

const (
	refTag = "latest"

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
            "mediaType": "application/vnd.oci.image.layer.tar+gzip",
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

func TestValidateLayout(t *testing.T) {
	ctx := context.Background()

	root, err := ioutil.TempDir("", "oci-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	err = imagelayout.CreateDir(ctx, root)
	if err != nil {
		t.Fatal(err)
	}

	casEngine, err := caslayout.NewEngine(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	defer casEngine.Close()

	refsEngine, err := refslayout.NewEngine(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	defer refsEngine.Close()

	layer, err := createTarBlob(ctx, casEngine, []tarContent{
		tarContent{&tar.Header{Name: "test", Size: 4, Mode: 0600}, []byte("test")},
	})
	if err != nil {
		t.Fatal(err)
	}

	digest, err := casEngine.Put(ctx, strings.NewReader(configStr))
	config := specs.Descriptor{
		Digest: digest,
		MediaType: v1.MediaTypeImageConfig,
		Size: int64(len(configStr)),
	}

	_manifest := manifestStr
	_manifest = strings.Replace(_manifest, "<config_digest>", config.Digest, 1)
	_manifest = strings.Replace(_manifest, "<config_size>", strconv.FormatInt(config.Size, 10), 1)
	_manifest = strings.Replace(_manifest, "<layer_digest>", layer.Digest, 1)
	_manifest = strings.Replace(_manifest, "<layer_size>", strconv.FormatInt(layer.Size, 10), 1)
	digest, err = casEngine.Put(ctx, strings.NewReader(_manifest))
	manifest := specs.Descriptor{
		Digest: digest,
		MediaType: v1.MediaTypeImageManifest,
		Size: int64(len(_manifest)),
	}

	err = refsEngine.Put(ctx, refTag, &manifest)
	if err != nil {
		t.Fatal(err)
	}

	err = Validate(ctx, root, []string{refTag}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func createTarBlob(ctx context.Context, engine cas.Engine, list []tarContent) (descriptor *specs.Descriptor, err error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, content := range list {
		if err = tarWriter.WriteHeader(content.header); err != nil {
			return nil, err
		}
		if _, err = io.Copy(tarWriter, bytes.NewReader(content.b)); err != nil {
			return nil, err
		}
	}

	err = tarWriter.Close()
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	var desc = specs.Descriptor{
		MediaType: v1.MediaTypeImageLayer,
		Size: int64(buffer.Len()),
	}

	digest, err := engine.Put(ctx, &buffer)
	if err != nil {
		return nil, err
	}

	desc.Digest = digest

	return &desc, nil
}
