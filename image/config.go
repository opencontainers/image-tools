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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opencontainers/image-spec/schema"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

type config v1.Image

func findConfig(w walker, d *descriptor) (*config, error) {
	var c config
	cpath := filepath.Join("blobs", d.algo(), d.hash())

	switch err := w.walk(func(path string, info os.FileInfo, r io.Reader) error {
		if info.IsDir() || filepath.Clean(path) != cpath {
			return nil
		}
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrapf(err, "%s: error reading config", path)
		}

		if err := schema.MediaTypeImageConfig.Validate(bytes.NewReader(buf)); err != nil {
			return errors.Wrapf(err, "%s: config validation failed", path)
		}

		if err := json.Unmarshal(buf, &c); err != nil {
			return err
		}
		// check if the rootfs type is 'layers'
		if c.RootFS.Type != "layers" {
			return fmt.Errorf("%q is an unknown rootfs type, MUST be 'layers'", c.RootFS.Type)
		}
		return errEOW
	}); err {
	case nil:
		return nil, fmt.Errorf("%s: config not found", cpath)
	case errEOW:
		return &c, nil
	default:
		return nil, err
	}
}

func (c *config) runtimeSpec(rootfs string) (*specs.Spec, error) {
	if c.OS != "linux" {
		return nil, fmt.Errorf("%s: unsupported OS", c.OS)
	}

	// TODO: Get this default from somewhere, rather than hardcoding it here.
	// runC has a default as well, maybe the two should be consolidated in the
	// runtime-spec (or in the runtime-tooling)?
	s := specs.Spec{
		Version: specs.Version,
		Platform: specs.Platform{
			OS:   "unknown",
			Arch: "unknown",
		},
		Process: specs.Process{
			Terminal: true,
			Cwd:      "/",
			Env:      []string{},
			Args:     []string{},
			User: specs.User{
				UID: 0,
				GID: 0,
			},
		},
		Root: specs.Root{
			Path:     "rootfs",
			Readonly: false,
		},
		// These are all required by the runtime-spec to be included. The
		// actual options are from the default runC spec.
		Mounts: []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
				Options:     nil,
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev", "ro"},
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
			},
		},
		Linux: &specs.Linux{
			MaskedPaths: []string{
				"/proc/kcore",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/sys/firmware",
			},
			ReadonlyPaths: []string{
				"/proc/asound",
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
			// We need at least CLONE_NEWNS in order to set up the mounts.
			// This is also required to make sure we have a sane rootfs setup.
			Namespaces: []specs.Namespace{
				{
					Type: "mount",
				},
			},
		},
	}

	// Now fill all of the fields.
	s.Root.Path = rootfs
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

	s.Linux.Resources = &specs.Resources{
		CPU: &specs.CPU{
			Shares: &shares,
		},

		Memory: &specs.Memory{
			Limit:       &mem,
			Reservation: &mem,
			Swap:        &swap,
		},
	}

	for vol := range c.Config.Volumes {
		s.Mounts = append(
			s.Mounts,
			specs.Mount{
				Destination: vol,
				Type:        "bind",
				Options:     []string{"rbind"},
			},
		)
	}

	return &s, nil
}
