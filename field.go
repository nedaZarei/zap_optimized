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

package zap

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap/internal/stacktrace"
	"go.uber.org/zap/zapcore"
)

// Field is an alias for Field. Aliasing this type dramatically
// improves the navigability of this package's API documentation.
type Field = zapcore.Field

var (
	_minTimeInt64 = time.Unix(0, math.MinInt64)
	_maxTimeInt64 = time.Unix(0, math.MaxInt64)
)

// Skip constructs a no-op field, which is often useful when handling invalid
// inputs in other Field constructors.
func Skip() Field {
	return Field{Type: zapcore.SkipType}
}

// nilField returns a field which will marshal explicitly as nil.
// Implements this with consistent no-allocation for common pointer-typed fields, 
// avoiding generic Reflect where not needed.
var nilReflectField = Field{Type: zapcore.ReflectType, Interface: nil}

func nilField(key string) Field {
	// Use a static instance when possible, to avoid repeated allocations.
	// Must set the Key field as it's per-log statement.
	f := nilReflectField
	f.Key = key
	return f
}

// Binary constructs a field that carries an opaque binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
// use ByteString.
func Binary(key string, val []byte) Field {
	return Field{Key: key, Type: zapcore.BinaryType, Interface: val}
}

// Bool constructs a field that carries a bool.
func Bool(key string, val bool) Field {
	var ival int64
	if val {
		ival = 1
	}
	return Field{Key: key, Type: zapcore.BoolType, Integer: ival}
}

// Boolp constructs a field that carries a *bool. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Boolp(key string, val *bool) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.BoolType, Integer: boolToInt64(val)}
}
func boolToInt64(v *bool) int64 {
	if *v {
		return 1
	}
	return 0
}

// ByteString constructs a field that carries UTF-8 encoded text as a []byte.
// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
// Binary.
func ByteString(key string, val []byte) Field {
	return Field{Key: key, Type: zapcore.ByteStringType, Interface: val}
}

// Complex128 constructs a field that carries a complex number.
func Complex128(key string, val complex128) Field {
	return Field{Key: key, Type: zapcore.Complex128Type, Interface: val}
}

// Complex128p constructs a field that carries a *complex128.
func Complex128p(key string, val *complex128) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Complex128Type, Interface: *val}
}

// Complex64 constructs a field that carries a complex number.
func Complex64(key string, val complex64) Field {
	return Field{Key: key, Type: zapcore.Complex64Type, Interface: val}
}

// Complex64p constructs a field that carries a *complex64.
func Complex64p(key string, val *complex64) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Complex64Type, Interface: *val}
}

// Float64 constructs a field that carries a float64.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

// Float64p constructs a field that carries a *float64.
func Float64p(key string, val *float64) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(*val))}
}

// Float32 constructs a field that carries a float32.
func Float32(key string, val float32) Field {
	return Field{Key: key, Type: zapcore.Float32Type, Integer: int64(math.Float32bits(val))}
}

// Float32p constructs a field that carries a *float32.
func Float32p(key string, val *float32) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Float32Type, Integer: int64(math.Float32bits(*val))}
}

// Int constructs a field with the given key and value.
func Int(key string, val int) Field {
	return Int64(key, int64(val))
}

// Intp constructs a field that carries a *int.
func Intp(key string, val *int) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Int64Type, Integer: int64(*val)}
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

// Int64p constructs a field that carries a *int64.
func Int64p(key string, val *int64) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Int64Type, Integer: *val}
}

// Int32 constructs a field with the given key and value.
func Int32(key string, val int32) Field {
	return Field{Key: key, Type: zapcore.Int32Type, Integer: int64(val)}
}

// Int32p constructs a field that carries a *int32.
func Int32p(key string, val *int32) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Int32Type, Integer: int64(*val)}
}

// Int16 constructs a field with the given key and value.
func Int16(key string, val int16) Field {
	return Field{Key: key, Type: zapcore.Int16Type, Integer: int64(val)}
}

// Int16p constructs a field that carries a *int16.
func Int16p(key string, val *int16) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Int16Type, Integer: int64(*val)}
}

// Int8 constructs a field with the given key and value.
func Int8(key string, val int8) Field {
	return Field{Key: key, Type: zapcore.Int8Type, Integer: int64(val)}
}

