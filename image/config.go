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

package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/opencontainers/image-spec/schema"
	imagespecs "github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/image-tools/image/cas"
	runtimespecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func findConfig(ctx context.Context, engine cas.Engine, descriptor *imagespecs.Descriptor) (config *v1.Image, err error) {
	err = validateMediaType(descriptor.MediaType, []string{v1.MediaTypeImageConfig})
	if err != nil {
		return nil, errors.Wrap(err, "invalid config media type")
	}

	err = validateDescriptor(ctx, engine, descriptor)
	if err != nil {
		return nil, errors.Wrap(err, "invalid config descriptor")
	}

	reader, err := engine.Get(ctx, descriptor.Digest)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch %s", descriptor.Digest)
	}

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "%s: error reading manifest", descriptor.Digest)
	}

	if err := schema.MediaTypeImageConfig.Validate(bytes.NewReader(buf)); err != nil {
		return nil, errors.Wrapf(err, "%s: config validation failed", descriptor.Digest)
	}

	var c v1.Image
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, err
	}

	// check if the rootfs type is 'layers'
	if c.RootFS.Type != "layers" {
		return nil, fmt.Errorf("%q is an unknown rootfs type, MUST be 'layers'", c.RootFS.Type)
	}

	return &c, nil
}

func runtimeSpec(c *v1.Image, rootfs string) (*runtimespecs.Spec, error) {
	if c.OS != "linux" {
		return nil, fmt.Errorf("%s: unsupported OS", c.OS)
	}

	var s runtimespecs.Spec
	s.Version = runtimespecs.Version
	// we should at least apply the default spec, otherwise this is totally useless
	s.Process.Terminal = true
	s.Root.Path = rootfs
	s.Process.Cwd = "/"
	if c.Config.WorkingDir != "" {
		s.Process.Cwd = c.Config.WorkingDir
	}
	s.Process.Env = append(s.Process.Env, c.Config.Env...)
	s.Process.Args = append(s.Process.Args, c.Config.Entrypoint...)
	s.Process.Args = append(s.Process.Args, c.Config.Cmd...)

	if len(s.Process.Args) == 0 {
		s.Process.Args = append(s.Process.Args, "sh")
	}

	if uid, err := strconv.Atoi(c.Config.User); err == nil {
		s.Process.User.UID = uint32(uid)
	} else if ug := strings.Split(c.Config.User, ":"); len(ug) == 2 {
		uid, err := strconv.Atoi(ug[0])
		if err != nil {
			return nil, errors.New("config.User: unsupported uid format")
		}

		gid, err := strconv.Atoi(ug[1])
		if err != nil {
			return nil, errors.New("config.User: unsupported gid format")
		}

		s.Process.User.UID = uint32(uid)
		s.Process.User.GID = uint32(gid)
	} else if c.Config.User != "" {
		return nil, errors.New("config.User: unsupported format")
	}

	s.Platform.OS = c.OS
	s.Platform.Arch = c.Architecture

	mem := uint64(c.Config.Memory)
	swap := uint64(c.Config.MemorySwap)
	shares := uint64(c.Config.CPUShares)

	s.Linux.Resources = &runtimespecs.Resources{
		CPU: &runtimespecs.CPU{
			Shares: &shares,
		},

		Memory: &runtimespecs.Memory{
			Limit:       &mem,
			Reservation: &mem,
			Swap:        &swap,
		},
	}

	for vol := range c.Config.Volumes {
		s.Mounts = append(
			s.Mounts,
			runtimespecs.Mount{
				Destination: vol,
				Type:        "bind",
				Options:     []string{"rbind"},
			},
		)
	}

	return &s, nil
}
