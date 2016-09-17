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

// Package layout implements the refs interface using the image-spec's
// image-layout [1].
//
// [1]: https://github.com/opencontainers/image-spec/blob/master/image-layout.md
package layout

import (
	"fmt"
	"os"
	"strings"

	"github.com/opencontainers/image-tools/image/refs"
	"golang.org/x/net/context"
)

// NewEngine instantiates an engine with the appropriate backend (tar,
// HTTP, ...).
func NewEngine(ctx context.Context, path string) (engine refs.Engine, err error) {
	engine, err = NewDirEngine(ctx, path)
	if err == nil {
		return engine, err
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err == nil {
		return NewTarEngine(ctx, file)
	}

	return nil, fmt.Errorf("unrecognized engine at %q", path)
}

// refPath returns the PATH to the NAME reference.  SEPARATOR selects
// the path separator used between components.
func refPath(name string, separator string) (path string, err error) {
	components := []string{".", "refs", name}
	return strings.Join(components, separator), nil
}
