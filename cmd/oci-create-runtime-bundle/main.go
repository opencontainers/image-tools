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
	"log"
	"os"
	"strings"

	specs "github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-tools/image"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// gitCommit will be the hash that the binary was built from
// and will be populated by the Makefile
var gitCommit = ""

// supported bundle types
var bundleTypes = []string{
	image.TypeImage,
}

type bundleCmd struct {
	stdout  *log.Logger
	stderr  *log.Logger
	typ     string // the type to bundle, can be empty string
	ref     string
	root    string
	version bool
}

func main() {
	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	cmd := newBundleCmd(stdout, stderr)
	if err := cmd.Execute(); err != nil {
		stderr.Println(err)
		os.Exit(1)
	}
}

func newBundleCmd(stdout, stderr *log.Logger) *cobra.Command {
	v := &bundleCmd{
		stdout: stdout,
		stderr: stderr,
	}

	cmd := &cobra.Command{
		Use:   "oci-create-runtime-bundle [src] [dest]",
		Short: "Create an OCI image runtime bundle",
		Long:  `Creates an OCI image runtime bundle at the destination directory [dest] from an OCI image present at [src].`,
		Run:   v.Run,
	}

	cmd.Flags().StringVar(
		&v.typ, "type", "",
		fmt.Sprintf(
			`Type of the file to unpack. If unset, oci-image-tool will try to auto-detect the type. One of "%s"`,
			strings.Join(bundleTypes, ","),
		),
	)

	cmd.Flags().StringVar(
		&v.ref, "ref", "v1.0",
		`The ref pointing to the manifest of the OCI image. This must be present in the "refs" subdirectory of the image.`,
	)

	cmd.Flags().StringVar(
		&v.root, "rootfs", "rootfs",
		`A directory representing the root filesystem of the container in the OCI runtime bundle.
It is strongly recommended to keep the default value.`,
	)

	cmd.Flags().BoolVar(
		&v.version, "version", false,
		`Print version information and exit`,
	)
	return cmd
}

func (v *bundleCmd) Run(cmd *cobra.Command, args []string) {
	if v.version {
		v.stdout.Printf("commit: %s", gitCommit)
		v.stdout.Printf("spec: %s", specs.Version)
		os.Exit(0)

	}
	if len(args) != 2 {
		v.stderr.Print("both src and dest must be provided")
		if err := cmd.Usage(); err != nil {
			v.stderr.Println(err)
		}
		os.Exit(1)
	}

	ctx := context.Background()

	if _, err := os.Stat(args[1]); os.IsNotExist(err) {
		v.stderr.Printf("destination path %s does not exist", args[1])
		os.Exit(1)
	}

	if v.typ == "" {
		typ, err := image.Autodetect(args[0])
		if err != nil {
			v.stderr.Printf("%q: autodetection failed: %v", args[0], err)
			os.Exit(1)
		}
		v.typ = typ
	}

	var err error
	switch v.typ {
	case image.TypeImage:
		err = image.CreateRuntimeBundle(ctx, args[0], args[1], v.ref, v.root)
	}

	if err != nil {
		v.stderr.Printf("unpacking failed: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
