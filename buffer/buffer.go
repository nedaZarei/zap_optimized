// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package buffer provides a thin wrapper around a byte slice. Unlike the
// standard library's bytes.Buffer, it supports a portion of the strconv
// package's zero-allocation formatters.
package buffer // import "go.uber.org/zap/buffer"

import (
	"strconv"
	"time"
)

const (
	_size         = 1024 // by default, create 1 KiB buffers
	_minSize      = 64   // minimum size to maintain in the pool
	_maxExcessCap = 4    // maximum excess capacity ratio before we trim the buffer
)

// Buffer is a thin wrapper around a byte slice. It's intended to be pooled, so
// the only way to construct one is via a Pool.
type Buffer struct {
	bs   []byte
	pool Pool
}

// AppendByte writes a single byte to the Buffer.
func (b *Buffer) AppendByte(v byte) {
	b.bs = append(b.bs, v)
}

// AppendBytes writes the given slice of bytes to the Buffer.
func (b *Buffer) AppendBytes(v []byte) {
	if len(v) == 0 {
		return
	}
	b.bs = append(b.bs, v...)
}

// AppendString writes a string to the Buffer.
func (b *Buffer) AppendString(s string) {
	if len(s) == 0 {
		return
	}
	b.bs = append(b.bs, s...)
}

// AppendInt appends an integer to the underlying buffer (assuming base 10).
func (b *Buffer) AppendInt(i int64) {
	b.bs = strconv.AppendInt(b.bs, i, 10)
}

// AppendTime appends the time formatted using the specified layout.
func (b *Buffer) AppendTime(t time.Time, layout string) {
	b.bs = t.AppendFormat(b.bs, layout)
}

// AppendUint appends an unsigned integer to the underlying buffer (assuming
// base 10).
func (b *Buffer) AppendUint(i uint64) {
	b.bs = strconv.AppendUint(b.bs, i, 10)
}

// AppendBool appends a bool to the underlying buffer.
func (b *Buffer) AppendBool(v bool) {
	b.bs = strconv.AppendBool(b.bs, v)
}

// AppendFloat appends a float to the underlying buffer. It doesn't quote NaN
// or +/- Inf.
func (b *Buffer) AppendFloat(f float64, bitSize int) {
	b.bs = strconv.AppendFloat(b.bs, f, 'f', -1, bitSize)
}

// Len returns the length of the underlying byte slice.
func (b *Buffer) Len() int {
	return len(b.bs)
}

// Cap returns the capacity of the underlying byte slice.
func (b *Buffer) Cap() int {
	return cap(b.bs)
}

// Bytes returns a mutable reference to the underlying byte slice.
func (b *Buffer) Bytes() []byte {
	return b.bs
}

// String returns a string copy of the underlying byte slice.
func (b *Buffer) String() string {
	return string(b.bs)
}

// Reset resets the underlying byte slice. Subsequent writes re-use the slice's
// backing array.
func (b *Buffer) Reset() {
	b.bs = b.bs[:0]
}

// EnsureCapacity ensures the buffer has at least the specified capacity.
// If the current buffer capacity is less than the requested capacity,
// the buffer will grow to accommodate it without multiple allocations.
func (b *Buffer) EnsureCapacity(capacity int) {
	if cap(b.bs) < capacity {
		// Create a new buffer with the desired capacity
		newBs := make([]byte, len(b.bs), capacity)
		copy(newBs, b.bs)
		b.bs = newBs
	}
}

// Write implements io.Writer.
func (b *Buffer) Write(bs []byte) (int, error) {
	if len(bs) == 0 {
		return 0, nil
	}
	
	// Pre-grow the buffer if necessary to avoid multiple allocations
	requiredCap := len(b.bs) + len(bs)
	if cap(b.bs) < requiredCap {
		b.EnsureCapacity(requiredCap)
	}
	
	b.bs = append(b.bs, bs...)
	return len(bs), nil
}

// WriteByte writes a single byte to the Buffer.
//
// Error returned is always nil, function signature is compatible
// with bytes.Buffer and bufio.Writer
func (b *Buffer) WriteByte(v byte) error {
	b.AppendByte(v)
	return nil
}

// WriteString writes a string to the Buffer.
//
// Error returned is always nil, function signature is compatible
// with bytes.Buffer and bufio.Writer
func (b *Buffer) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	
	// Pre-grow the buffer if necessary to avoid multiple allocations
	requiredCap := len(b.bs) + len(s)
	if cap(b.bs) < requiredCap {
		b.EnsureCapacity(requiredCap)
	}
	
	b.AppendString(s)
	return len(s), nil
}

// TrimNewline trims any final "\n" byte from the end of the buffer.
func (b *Buffer) TrimNewline() {
	if i := len(b.bs) - 1; i >= 0 {
		if b.bs[i] == '\n' {
			b.bs = b.bs[:i]
		}
	}
}

// Free returns the Buffer to its Pool.
// 
// If the buffer capacity has grown too large,
// it will be trimmed before returning to the pool to avoid
// holding onto excessive memory.
//
// Callers must not retain references to the Buffer after calling Free.
func (b *Buffer) Free() {
	// Optimize memory usage by trimming excessively large buffers
	// to avoid keeping large unused buffers in the pool
	currentCap := cap(b.bs)
	
	// If the buffer has grown too large, replace it with a smaller one
	// but still maintain a reasonable minimum size
	if currentCap > _size*_maxExcessCap {
		// Allocate a new buffer with the default size
		// We don't preserve the contents since this buffer is being freed
		b.bs = make([]byte, 0, _size)
	} else if len(b.bs) > 0 {
		// Otherwise just reset the buffer, keeping the capacity
		b.Reset()
	}
	
	b.pool.put(b)
}