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

	"github.com/opencontainers/image-tools/image/refs/layout"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type listCmd struct {
	path string
}

func newListCmd() *cobra.Command {
	state := &listCmd{}

	return &cobra.Command{
		Use:   "list PATH",
		Short: "Return available names from the store.",
		Run:   state.Run,
	}
}

func (state *listCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "PATH must be provided")
		if err := cmd.Usage(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	state.path = args[0]

	err := state.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

func (state *listCmd) run() (err error) {
	ctx := context.Background()

	engine, err := layout.NewEngine(state.path)
	if err != nil {
		return err
	}
	defer engine.Close()

	return engine.List(ctx, "", -1, 0, state.printName)
}

func (state *listCmd) printName(ctx context.Context, name string) (err error) {
	n, err := fmt.Fprintln(os.Stdout, name)
	if n < len(name) {
		return fmt.Errorf("wrote %d of %d name", n, len(name))
	}
	return err
}
