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

// Package pools provides a collection of pools which provide various
// data types with buffers. These can be used to lower the number of
// memory allocations and reuse buffers.
package utils

import (
	"bufio"
	"io"
	"sync"
)

var (
	// BufioReader32KPool is a pool which returns bufio.Reader with a 32K buffer.
	BufioReader32KPool *BufioReaderPool
	// BufioWriter32KPool is a pool which returns bufio.Writer with a 32K buffer.
	BufioWriter32KPool *BufioWriterPool
)

const buffer32K = 32 * 1024

// BufioReaderPool is a bufio reader that uses sync.Pool.
type BufioReaderPool struct {
	pool sync.Pool
}

func init() {
	BufioReader32KPool = newBufioReaderPoolWithSize(buffer32K)
	BufioWriter32KPool = newBufioWriterPoolWithSize(buffer32K)
}

// newBufioReaderPoolWithSize is unexported because new pools should be
// added here to be shared where required.
func newBufioReaderPoolWithSize(size int) *BufioReaderPool {
	pool := sync.Pool{
		New: func() interface{} { return bufio.NewReaderSize(nil, size) },
	}
	return &BufioReaderPool{pool: pool}
}

// Get returns a bufio.Reader which reads from r. The buffer size is that of the pool.
func (bufPool *BufioReaderPool) Get(r io.Reader) *bufio.Reader {
	buf := bufPool.pool.Get().(*bufio.Reader)
	buf.Reset(r)
	return buf
}

// Put puts the bufio.Reader back into the pool.
func (bufPool *BufioReaderPool) Put(b *bufio.Reader) {
	b.Reset(nil)
	bufPool.pool.Put(b)
}

// Copy is a convenience wrapper which uses a buffer to avoid allocation in io.Copy.
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := BufioReader32KPool.Get(src)
	written, err = io.Copy(dst, buf)
	BufioReader32KPool.Put(buf)
	return
}

// BufioWriterPool is a bufio writer that uses sync.Pool.
type BufioWriterPool struct {
	pool sync.Pool
}

// newBufioWriterPoolWithSize is unexported because new pools should be
// added here to be shared where required.
func newBufioWriterPoolWithSize(size int) *BufioWriterPool {
	pool := sync.Pool{
		New: func() interface{} { return bufio.NewWriterSize(nil, size) },
	}
	return &BufioWriterPool{pool: pool}
}

// Get returns a bufio.Writer which writes to w. The buffer size is that of the pool.
func (bufPool *BufioWriterPool) Get(w io.Writer) *bufio.Writer {
	buf := bufPool.pool.Get().(*bufio.Writer)
	buf.Reset(w)
	return buf
}

// Put puts the bufio.Writer back into the pool.
func (bufPool *BufioWriterPool) Put(b *bufio.Writer) {
	b.Reset(nil)
	bufPool.pool.Put(b)
}
