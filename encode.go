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
	"strings"
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

type writer interface {
	io.Writer
	WriteByte(c byte) error
	WriteString(s string) (n int, err error)
}

type bytesBuffer struct {
	bytes.Buffer
}

type Encoder struct {
	w io.Writer
	bytesBuffer
	refDetector
	indenter
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:           w,
		refDetector: newRefDetector()}
}

func (enc *Encoder) marshal(v interface{}) error {
	enc.populate(reflect.ValueOf(v))
	if err := enc.encode(reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) marshalIndent(v interface{}, prefix, indent string) error {
	enc.indenter.start(enc, prefix, indent)
	defer func() {
		enc.indenter.stop()
	}()
	return enc.marshal(v)
}

func (enc *Encoder) Encode(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	enc.populate(rv)
	if enc.serial > 1 {
		if rv.Kind() != reflect.Ptr && !rv.CanAddr() {
			return fmt.Errorf("object with cyclic reference must be addressable, %v", v)
		}
	}
	if err := enc.encode(rv); err != nil {
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
	if v.CanAddr() {
		addr := v.Addr().Pointer()
		id := enc.getPtrID(addr)
		if id > 0 && !enc.m[addr].defined {
			enc.define(addr)
			enc.WriteString(fmt.Sprintf("^%d ", id))
		}
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
	case reflect.Ptr:
		enc.encodePtr(v)
	case reflect.Interface:
		enc.encodeInterface(v)
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

func (enc *Encoder) encodeStruct(v reflect.Value) {
	t := v.Type()
	enc.listStart(enc, t.NumField())
	fieldNameMax := 0
	for i := 0; i < t.NumField(); i++ {
		l := len(t.Field(i).Name)
		if l > fieldNameMax {
			fieldNameMax = l
		}
	}
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			enc.listSep(enc)
		}
		fieldName := t.Field(i).Name
		enc.WriteString(fieldName)
		enc.WriteString(": " + strings.Repeat(" ", fieldNameMax-len(fieldName)))
		enc.encode(v.Field(i))
	}
	enc.listEnd(enc, t.NumField())
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
	enc.listStart(enc, v.Len())
	var sv stringValues = v.MapKeys()
	sort.Sort(sv)
	for i, k := range sv {
		if i > 0 {
			enc.listSep(enc)
		}
		enc.encodeKey(k)
		enc.WriteString(": ")
		enc.encode(v.MapIndex(k))
	}
	enc.listEnd(enc, v.Len())
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

func (enc *Encoder) encodePtr(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	addr := v.Pointer()
	id := enc.getPtrID(addr)
	if id > 0 {
		if enc.m[addr].defined {
			enc.WriteString(fmt.Sprintf("^%d", id))
		} else {
			enc.define(addr)
			enc.WriteString(fmt.Sprintf("^%d ", id))
			enc.encode(v.Elem())
		}
	} else {
		enc.encode(v.Elem())
	}
}

func (enc *Encoder) encodeInterface(v reflect.Value) {
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
	enc.listStart(enc, v.Len())
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			enc.listSep(enc)
		}
		enc.encode(v.Index(i))
	}
	enc.listEnd(enc, v.Len())
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

type refInfo struct {
	id      int
	defined bool
}

type refDetector struct {
	m      map[uintptr]refInfo
	serial int
}

func newRefDetector() refDetector {
	return refDetector{make(map[uintptr]refInfo), 1}
}

func (d *refDetector) getPtrID(addr uintptr) int {
	if ref := d.m[addr]; ref.id > 0 {
		return ref.id
	}
	return 0
}

func (d *refDetector) define(addr uintptr) {
	ref := d.m[addr]
	ref.defined = true
	d.m[addr] = ref
}

func (d *refDetector) add(addr uintptr) {
	ref := d.m[addr]
	switch ref.id {
	case 0:
		ref.id = -1
	case -1:
		ref.id = d.serial
		d.serial++
	}
	d.m[addr] = ref
}

func (d *refDetector) populate(v reflect.Value) {
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		addr := v.Addr().Pointer()
		d.add(addr)
		if d.m[addr].id > 0 {
			return
		}
	}
	switch v.Kind() {
	case reflect.Ptr:
		d.populate(v.Elem())
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			d.populate(v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.Type().NumField(); i++ {
			d.populate(v.Field(i))
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			d.populate(v.MapIndex(k))
		}
	}
}
