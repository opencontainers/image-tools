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
)

// gitCommit will be the hash that the binary was built from
// and will be populated by the Makefile
var gitCommit = ""

// supported unpack types
var unpackTypes = []string{
	image.TypeImageLayout,
	image.TypeImage,
}

type unpackCmd struct {
	stdout  *log.Logger
	stderr  *log.Logger
	typ     string // the type to unpack, can be empty string
	ref     string
	version bool
}

func main() {
	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	cmd := newUnpackCmd(stdout, stderr)
	if err := cmd.Execute(); err != nil {
		stderr.Println(err)
		os.Exit(1)
	}
}

func newUnpackCmd(stdout, stderr *log.Logger) *cobra.Command {
	v := &unpackCmd{
		stdout: stdout,
		stderr: stderr,
	}

	cmd := &cobra.Command{
		Use:   "unpack [src] [dest]",
		Short: "Unpack an image or image source layout",
		Long:  `Unpack the OCI image .tar file or OCI image layout directory present at [src] to the destination directory [dest].`,
		Run:   v.Run,
	}

	cmd.Flags().StringVar(
		&v.typ, "type", "",
		fmt.Sprintf(
			`Type of the file to unpack. If unset, oci-unpack will try to auto-detect the type. One of "%s"`,
			strings.Join(unpackTypes, ","),
		),
	)

	cmd.Flags().StringVar(
		&v.ref, "ref", "v1.0",
		`The ref pointing to the manifest to be unpacked. This must be present in the "refs" subdirectory of the image.`,
	)
	cmd.Flags().BoolVar(
		&v.version, "version", false,
		`Print version information and exit`,
	)
	return cmd
}

func (v *unpackCmd) Run(cmd *cobra.Command, args []string) {
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
	case image.TypeImageLayout:
		err = image.UnpackLayout(args[0], args[1], v.ref)

	case image.TypeImage:
		err = image.Unpack(args[0], args[1], v.ref)

	default:
		err = fmt.Errorf("cannot unpack %q", v.typ)
	}

	if err != nil {
		v.stderr.Printf("unpacking failed: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
