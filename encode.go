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
	WriteString(s string) error
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

func (enc *Encoder) WriteString(s string) error {
	_, err := enc.bytesBuffer.WriteString(s)
	return err
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
	if encoding, ok := typeToValueEncoding[v.Type()]; ok {
		if encoding.Encode != nil {
			return encoding.Encode(v, enc)
		}
	} else {
		// TODO: unexpected type
	}
	switch v.Kind() {
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

func (enc *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()
	fieldNameMax := 0
	for i := 0; i < t.NumField(); i++ {
		l := len(t.Field(i).Name)
		if l > fieldNameMax {
			fieldNameMax = l
		}
	}
	return enc.encodeList(t.NumField(), func(i int) error {
		fieldName := t.Field(i).Name
		enc.WriteString(fieldName)
		enc.WriteString(": " + strings.Repeat(" ", fieldNameMax-len(fieldName)))
		return enc.encode(v.Field(i))
	})
}

func (enc *Encoder) disableIndent(f func()) {
	m := enc.indentMode
	enc.indentMode = false
	defer func() {
		enc.indentMode = m
	}()
	f()
}

func (enc *Encoder) encodeKey(v reflect.Value) {
	m := enc.indentMode
	enc.indentMode = false
	defer func() {
		enc.indentMode = m
	}()

	enc.encode(v)
}

func (enc *Encoder) encodeMap(v reflect.Value) error {
	if v.IsNil() {
		enc.encodeNil()
		return nil
	}
	var keys stringValues = v.MapKeys()
	sort.Sort(keys)
	return enc.encodeList(v.Len(), func(i int) error {
		key := keys[i]
		enc.disableIndent(func() {
			enc.encode(key)
		})
		enc.WriteString(": ")
		return enc.encode(v.MapIndex(key))
	})
}

func (enc *Encoder) encodeSlice(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
	} else {
		enc.encodeArray(v)
	}
}

func (enc *Encoder) encodeArray(v reflect.Value) error {
	return enc.encodeList(v.Len(), func(i int) error {
		return enc.encode(v.Index(i))
	})
}

func (enc *Encoder) encodeList(length int, encodeElem func(i int) error) error {
	enc.listStart(enc, length)
	for i := 0; i < length; i++ {
		if i > 0 {
			enc.listSep(enc)
		}
		if err := encodeElem(i); err != nil {
			return err
		}
	}
	enc.listEnd(enc, length)
	return nil
}

func (enc *Encoder) encodeNil() {
	enc.WriteString("nil")
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
