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
	"strings"

	"github.com/opencontainers/image-tools/image"
	"github.com/urfave/cli"
)

// supported bundle types
var bundleTypes = []string{
	image.TypeImageLayout,
	image.TypeImage,
}

type bundleCmd struct {
	typ  string // the type to bundle, can be empty string
	ref  string
	root string
}

func createAction(context *cli.Context) error {
	if len(context.Args()) != 2 {
		return fmt.Errorf("both src and dest must be provided")
	}

	v := bundleCmd{
		typ:  context.String("type"),
		ref:  context.String("ref"),
		root: context.String("rootfs"),
	}

	if v.typ == "" {
		typ, err := image.Autodetect(context.Args()[0])
		if err != nil {
			return fmt.Errorf("%q: autodetection failed: %v", context.Args()[0], err)
		}
		v.typ = typ
	}

	var err error
	switch v.typ {
	case image.TypeImageLayout:
		err = image.CreateRuntimeBundleLayout(context.Args()[0], context.Args()[1], v.ref, v.root)

	case image.TypeImage:
		err = image.CreateRuntimeBundleFile(context.Args()[0], context.Args()[1], v.ref, v.root)

	default:
		err = fmt.Errorf("cannot create %q", v.typ)

	}

	if err != nil {
		fmt.Printf("creating failed: %v\n", err)
	}

	return err
}

var createCommand = cli.Command{
	Name:   "create",
	Usage:  "Create an OCI image runtime bundle",
	Action: createAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "type",
			Usage: fmt.Sprintf(
				`Type of the file to unpack. If unset, oci-image-tool-validate will try to auto-detect the type. One of "%s".`,
				strings.Join(bundleTypes, ","),
			),
		},
		cli.StringFlag{
			Name:  "ref",
			Value: "v1.0",
			Usage: "The ref pointing to the manifest of the OCI image. This must be present in the 'refs' subdirectory of the image.",
		},
		cli.StringFlag{
			Name:  "rootfs",
			Value: "rootfs",
			Usage: "A directory representing the root filesystem of the container in the OCI runtime bundle. It is strongly recommended to keep the default value.",
		},
	},
}
