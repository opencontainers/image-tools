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
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/image-tools/image/cas"
	"github.com/opencontainers/image-tools/image/layout"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// DirEngine is a cas.Engine backed by a directory.
type DirEngine struct {
	path string
	temp string
}

// NewDirEngine returns a new DirEngine.
func NewDirEngine(ctx context.Context, path string) (eng cas.Engine, err error) {
	engine := &DirEngine{
		path: path,
	}

	err = layout.CheckDirVersion(ctx, engine.path)
	if err != nil {
		return nil, err
	}

	tempDir, err := ioutil.TempDir(path, "tmp-")
	if err != nil {
		return nil, err
	}
	engine.temp = tempDir

	return engine, nil
}

// Put adds a new blob to the store.
func (engine *DirEngine) Put(ctx context.Context, reader io.Reader) (digest string, err error) {
	hash := sha256.New()
	algorithm := "sha256"

	var file *os.File
	file, err = ioutil.TempFile(engine.temp, "blob-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			err2 := os.Remove(file.Name())
			if err2 != nil {
				err = errors.Wrap(err, err2.Error())
			}
		}
	}()
	defer file.Close()

	hashingWriter := io.MultiWriter(file, hash)
	_, err = io.Copy(hashingWriter, reader)
	if err != nil {
		return "", err
	}
	file.Close()

	digest = fmt.Sprintf("%s:%x", algorithm, hash.Sum(nil))
	targetName, err := blobPath(digest, string(os.PathSeparator))
	if err != nil {
		return "", err
	}

	path := filepath.Join(engine.path, targetName)
	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return "", err
	}

	err = os.Rename(file.Name(), path)
	if err != nil {
		return "", err
	}

	return digest, nil
}

// Get returns a reader for retrieving a blob from the store.
func (engine *DirEngine) Get(ctx context.Context, digest string) (reader io.ReadCloser, err error) {
	targetName, err := blobPath(digest, string(os.PathSeparator))
	if err != nil {
		return nil, err
	}

	return os.Open(filepath.Join(engine.path, targetName))
}

// Delete removes a blob from the store.
func (engine *DirEngine) Delete(ctx context.Context, digest string) (err error) {
	return layout.DirDelete(ctx, engine.path, digest, blobPath)
}

// Close releases resources held by the engine.
func (engine *DirEngine) Close() (err error) {
	return os.RemoveAll(engine.temp)
}
