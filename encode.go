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

type bytesBuffer struct {
	bytes.Buffer
}

type Encoder struct {
	w io.Writer
	bytesBuffer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (enc *Encoder) Encode(v interface{}) error {
	if err := enc.encode(reflect.ValueOf(v)); err != nil {
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
	case reflect.Invalid, reflect.Chan, reflect.Func,
		reflect.UnsafePointer:
		return fmt.Errorf("unsupported variable type: %s", v.Type().String())
	default:
		enc.WriteString(fmt.Sprint(v.Interface()))
	}
	return nil
}

func (enc *Encoder) encodeStruct(v reflect.Value) {
	enc.WriteByte('{')
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			enc.WriteString(", ")
		}
		enc.WriteString(t.Field(i).Name)
		enc.WriteString(": ")
		enc.encode(v.Field(i))
	}
	enc.WriteByte('}')
}

func (enc *Encoder) encodeMap(v reflect.Value) {
	if v.IsNil() {
		enc.encodeNil()
		return
	}
	enc.WriteByte('{')
	var sv stringValues = v.MapKeys()
	sort.Sort(sv)
	for i, k := range sv {
		if i > 0 {
			enc.WriteString(", ")
		}
		enc.encode(k)
		enc.WriteString(": ")
		enc.encode(v.MapIndex(k))
	}
	enc.WriteByte('}')
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
	enc.WriteByte('{')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			enc.WriteString(", ")
		}
		enc.encode(v.Index(i))
	}
	enc.WriteByte('}')
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
