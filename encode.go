// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
)

func Marshal(v interface{}) ([]byte, error) {
	enc := NewEncoder(nil)
	if err := enc.marshal(v); err != nil {
		return nil, err
	}
	return enc.Bytes(), nil
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	enc := NewEncoder(nil)
	if err := enc.marshalIndent(v, prefix, indent); err != nil {
		return nil, err
	}
	return enc.Bytes(), nil
}

type bytesBuffer struct {
	bytes.Buffer
}

type Encoder struct {
	w io.Writer
	bytesBuffer
	indentMode bool
	prefix     string
	indent     string
	depth      int
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (enc *Encoder) marshal(v interface{}) error {
	if err := enc.encode(reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) marshalIndent(v interface{}, prefix, indent string) error {
	enc.indentMode = true
	defer func() {
		enc.indentMode = false
	}()
	enc.prefix = prefix
	enc.indent = indent
	enc.depth = 0
	enc.WriteString(prefix)
	return enc.marshal(v)
}

func (enc *Encoder) Encode(v interface{}) error {
	if err := enc.marshal(v); err != nil {
		return err
	}
	_, err := enc.w.Write(enc.Bytes())
	return err
}

// encode never returns an error, it may panics with bytes.ErrTooLarge.
func (enc *Encoder) encode(v reflect.Value) error {
	if !v.IsValid() {
		enc.encodeNil()
		return nil
	}
	switch v.Kind() {
	case reflect.Bool:
		enc.encodeBool(v)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		enc.encodeInt(v)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		enc.encodeUint(v)
	case reflect.Float32:
		enc.encodeFloat(v, 32)
	case reflect.Float64:
		enc.encodeFloat(v, 64)
	case reflect.Complex64:
		enc.encodeComplex(v, 32)
	case reflect.Complex128:
		enc.encodeComplex(v, 64)
	case reflect.String:
		enc.encodeString(v)
	case reflect.Array:
		enc.encodeArray(v)
	case reflect.Slice:
		enc.encodeSlice(v)
	case reflect.Struct:
		enc.encodeStruct(v)
	case reflect.Map:
		enc.encodeMap(v)
	case reflect.Ptr, reflect.Interface:
		enc.encodeRef(v)
	default:
		// case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("unsupported variable type: %s", v.Type().String())
	}
	return nil
}

func (enc *Encoder) encodeBool(v reflect.Value) {
	enc.WriteString(strconv.FormatBool(v.Bool()))
}

func (enc *Encoder) encodeInt(v reflect.Value) {
	enc.WriteString(strconv.FormatInt(v.Int(), 10))
}

func (enc *Encoder) encodeUint(v reflect.Value) {
	enc.WriteString(strconv.FormatUint(v.Uint(), 10))
}

func (enc *Encoder) encodeFloat(v reflect.Value, bit int) {
	enc.WriteString(strconv.FormatFloat(v.Float(), 'g', -1, bit))
}

func (enc *Encoder) newLine() {
	enc.WriteByte('\n')
	enc.WriteString(enc.prefix)
	for i := 0; i < enc.depth; i++ {
		enc.WriteString(enc.indent)
	}
}

func (enc *Encoder) encodeStruct(v reflect.Value) {
	t := v.Type()
	enc.listStart(t.NumField())
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			enc.listSep()
		}
		enc.WriteString(t.Field(i).Name)
		enc.WriteString(": ")
		enc.encode(v.Field(i))
	}
	enc.listEnd(t.NumField())
}

func (enc *Encoder) listStart(count int) {
	enc.WriteByte('{')
	if enc.indentMode {
		if count > 0 {
			enc.depth++
			enc.newLine()
		}
	}
}

func (enc *Encoder) listSep() {
	if enc.indentMode {
		enc.WriteByte(',')
		enc.newLine()
	} else {
		enc.WriteString(", ")
	}
}

func (enc *Encoder) listEnd(count int) {
	if enc.indentMode {
		if count > 0 {
			enc.depth--
			enc.WriteByte(',')
			enc.newLine()
		}
	}
	enc.WriteByte('}')
}

func (enc *Encoder) encodeKey(v reflect.Value) {
	m := enc.indentMode
	enc.indentMode = false
	defer func() {
		enc.indentMode = m
	}()

	enc.encode(v)
}

func (enc *Encoder) encodeMap(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	enc.listStart(v.Len())
	var sv stringValues = v.MapKeys()
	sort.Sort(sv)
	for i, k := range sv {
		if i > 0 {
			enc.listSep()
		}
		enc.encodeKey(k)
		enc.WriteString(": ")
		enc.encode(v.MapIndex(k))
	}
	enc.listEnd(v.Len())
}

func (enc *Encoder) encodeNil() {
	enc.WriteString("nil")
}

func (enc *Encoder) encodeComplex(v reflect.Value, bitSize int) {
	c := v.Complex()
	r, i := real(c), imag(c)
	enc.WriteString(strconv.FormatFloat(r, 'g', -1, bitSize))
	if i >= 0 {
		enc.WriteByte('+')
	}
	enc.WriteString(strconv.FormatFloat(i, 'g', -1, bitSize))
	enc.WriteByte('i')
}

func (enc *Encoder) encodeString(v reflect.Value) {
	enc.WriteString(
		strconv.Quote(v.String()),
	)
}

func (enc *Encoder) encodeRef(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	enc.encode(v.Elem())
}

func (enc *Encoder) encodeSlice(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
	} else {
		enc.encodeArray(v)
	}
}

func (enc *Encoder) encodeArray(v reflect.Value) {
	enc.listStart(v.Len())
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			enc.listSep()
		}
		enc.encode(v.Index(i))
	}
	enc.listEnd(v.Len())
}

func (enc *Encoder) EncodeValue(value reflect.Value) error {
	return nil
}

// stringValues is a slice of reflect.Value holding *reflect.StringValue.
// It implements the methods to sort by string.
type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }
