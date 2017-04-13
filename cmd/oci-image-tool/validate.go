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

	"github.com/opencontainers/image-spec/schema"
	"github.com/opencontainers/image-tools/image"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// supported validation types
var validateTypes = []string{
	image.TypeImageLayout,
	image.TypeImage,
	image.TypeManifest,
	image.TypeImageIndex,
	image.TypeConfig,
}

type validateCmd struct {
	stdout *log.Logger
	typ    string // the type to validate, can be empty string
	refs   []string
}

var v validateCmd

func validateHandler(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("no files specified")
	}

	v = validateCmd{
		typ:  context.String("type"),
		refs: context.StringSlice("ref"),
	}

	var errs []string
	for _, arg := range context.Args() {
		err := validatePath(arg)

		if err == nil {
			fmt.Printf("%s: OK\n", arg)
			continue
		}

		if verr, ok := errors.Cause(err).(schema.ValidationError); ok {
			errs = append(errs, fmt.Sprintf("%v", verr.Errs))
		} else if serr, ok := errors.Cause(err).(*schema.SyntaxError); ok {
			errs = append(errs, fmt.Sprintf("%s:%d:%d: validation failed: %v", arg, serr.Line, serr.Col, err))
			continue
		} else {
			errs = append(errs, fmt.Sprintf("%s: validation failed: %v", arg, err))
			continue
		}

	}

	if len(errs) > 0 {
		return fmt.Errorf("%d errors detected: \n%s", len(errs), strings.Join(errs, "\n"))
	}
	fmt.Println("Validation succeeded")
	return nil
}

func validatePath(name string) error {
	var (
		err error
		typ = v.typ
	)

	if typ == "" {
		if typ, err = image.Autodetect(name); err != nil {
			return errors.Wrap(err, "unable to determine type")
		}
	}

	if v.stdout == nil {
		v.stdout = log.New(os.Stdout, "oci-image-tool: ", 0)
	}

	switch typ {
	case image.TypeImageLayout:
		return image.ValidateLayout(name, v.refs, v.stdout)
	case image.TypeImage:
		return image.ValidateFile(name, v.refs, v.stdout)
	}

	if len(v.refs) != 0 {
		fmt.Printf("WARNING: type %q does not support refs, which are only appropriate if type is image or imageLayout.\n", typ)
	}

	f, err := os.Open(name)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}
	defer f.Close()

	switch typ {
	case image.TypeManifest:
		return schema.ValidatorMediaTypeManifest.Validate(f)
	case image.TypeImageIndex:
		return schema.ValidatorMediaTypeImageIndex.Validate(f)
	case image.TypeConfig:
		return schema.ValidatorMediaTypeImageConfig.Validate(f)
	}

	return fmt.Errorf("type %q unimplemented", typ)
}

var validateCommand = cli.Command{
	Name:   "validate",
	Usage:  "Validate one or more image files",
	Action: validateHandler,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "type",
			Usage: fmt.Sprintf(
				`Type of the file to validate. If unset, oci-image-tool will try to auto-detect the type. One of "%s".`,
				strings.Join(validateTypes, ","),
			),
		},
		cli.StringSliceFlag{
			Name:  "ref",
			Usage: "A set of refs pointing to the manifests to be validated. Each reference must be present in the refs subdirectory of the image. Only applicable if type is image or imageLayout.",
		},
	},
}
