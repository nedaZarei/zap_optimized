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
	"unsafe"
)

const _size = 1024 // by default, create 1 KiB buffers

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
	b.bs = append(b.bs, v...)
}

/*
AppendString writes a string to the Buffer.

Optimized for high-frequency usage in logging workloads. Uses batch copy for
large strings to avoid repeated slice growth and to leverage memmove
performance. Fall back to append for very small strings to avoid overhead.

If the buffer has sufficient capacity, grows without reallocating.
*/
func (b *Buffer) AppendString(s string) {
	// Fast path: empty string, nothing to do.
	if len(s) == 0 {
		return
	}

	l := len(b.bs)
	needed := len(s)
	capLeft := cap(b.bs) - l

	if needed == 0 {
		return
	}

	if capLeft >= needed {
		// Enough capacity, can grow in-place
		newSlice := b.bs[:l+needed]
		copy(newSlice[l:], s)
		b.bs = newSlice
		return
	}

	// Not enough capacity: grow efficiently (exponential growth).
	// Start with at least double the capacity or enough for the new string.
	newCap := cap(b.bs) * 2
	if newCap < l+needed {
		newCap = l + needed
	}
	if newCap < _size {
		newCap = _size
	}
	newBs := make([]byte, l+needed, newCap)
	copy(newBs, b.bs)
	copy(newBs[l:], s)
	b.bs = newBs
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

/*
String returns a string copy of the underlying byte slice.

This always allocates and copies the buffer data. If you want a zero-copy,
zero-allocation string, see UnsafeString().
*/
func (b *Buffer) String() string {
	return string(b.bs)
}

/*
UnsafeString returns a zero-copy string representation of the buffer
contents. WARNING: The returned string aliases the underlying buffer memory,
so it is only valid until the buffer is next modified or returned to its pool.

This method uses unsafe pointer conversions to achieve zero allocations. Use
this only when the consumer will not keep or reference the string beyond this
buffer's next modification.
*/
func (b *Buffer) UnsafeString() string {
	var s string
	bsHdr := (*[2]uintptr)(unsafe.Pointer(&b.bs))
	strHdr := (*[2]uintptr)(unsafe.Pointer(&s))
	strHdr[0] = bsHdr[0]
	strHdr[1] = uintptr(len(b.bs))
	return s
}

// Reset resets the underlying byte slice. Subsequent writes re-use the slice's
// backing array.
func (b *Buffer) Reset() {
	b.bs = b.bs[:0]
}

// Write implements io.Writer.
func (b *Buffer) Write(bs []byte) (int, error) {
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
// Callers must not retain references to the Buffer after calling Free.
func (b *Buffer) Free() {
	b.pool.put(b)
}