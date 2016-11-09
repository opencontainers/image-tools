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
	"encoding/json"
	"fmt"
	"os"

	"github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-tools/image/refs/layout"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type putCmd struct {
	path string
	name string
}

func newPutCmd() *cobra.Command {
	state := &putCmd{}

	return &cobra.Command{
		Use:   "put PATH NAME",
		Short: "Write a reference to the store",
		Long:  "Read descriptor JSON from stdin and write it to the store.",
		Run:   state.Run,
	}
}

func (state *putCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		if err := cmd.Usage(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	state.path = args[0]
	state.name = args[1]

	err := state.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

func (state *putCmd) run() (err error) {
	ctx := context.Background()

	engine, err := layout.NewEngine(ctx, state.path)
	if err != nil {
		return err
	}
	defer engine.Close()

	decoder := json.NewDecoder(os.Stdin)
	var descriptor specs.Descriptor
	err = decoder.Decode(&descriptor)
	if err != nil {
		return err
	}

	return engine.Put(ctx, state.name, &descriptor)
}
