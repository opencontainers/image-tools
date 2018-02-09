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

// supported unpack types
var unpackTypes = []string{
	image.TypeImageLayout,
	image.TypeImage,
	image.TypeImageZip,
}

type unpackCmd struct {
	typ      string // the type to unpack, can be empty string
	selects  []string
	platform string
}

func unpackAction(context *cli.Context) error {
	if len(context.Args()) != 2 {
		return fmt.Errorf("both src and dest must be provided")
	}

	v := unpackCmd{
		typ:      context.String("type"),
		selects:  context.StringSlice("select"),
		platform: context.String("platform"),
	}

	if len(v.selects) == 0 {
		return fmt.Errorf("select must be provided")
	}

	for index, sel := range v.selects {
		for i := index + 1; i < len(v.selects); i++ {
			if sel == v.selects[i] {
				fmt.Printf("WARNING: selects contains duplicate selection %q.\n", v.selects[i])
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
		err = image.UnpackLayout(context.Args()[0], context.Args()[1], v.platform, v.selects)

	case image.TypeImageZip:
		err = image.UnpackZip(context.Args()[0], context.Args()[1], v.platform, v.selects)

	case image.TypeImage:
		err = image.UnpackFile(context.Args()[0], context.Args()[1], v.platform, v.selects)

	default:
		err = fmt.Errorf("cannot unpack %q", v.typ)
	}

	return err
}

var unpackCommand = cli.Command{
	Name:   "unpack",
	Usage:  "Unpack an image or image source layout",
	Action: unpackAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "type",
			Usage: fmt.Sprintf(
				`Type of the file to unpack. If unset, oci-image-tool will try to auto-detect the type. One of "%s".`,
				strings.Join(unpackTypes, ","),
			),
		},
		cli.StringSliceFlag{
			Name:  "select",
			Usage: "Select the search criteria for the validated reference, format is A=B. Only support 'org.opencontainers.ref.name', 'platform.os' and 'digest' three cases.",
		},
		cli.StringFlag{
			Name:  "platform",
			Usage: "Specify the os and architecture of the manifest, format is OS:Architecture. Only applicable if reftype is index.",
		},
	},
}
