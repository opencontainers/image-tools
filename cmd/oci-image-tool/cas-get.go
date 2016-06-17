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

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-tools/image/cas/layout"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

type casGetCmd struct {
	path   string
	digest digest.Digest
}

var casGetCommand = cli.Command{
	Name:      "get",
	Usage:     "Retrieve a blob from the store and write it to stdout.",
	ArgsUsage: "PATH DIGEST",
	Action:    casGetHandle,
}

func casGetHandle(ctx *cli.Context) (err error) {
	if ctx.NArg() != 2 {
		return fmt.Errorf("both PATH and DIGEST must be provided")
	}

	state := &casGetCmd{}
	state.path = ctx.Args().Get(0)
	state.digest, err = digest.Parse(ctx.Args().Get(1))
	if err != nil {
		return err
	}

	engineContext := context.Background()

	engine, err := layout.NewEngine(engineContext, state.path)
	if err != nil {
		return err
	}
	defer engine.Close()

	reader, err := engine.Get(engineContext, state.digest)
	if err != nil {
		return err
	}
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	n, err := os.Stdout.Write(bytes)
	if err != nil {
		return err
	}
	if n < len(bytes) {
		return fmt.Errorf("wrote %d of %d bytes", n, len(bytes))
	}

	return nil
}
