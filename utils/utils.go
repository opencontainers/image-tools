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

package utils

import (
	"syscall"
)

// StatT type contains status of a file. It contains metadata
// like permission, owner, group, size, etc about a file.
type StatT struct {
	mode uint32
	uid  uint32
	gid  uint32
	rdev uint64
	size int64
	mtim syscall.Timespec
}

// Mode returns file's permission mode.
func (s StatT) Mode() uint32 {
	return s.mode
}

// UID returns file's user id of owner.
func (s StatT) UID() uint32 {
	return s.uid
}

// GID returns file's group id of owner.
func (s StatT) GID() uint32 {
	return s.gid
}

// Rdev returns file's device ID (if it's special file).
func (s StatT) Rdev() uint64 {
	return s.rdev
}

// Size returns file's size.
func (s StatT) Size() int64 {
	return s.size
}

// Mtim returns file's last modification time.
func (s StatT) Mtim() syscall.Timespec {
	return s.mtim
}

// GetLastModification returns file's last modification time.
func (s StatT) GetLastModification() syscall.Timespec {
	return s.Mtim()
}