// Int8p constructs a field that carries a *int8.
func Int8p(key string, val *int8) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Int8Type, Integer: int64(*val)}
}

// String constructs a field with the given key and value.
func String(key string, val string) Field {
	return Field{Key: key, Type: zapcore.StringType, String: val}
}

// Stringp constructs a field that carries a *string.
func Stringp(key string, val *string) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.StringType, String: *val}
}

// Uint constructs a field with the given key and value.
func Uint(key string, val uint) Field {
	return Uint64(key, uint64(val))
}

// Uintp constructs a field that carries a *uint.
func Uintp(key string, val *uint) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(*val)}
}

// Uint64 constructs a field with the given key and value.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Uint64p constructs a field that carries a *uint64.
func Uint64p(key string, val *uint64) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(*val)}
}

// Uint32 constructs a field with the given key and value.
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: zapcore.Uint32Type, Integer: int64(val)}
}

// Uint32p constructs a field that carries a *uint32.
func Uint32p(key string, val *uint32) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Uint32Type, Integer: int64(*val)}
}

// Uint16 constructs a field with the given key and value.
func Uint16(key string, val uint16) Field {
	return Field{Key: key, Type: zapcore.Uint16Type, Integer: int64(val)}
}

// Uint16p constructs a field that carries a *uint16.
func Uint16p(key string, val *uint16) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Uint16Type, Integer: int64(*val)}
}

// Uint8 constructs a field with the given key and value.
func Uint8(key string, val uint8) Field {
	return Field{Key: key, Type: zapcore.Uint8Type, Integer: int64(val)}
}

// Uint8p constructs a field that carries a *uint8.
func Uint8p(key string, val *uint8) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.Uint8Type, Integer: int64(*val)}
}

// Uintptr constructs a field with the given key and value.
func Uintptr(key string, val uintptr) Field {
	return Field{Key: key, Type: zapcore.UintptrType, Integer: int64(val)}
}

// Uintptrp constructs a field that carries a *uintptr.
func Uintptrp(key string, val *uintptr) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.UintptrType, Integer: int64(*val)}
}

// Reflect constructs a field with the given key and an arbitrary object.
func Reflect(key string, val interface{}) Field {
	return Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

// Namespace creates a named, isolated scope within the logger's context.
func Namespace(key string) Field {
	return Field{Key: key, Type: zapcore.NamespaceType}
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func Stringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Type: zapcore.StringerType, Interface: val}
}

// Time constructs a Field with the given key and value. The encoder
// controls how the time is serialized.
func Time(key string, val time.Time) Field {
	if val.Before(_minTimeInt64) || val.After(_maxTimeInt64) {
		return Field{Key: key, Type: zapcore.TimeFullType, Interface: val}
	}
	return Field{Key: key, Type: zapcore.TimeType, Integer: val.UnixNano(), Interface: val.Location()}
}

// Timep constructs a field that carries a *time.Time.
func Timep(key string, val *time.Time) Field {
	if val == nil {
		return nilField(key)
	}
	return Time(key, *val)
}

// Stack constructs a field that stores a stacktrace of the current goroutine
// under provided key.
func Stack(key string) Field {
	return StackSkip(key, 1) // skip Stack
}

// StackSkip constructs a field similarly to Stack, but also skips the given
// number of frames from the top of the stacktrace.
func StackSkip(key string, skip int) Field {
	return String(key, stacktrace.Take(skip+1)) // skip StackSkip
}

// Duration constructs a field with the given key and value.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: zapcore.DurationType, Integer: int64(val)}
}

// Durationp constructs a field that carries a *time.Duration.
func Durationp(key string, val *time.Duration) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.DurationType, Integer: int64(*val)}
}

// Object constructs a field with the given key and ObjectMarshaler.
func Object(key string, val zapcore.ObjectMarshaler) Field {
	if val == nil {
		return nilField(key)
	}
	return Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: val}
}

// Inline constructs a Field that is similar to Object, but it
// will add the elements of the provided ObjectMarshaler to the
// current namespace.
func Inline(val zapcore.ObjectMarshaler) Field {
	return zapcore.Field{
		Type:      zapcore.InlineMarshalerType,
		Interface: val,
	}
}

// Dict constructs a field containing the provided key-value pairs.
// It acts similar to [Object], but with the fields specified as arguments.
func Dict(key string, val ...Field) Field {
	return dictField(key, val)
}

