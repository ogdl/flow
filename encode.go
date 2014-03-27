// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
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

type Encoder struct {
	w io.Writer
	refDetector
	composer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:           w,
		refDetector: newRefDetector()}
}

func (enc *Encoder) marshal(v interface{}) error {
	enc.populate(reflect.ValueOf(v))
	if err := enc.ComposeAny(reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) marshalIndent(v interface{}, prefix, indent string) error {
	enc.start(prefix, indent)
	defer func() {
		enc.stop()
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
	if err := enc.ComposeAny(rv); err != nil {
		return err
	}
	_, err := enc.w.Write(enc.Bytes())
	return err
}

// encode never returns an error, it may panics with bytes.ErrTooLarge.
func (enc *Encoder) ComposeAny(v reflect.Value) error {
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
	for _, match := range matchFuncs {
		if encoding, ok := match(v); ok && encoding.Encode != nil {
			return encoding.Encode(v, enc)
		}
	}
	switch v.Kind() {
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
	if enc.indentMode {
		for i := 0; i < t.NumField(); i++ {
			l := len(t.Field(i).Name)
			if l > fieldNameMax {
				fieldNameMax = l
			}
		}
	}
	return enc.ComposeList(t.NumField(), func(i int) error {
		fieldName := t.Field(i).Name
		enc.ComposeValue(fieldName)
		enc.ComposeValue(": ")
		if enc.indentMode {
			enc.ComposeValue(strings.Repeat(" ", fieldNameMax-len(fieldName)))
		}
		return enc.ComposeAny(v.Field(i))
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

func (enc *Encoder) encodeMap(v reflect.Value) error {
	if v.IsNil() {
		enc.encodeNil()
		return nil
	}
	var keys stringValues = v.MapKeys()
	sort.Sort(keys)
	/*
	keyMax := 0
	if enc.indentMode {
		for _, key := range keys {
			l := len(key.String())
			if l > keyMax {
				keyMax = l
			}
		}
	}
	*/
	return enc.ComposeList(v.Len(), func(i int) error {
		key := keys[i]
		var buf bytesBuffer
		en := NewEncoder(&buf)
		en.Encode(key)
		enc.ComposeValue(buf.String())
		enc.ComposeValue(": ")
		/*
		if enc.indentMode {
			enc.ComposeValue(strings.Repeat(" ", keyMax-len(key.String())))
		}
		*/
		return enc.ComposeAny(v.MapIndex(key))
	})
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
			enc.ComposeAny(v.Elem())
		}
	} else {
		enc.ComposeAny(v.Elem())
	}
}

func (enc *Encoder) encodeInterface(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	enc.ComposeAny(v.Elem())
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
