// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
)

type Decoder struct {
	*scanner
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		scanner: newScanner(r),
	}
}

func (dec *Decoder) Decode(e interface{}) error {
	if err := dec.next(); err != nil {
		return err
	}
	return dec.decode(reflect.ValueOf(e))
}

func (dec *Decoder) decode(v reflect.Value) error {
	if !v.CanSet() && v.Kind() != reflect.Ptr { // interface should also be allowed.
		return fmt.Errorf("unsetable nonpointer value: %v", v)
	}
	if v.Kind() == reflect.Ptr {
		v = deref(v).Elem()
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			return dec.decodeList(v)
		case reflect.Struct:
			return dec.decodeStruct(v)
		}
	}

	// values
	tokenVal, err := dec.expectValue()
	if err != nil {
		return err
	}
	switch v.Kind() {
	case reflect.Bool:
		return dec.decodeBool(v, tokenVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return dec.decodeInt(v, tokenVal)
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return dec.decodeUint(v, tokenVal)
	case reflect.Float32:
		return dec.decodeFloat32(v, tokenVal)
	case reflect.Float64:
		return dec.decodeFloat64(v, tokenVal)
	case reflect.Complex64:
		return dec.decodeComplex64(v, tokenVal)
	case reflect.Complex128:
		return dec.decodeComplex128(v, tokenVal)
	case reflect.String:
		return dec.decodeString(v, tokenVal)
	}
	return nil
}

func (dec *Decoder) decodeBool(v reflect.Value, val []byte) error {
	switch string(val) {
	case "true":
		v.SetBool(true)
	case "false":
		v.SetBool(false)
	default:
		return fmt.Errorf("unexpected bool value: %s", strconv.Quote(string(val)))
	}
	return nil
}

func (dec *Decoder) decodeInt(v reflect.Value, val []byte) error {
	i := big.NewInt(0)
	i, ok := i.SetString(string(val), 10)
	if !ok {
		return fmt.Errorf("unexpected int value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetInt(i.Int64())
	return nil
}

func (dec *Decoder) decodeUint(v reflect.Value, val []byte) error {
	i := big.NewInt(0)
	i, ok := i.SetString(string(val), 10)
	if !ok {
		return fmt.Errorf("unexpected int value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetUint(i.Uint64())
	return nil
}

func (dec *Decoder) decodeFloat32(v reflect.Value, val []byte) error {
	f, err := strconv.ParseFloat(string(val), 32)
	if err != nil {
		return fmt.Errorf("unexpected float value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetFloat(f)
	return nil
}

func (dec *Decoder) decodeFloat64(v reflect.Value, val []byte) error {
	f, err := strconv.ParseFloat(string(val), 64)
	if err != nil {
		return fmt.Errorf("unexpected float value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetFloat(f)
	return nil
}

var reComplex = regexp.MustCompile(`((?:\+|-)?.+)?((?:\+|-).+)i`)

func (dec *Decoder) decodeComplex64(v reflect.Value, val []byte) error {
	m := reComplex.FindSubmatch(val)
	if m == nil {
		return fmt.Errorf("unexpected complex value: %s", strconv.Quote(string(val)))
	}
	if len(m) > 2 {
		r, err := strconv.ParseFloat(string(m[1]), 32)
		if err != nil {
			return err
		}
		i, err := strconv.ParseFloat(string(m[2]), 32)
		if err != nil {
			return err
		}
		v.SetComplex(complex(r, i))
	}
	// TODO: handle error
	return nil
}

func (dec *Decoder) decodeComplex128(v reflect.Value, val []byte) error {
	m := reComplex.FindSubmatch(val)
	if m == nil {
		return fmt.Errorf("unexpected complex value: %s", strconv.Quote(string(val)))
	}
	if len(m) > 2 {
		r, err := strconv.ParseFloat(string(m[1]), 64)
		if err != nil {
			return err
		}
		i, err := strconv.ParseFloat(string(m[2]), 64)
		if err != nil {
			return err
		}
		v.SetComplex(complex(r, i))
	}
	// TODO: handle error
	return nil
}

func (dec *Decoder) decodeString(v reflect.Value, val []byte) error {
	s, err := strconv.Unquote(string(val))
	if err != nil {
		return err
	}
	v.SetString(s)
	return nil
}

func (dec *Decoder) decodeList(sv reflect.Value) error {
	isNil, err := dec.expectNilOrListStart()
	if err != nil {
		return err
	}
	if isNil {
		sv.Set(reflect.Zero(sv.Type()))
		return nil
	}
	if err := dec.next(); err != nil {
		return err
	}
	for i := 0; ; i++ {
		isElem, err := dec.expectListEndOrElem()
		if err != nil {
			return err
		}
		if !isElem {
			break
		}
		if sv.Kind() == reflect.Slice {
			sv.Set(reflect.Append(sv, reflect.New(sv.Type().Elem()).Elem()))
		}
		if err := dec.decode(sv.Index(i)); err != nil {
			return err
		}
		if err := dec.next(); err != nil {
			return err
		}
		if err := dec.expectElemSep(); err != nil {
			return err
		}
	}
	return nil
}

func (dec *Decoder) decodeStruct(sv reflect.Value) error {
	return nil
}

func (dec *Decoder) next() error {
	for dec.scan() {
		if dec.token.typ == tokenComment {
			continue
		} else {
			break
		}
	}
	return dec.err
}

func (dec *Decoder) expectValue() ([]byte, error) {
	if dec.token.typ == tokenString {
		return dec.token.val, nil
	}
	return nil, dec.error()
}

func (dec *Decoder) expectNilOrListStart() (isNil bool, _ error) {
	if dec.token.typ == tokenLeftBrace {
		return false, nil
	} else if dec.token.typ == tokenString && string(dec.token.val) == "nil" {
		return true, nil

	}
	return false, dec.error()
}

func (dec *Decoder) expectListEndOrElem() (isElem bool, _ error) {
	if dec.token.typ == tokenRightBrace {
		return false, nil
	} else {
		return true, nil
	}
	return false, dec.error()
}

func (dec *Decoder) expectElemSep() error {
	if dec.token.typ == tokenComma {
		if err := dec.next(); err != nil {
			return err
		}
		return nil
	} else if dec.token.typ == tokenRightBrace {
		return nil
	}
	return dec.error()
}

func (dec *Decoder) error() error {
	return fmt.Errorf("unexpected token: %v, %s", dec.token.typ,
		string(dec.token.val))
}

// A DecodeValueError happens when the decoded string and the type of provided value do not match.
type DecodeValueError struct {
	Value string
	Type  reflect.Type
}

func (e *DecodeValueError) Error() string {
	return "ogdl: cannot unmarshal " + strconv.Quote(e.Value) + " into Go value of type " + e.Type.String()
}

// TODO:
// Improve panic message of gspec!!!

func deref(v reflect.Value) reflect.Value {
	if isRef(v) {
		if v.IsNil() && v.Kind() == reflect.Ptr {
			v.Set(reflect.New(v.Type().Elem()))
		} // for interface, there is no way to alloc an object, so a register
		// method is needed
		if isRef(v.Elem()) {
			return deref(v.Elem())
		}
		return v
	}
	return v
}

func isRef(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return true
	}
	return false
}
