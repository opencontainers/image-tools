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

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/image-tools/version"
	image_spec "github.com/opencontainers/image-spec/specs-go"
	runtime_spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

// gitCommit will be the hash that the binary was built from
// and will be populated by the Makefile
var gitCommit = ""

func main() {
	app := cli.NewApp()
	app.Name = "oci-image-tool"
	if gitCommit != "" {
		app.Version = fmt.Sprintf("%s commit: %s", version.Version, gitCommit)
	} else {
		app.Version = version.Version
	}
	app.Description = fmt.Sprintf("Tools for working with OCI images.  Currently supported specifications are:\n\n   * OCI Image Format Specification: %s\n   * OCI Runtime Specification: %s", image_spec.Version, runtime_spec.Version)
	app.Usage = "OCI (Open Container Initiative) image tools"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Commands = []cli.Command{
		validateCommand,
		unpackCommand,
		createCommand,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
