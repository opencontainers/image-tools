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
	"github.com/opencontainers/image-tools/logger"
	"github.com/urfave/cli"
)

// supported unpack types
var unpackTypes = []string{
	image.TypeImageLayout,
	image.TypeImage,
}

type unpackCmd struct {
	typ string // the type to unpack, can be empty string
	ref string
}

func unpackHandle(context *cli.Context) error {
	if len(context.Args()) != 2 {
		return fmt.Errorf("both src and dest must be provided")
	}

	v := unpackCmd{
		typ: context.String("type"),
		ref: context.String("ref"),
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
		err = image.UnpackLayout(context.Args()[0], context.Args()[1], v.ref)

	case image.TypeImage:
		err = image.UnpackFile(context.Args()[0], context.Args()[1], v.ref)

	default:
		err = fmt.Errorf("cannot unpack %q", v.typ)
	}

	if err != nil {
		logger.G(globalCtx).WithError(err).Errorf("unpacking failed")
	}
	return err
}

var unpackCommand = cli.Command{
	Name:   "unpack",
	Usage:  "Unpack an image or image source layout",
	Action: unpackHandle,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "type",
			Usage: fmt.Sprintf(
				`Type of the file to unpack. If unset, oci-image-tool will try to auto-detect the type. One of "%s".`,
				strings.Join(unpackTypes, ","),
			),
		},
		cli.StringFlag{
			Name:  "ref",
			Value: "v1.0",
			Usage: "The ref pointing to the manifest of the OCI image. This must be present in the 'refs' subdirectory of the image.",
		},
	},
}
