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

	"github.com/opencontainers/image-tools/image/cas/layout"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type getCmd struct {
	path   string
	digest string
}

func newGetCmd() *cobra.Command {
	state := &getCmd{}

	return &cobra.Command{
		Use:   "get PATH DIGEST",
		Short: "Retrieve a blob from the store",
		Long:  "Retrieve a blob from the store and write it to stdout.",
		Run:   state.Run,
	}
}

func (state *getCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "both PATH and DIGEST must be provided")
		if err := cmd.Usage(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	state.path = args[0]
	state.digest = args[1]

	err := state.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

func (state *getCmd) run() (err error) {
	ctx := context.Background()

	engine, err := layout.NewEngine(state.path)
	if err != nil {
		return err
	}
	defer engine.Close()

	reader, err := engine.Get(ctx, state.digest)
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
