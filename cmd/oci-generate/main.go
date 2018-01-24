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

var generateTypes = []string{
	"imageLayout",
	"image",
}

type generateCmd struct {
	stdout  *log.Logger
	stderr  *log.Logger
	typ     string //the type to generate, can be empty string
	version bool
}

func main() {
	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	cmd := newGenerateCmd(stdout, stderr)
	if err := cmd.Execute(); err != nil {
		stderr.Println(err)
		os.Exit(1)
	}
}

func newGenerateCmd(stdout, stderr *log.Logger) *cobra.Command {
	v := &generateCmd{
		stdout: stdout,
		stderr: stderr,
	}

	cmd := &cobra.Command{
		Use:   "generate [dest]",
		Short: "Generate an image or an imageLayout",
		Long:  `Generate the OCI iamge or imageLayout to the destination directory [dest].`,
		Run:   v.Run,
	}

	cmd.Flags().StringVar(
		&v.typ, "type", "imageLayout",
		fmt.Sprintf(
			`Type of the file to generate, one of "%s".`,
			strings.Join(generateTypes, ","),
		),
	)

	cmd.Flags().BoolVarP(
		&v.version, "version", "v", false,
		`Print version information and exit`,
	)

	origHelp := cmd.HelpFunc()

	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		origHelp(c, args)
		stdout.Println("\nMore information:")
		stdout.Printf("\treferences\t%s\n", image.SpecURL)
		stdout.Printf("\tbug report\t%s\n", image.IssuesURL)
	})

	return cmd
}

func (v *generateCmd) Run(cmd *cobra.Command, args []string) {
	if v.version {
		v.stdout.Printf("commit: %s", gitCommit)
		v.stdout.Printf("spec: %s", specs.Version)
		os.Exit(0)
	}

	if len(args) != 1 {
		v.stderr.Print("dest must be provided")
		if err := cmd.Usage(); err != nil {
			v.stderr.Println(err)
		}
		os.Exit(1)
	}

	var err error
	switch v.typ {
	case "imageLayout":
		err = image.GenerateLayout(args[0])
	case "image":
		err = image.Generate(args[0])
	default:
		err = fmt.Errorf("unsupport type %q", v.typ)
	}

	if err != nil {
		v.stderr.Printf("generating failed: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
