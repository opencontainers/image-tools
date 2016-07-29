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

	"github.com/opencontainers/image-tools/image/layout"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type imageLayout struct{}

func newImageLayoutCmd() *cobra.Command {
	state := &imageLayout{}

	return &cobra.Command{
		Use:   "image-layout PATH",
		Short: "Initialize an OCI image-layout repository",
		Run:   state.Run,
	}
}

func (state *imageLayout) Run(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		if err := cmd.Usage(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	path := args[0]

	ctx := context.Background()

	err := layout.CreateTarFile(ctx, path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}