// We need a function with the signature (string, T) for zap.Any.
func dictField(key string, val []Field) Field {
	return Object(key, dictObject(val))
}

type dictObject []Field

func (d dictObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, f := range d {
		f.AddTo(enc)
	}
	return nil
}

// DictObject constructs a [zapcore.ObjectMarshaler] with the given list of fields.
// The resulting object marshaler can be used as input to [Object], [Objects], or
// any other functions that expect an object marshaler.
func DictObject(val ...Field) zapcore.ObjectMarshaler {
	return dictObject(val)
}

// Custom ObjectMarshaler for map[string]string
type mapStringStringObject map[string]string

func (m mapStringStringObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range m {
		enc.AddString(k, v)
	}
	return nil
}

// Custom ObjectMarshaler for map[string]interface{}
type mapStringInterfaceObject map[string]interface{}

func (m mapStringInterfaceObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range m {
		// Use zap.Any for each value to get best handling
		Any(k, v).AddTo(enc)
	}
	return nil
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
func Any(key string, value interface{}) Field {
	switch v := value.(type) {
	case zapcore.ObjectMarshaler:
		return Object(key, v)
	case zapcore.ArrayMarshaler:
		return Array(key, v)
	case []Field:
		return dictField(key, v)
	case bool:
		return Bool(key, v)
	case *bool:
		return Boolp(key, v)
	case []bool:
		return Bools(key, v)
	case complex128:
		return Complex128(key, v)
	case *complex128:
		return Complex128p(key, v)
	case []complex128:
		return Complex128s(key, v)
	case complex64:
		return Complex64(key, v)
	case *complex64:
		return Complex64p(key, v)
	case []complex64:
		return Complex64s(key, v)
	case float64:
		return Float64(key, v)
	case *float64:
		return Float64p(key, v)
	case []float64:
		return Float64s(key, v)
	case float32:
		return Float32(key, v)
	case *float32:
		return Float32p(key, v)
	case []float32:
		return Float32s(key, v)
	case int:
		return Int(key, v)
	case *int:
		return Intp(key, v)
	case []int:
		return Ints(key, v)
	case int64:
		return Int64(key, v)
	case *int64:
		return Int64p(key, v)
	case []int64:
		return Int64s(key, v)
	case int32:
		return Int32(key, v)
	case *int32:
		return Int32p(key, v)
	case []int32:
		return Int32s(key, v)
	case int16:
		return Int16(key, v)
	case *int16:
		return Int16p(key, v)
	case []int16:
		return Int16s(key, v)
	case int8:
		return Int8(key, v)
	case *int8:
		return Int8p(key, v)
	case []int8:
		return Int8s(key, v)
	case string:
		return String(key, v)
	case *string:
		return Stringp(key, v)
	case []string:
		return Strings(key, v)
	case uint:
		return Uint(key, v)
	case *uint:
		return Uintp(key, v)
	case []uint:
		return Uints(key, v)
	case uint64:
		return Uint64(key, v)
	case *uint64:
		return Uint64p(key, v)
	case []uint64:
		return Uint64s(key, v)
	case uint32:
		return Uint32(key, v)
	case *uint32:
		return Uint32p(key, v)
	case []uint32:
		return Uint32s(key, v)
	case uint16:
		return Uint16(key, v)
	case *uint16:
		return Uint16p(key, v)
	case []uint16:
		return Uint16s(key, v)
	case uint8:
		return Uint8(key, v)
	case *uint8:
		return Uint8p(key, v)
	case []byte:
		return Binary(key, v)
	case uintptr:
		return Uintptr(key, v)
	case *uintptr:
		return Uintptrp(key, v)
	case []uintptr:
		return Uintptrs(key, v)
	case time.Time:
		return Time(key, v)
	case *time.Time:
		return Timep(key, v)
	case []time.Time:
		return Times(key, v)
	case time.Duration:
		return Duration(key, v)
	case *time.Duration:
		return Durationp(key, v)
	case []time.Duration:
		return Durations(key, v)
	case error:
		return NamedError(key, v)
	case []error:
		return Errors(key, v)
	case fmt.Stringer:
		return Stringer(key, v)
	case map[string]string:
		return Object(key, mapStringStringObject(v))
	case map[string]interface{}:
		return Object(key, mapStringInterfaceObject(v))
	default:
		return Reflect(key, value)
	}
}