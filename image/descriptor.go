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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/opencontainers/image-spec/specs-go"
	"github.com/opencontainers/image-tools/image/cas"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func validateMediaType(mediaType string, mediaTypes []string) error {
	for _, mt := range mediaTypes {
		if mt == mediaType {
			return nil
		}
	}
	return fmt.Errorf("invalid media type %q", mediaType)
}

func validateDescriptor(ctx context.Context, engine cas.Engine, descriptor *specs.Descriptor) error {
	reader, err := engine.Get(ctx, descriptor.Digest)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch %s", descriptor.Digest)
	}

	return validateContent(ctx, descriptor, reader)
}

func validateContent(ctx context.Context, descriptor *specs.Descriptor, r io.Reader) error {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return errors.Wrap(err, "error generating hash")
	}

	digest := "sha256:" + hex.EncodeToString(h.Sum(nil))

	if digest != descriptor.Digest {
		return errors.New("digest mismatch")
	}

	if n != descriptor.Size {
		return errors.New("size mismatch")
	}

	return nil
}
