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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-tools/image/layout"
	"github.com/opencontainers/image-tools/image/refs"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// DirEngine is a refs.Engine backed by a directory.
type DirEngine struct {
	path string
	temp string
}

// NewDirEngine returns a new DirEngine.
func NewDirEngine(ctx context.Context, path string) (eng refs.Engine, err error) {
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

// Put adds a new reference to the store.
func (engine *DirEngine) Put(ctx context.Context, name string, descriptor *specs.Descriptor) (err error) {
	var file *os.File
	file, err = ioutil.TempFile(engine.temp, "ref-")
	if err != nil {
		return err
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

	encoder := json.NewEncoder(file)
	err = encoder.Encode(descriptor)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	targetName, err := refPath(name, string(os.PathSeparator))
	if err != nil {
		return err
	}

	path := filepath.Join(engine.path, targetName)
	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return err
	}

	return os.Rename(file.Name(), path)
}

// Get returns a reference from the store.
func (engine *DirEngine) Get(ctx context.Context, name string) (descriptor *specs.Descriptor, err error) {
	targetName, err := refPath(name, string(os.PathSeparator))
	if err != nil {
		return nil, err
	}

	var file *os.File
	file, err = os.Open(filepath.Join(engine.path, targetName))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	var desc specs.Descriptor
	err = decoder.Decode(&desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

// List returns available names from the store.
func (engine *DirEngine) List(ctx context.Context, prefix string, size int, from int, nameCallback refs.ListNameCallback) (err error) {
	var i = 0

	pathPrefix, err := refPath(prefix, string(os.PathSeparator))
	if err != nil {
		return nil
	}
	var files []os.FileInfo
	files, err = ioutil.ReadDir(filepath.Join(engine.path, filepath.Dir(pathPrefix)))
	if err != nil {
		return err
	}
	for _, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := file.Name()
		if strings.HasPrefix(name, prefix) {
			i++
			if i > from {
				err = nameCallback(ctx, name)
				if err != nil {
					return err
				}
				if i-from == size {
					return nil
				}
			}
		}
	}

	return nil
}

// Delete removes a reference from the store.
func (engine *DirEngine) Delete(ctx context.Context, name string) (err error) {
	return layout.DirDelete(ctx, engine.path, name, refPath)
}

// Close releases resources held by the engine.
func (engine *DirEngine) Close() (err error) {
	return os.RemoveAll(engine.temp)
}
