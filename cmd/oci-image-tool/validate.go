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
	image.TypeImage,
	image.TypeManifest,
	image.TypeImageIndex,
	image.TypeConfig,
}

type validateCmd struct {
	stdout  *log.Logger
	typ     string // the type to validate, can be empty string
	selects []string
}

var v validateCmd

func validateAction(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("no files specified")
	}

	v = validateCmd{
		typ:     context.String("type"),
		selects: context.StringSlice("select"),
	}

	if v.typ == "" {
		return fmt.Errorf("--type must be set")
	}

	for index, sel := range v.selects {
		for i := index + 1; i < len(v.selects); i++ {
			if sel == v.selects[i] {
				fmt.Printf("WARNING: selects contains duplicate selection %q.\n", v.selects[i])
			}
		}
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
			errs = append(errs, fmt.Sprintf("%s:%d:%d: %v", arg, serr.Line, serr.Col, err))
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", arg, err))
		}

	}

	if len(errs) > 0 {
		return fmt.Errorf("%d errors detected: \n%s", len(errs), strings.Join(errs, "\n"))
	}

	fmt.Println("Validation succeeded")
	return nil
}

func validatePath(name string) error {
	var typ = v.typ

	if v.stdout == nil {
		v.stdout = log.New(os.Stdout, "oci-image-tool: ", 0)
	}

	if typ == image.TypeImage {
		imageType, err := image.Autodetect(name)
		if err != nil {
			return errors.Wrap(err, "unable to determine image type")
		}
		fmt.Println("autodetected image file type is:", imageType)
		switch imageType {
		case image.TypeImageLayout:
			return image.ValidateLayout(name, v.selects, v.stdout)
		case image.TypeImageZip:
			return image.ValidateZip(name, v.selects, v.stdout)
		case image.TypeImage:
			return image.ValidateFile(name, v.selects, v.stdout)
		}
	}

	if len(v.selects) != 0 {
		fmt.Println("WARNING: selects are only appropriate if type is image")
	}
	f, err := os.Open(name) // nolint: errcheck, gosec
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
	Action: validateAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "type",
			Usage: fmt.Sprintf(
				`Type of the file to validate. One of "%s".`,
				strings.Join(validateTypes, ","),
			),
		},
		cli.StringSliceFlag{
			Name:  "select",
			Usage: "Select the search criteria for the validated reference, format is A=B. Only support 'org.opencontainers.ref.name', 'platform.os' and 'digest' three cases. Only applicable if type is image",
		},
	},
}
