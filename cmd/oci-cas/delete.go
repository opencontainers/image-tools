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
	"os"

	"github.com/opencontainers/image-tools/image/cas/layout"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type deleteCmd struct {
	path   string
	digest string
}

func newDeleteCmd() *cobra.Command {
	state := &deleteCmd{}

	return &cobra.Command{
		Use:   "delete PATH DIGEST",
		Short: "Remove a blob from from the store",
		Run:   state.Run,
	}
}

func (state *deleteCmd) Run(cmd *cobra.Command, args []string) {
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

func (state *deleteCmd) run() (err error) {
	ctx := context.Background()

	engine, err := layout.NewEngine(ctx, state.path)
	if err != nil {
		return err
	}
	defer engine.Close()

	return engine.Delete(ctx, state.digest)
}
