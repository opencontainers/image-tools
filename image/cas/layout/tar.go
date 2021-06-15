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

package layout

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-tools/image/cas"
	"golang.org/x/net/context"
)

// TarEngine is a cas.Engine backed by a tar file.
type TarEngine struct {
	file ReadWriteSeekCloser
}

// NewTarEngine returns a new TarEngine.
func NewTarEngine(ctx context.Context, file ReadWriteSeekCloser) (engine cas.Engine, err error) {
	engine = &TarEngine{
		file: file,
	}

	return engine, nil
}

// Get returns a reader for retrieving a blob from the store.
func (engine *TarEngine) Get(ctx context.Context, digest digest.Digest) (reader io.ReadCloser, err error) {
	targetName, err := blobPath(digest, "/")
	if err != nil {
		return nil, err
	}

	_, err = engine.file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(engine.file)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			return nil, os.ErrNotExist
		} else if err != nil {
			return nil, err
		}

		if header.Name == targetName {
			return ioutil.NopCloser(tarReader), nil
		}
	}
}

// Close releases resources held by the engine.
func (engine *TarEngine) Close() (err error) {
	return engine.file.Close()
}
