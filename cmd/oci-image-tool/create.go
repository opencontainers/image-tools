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
	image.TypeImageZip,
}

type bundleCmd struct {
	typ      string // the type to bundle, can be empty string
	refs     []string
	root     string
	platform string
}

func createAction(context *cli.Context) error {
	if len(context.Args()) != 2 {
		return fmt.Errorf("both src and dest must be provided")
	}

	v := bundleCmd{
		typ:      context.String("type"),
		refs:     context.StringSlice("ref"),
		root:     context.String("rootfs"),
		platform: context.String("platform"),
	}

	if len(v.refs) == 0 {
		return fmt.Errorf("ref must be provided")
	}

	for index, ref := range v.refs {
		for i := index + 1; i < len(v.refs); i++ {
			if ref == v.refs[i] {
				fmt.Printf("WARNING: refs contains duplicate reference %q.\n", v.refs[i])
			}
		}
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
		err = image.CreateRuntimeBundleLayout(context.Args()[0], context.Args()[1], v.root, v.platform, v.refs)

	case image.TypeImageZip:
		err = image.CreateRuntimeBundleZip(context.Args()[0], context.Args()[1], v.root, v.platform, v.refs)

	case image.TypeImage:
		err = image.CreateRuntimeBundleFile(context.Args()[0], context.Args()[1], v.root, v.platform, v.refs)

	default:
		err = fmt.Errorf("cannot create %q", v.typ)

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
				`Type of the file to unpack. If unset, oci-image-tool will try to auto-detect the type. One of "%s".`,
				strings.Join(bundleTypes, ","),
			),
		},
		cli.StringSliceFlag{
			Name:  "ref",
			Usage: "A set of ref specify the search criteria for the validated reference, format is A=B. Only support 'name', 'platform.os' and 'digest' three cases.",
		},
		cli.StringFlag{
			Name:  "rootfs",
			Value: "rootfs",
			Usage: "A directory representing the root filesystem of the container in the OCI runtime bundle. It is strongly recommended to keep the default value.",
		},
		cli.StringFlag{
			Name:  "platform",
			Usage: "Specify the os and architecture of the manifest, format is OS:Architecture. Only applicable if reftype is index.",
		},
	},
}
