// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"strings"
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
	}
	switch v.Kind() {
	case reflect.Slice:
		return dec.decodeSlice(v)
	case reflect.Array:
		return dec.decodeArray(v)
	case reflect.Struct:
		return dec.decodeStruct(v)
	case reflect.Map:
		return dec.decodeMap(v)
	}

	// values
	tokenVal, err := dec.expectValue()
	if err != nil {
		return err
	}
	switch v.Kind() {
	case reflect.Bool:
		return dec.decodeBool(v, tokenVal)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return dec.decodeInt(v, tokenVal)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		return dec.decodeUint(v, tokenVal)
	case reflect.Float32:
		return dec.decodeFloat(v, tokenVal, 32)
	case reflect.Float64:
		return dec.decodeFloat(v, tokenVal, 64)
	case reflect.Complex64, reflect.Complex128:
		return dec.decodeComplex(v, tokenVal)
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

func (dec *Decoder) decodeFloat(v reflect.Value, val []byte, bit int) error {
	f, err := strconv.ParseFloat(string(val), bit)
	if err != nil {
		return fmt.Errorf("unexpected float value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetFloat(f)
	return nil
}

func (dec *Decoder) decodeComplex(v reflect.Value, val []byte) error {
	var c complex128
	if _, err := fmt.Sscan(string(val), &c); err != nil {
		return err
	}
	// TODO: handle overflow
	v.SetComplex(c)
	return nil
}

func (dec *Decoder) decodeString(v reflect.Value, val []byte) error {
	s, err := strconv.Unquote(string(val))
	if err != nil {
		s = string(val)
	}
	v.SetString(s)
	return nil
}

func (dec *Decoder) decodeSlice(v reflect.Value) error {
	if err := dec.expectNil(); err == nil {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	i := -1
	return dec.decodeList(v, func() error {
		i++
		if i == v.Len() {
			v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
		}
		elem := v.Index(i)
		if err := dec.decode(elem); err != nil {
			return err
		}
		return nil
	})
}

func (dec *Decoder) decodeArray(v reflect.Value) error {
	i := -1
	return dec.decodeList(v, func() error {
		i++
		elem := reflect.Value{}
		if i < v.Len() {
			elem = v.Index(i)
		}
		if err := dec.decode(elem); err != nil {
			return err
		}
		return nil
	})
}

func (dec *Decoder) decodeStruct(v reflect.Value) error {
	return dec.decodeList(v, func() error {
		fieldName, err := dec.expectFieldName()
		if err != nil {
			return err
		}
		elem := reflect.Value{}
		if field := v.FieldByName(fieldName); field.CanSet() {
			elem = field
		}
		if err := dec.next(); err != nil {
			return err
		}
		if err := dec.skipColon(); err != nil {
			return err
		}
		if err := dec.decode(elem); err != nil {
			return err
		}
		return nil
	})
}

func (dec *Decoder) decodeMap(v reflect.Value) error {
	if err := dec.expectNil(); err == nil {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	return dec.decodeList(v, func() error {
		key := reflect.New(v.Type().Key()).Elem()
		if err := dec.decode(key); err != nil {
			return err
		}
		elem := reflect.New(v.Type().Elem()).Elem()
		if err := dec.next(); err != nil {
			return err
		}
		if err := dec.skipColon(); err != nil {
			return err
		}
		if err := dec.decode(elem); err != nil {
			return err
		}
		v.SetMapIndex(key, elem)
		return nil
	})
}

func (dec *Decoder) decodeList(sv reflect.Value, decodeElem func() error) error {
	if err := dec.expectListStart(); err != nil {
		return err
	}
	if err := dec.next(); err != nil {
		return err
	}
	for {
		isElem, err := dec.expectListEndOrElem()
		if err != nil {
			return err
		}
		if !isElem {
			break
		}
		if err := decodeElem(); err != nil {
			return err
		}
		if err := dec.next(); err != nil {
			return err
		}
		if err := dec.expectElemSepOrListEnd(); err != nil {
			return err
		}
	}
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

func (dec *Decoder) skipColon() error {
	if dec.token.typ == tokenString && string(dec.token.val) == ":" {
		return dec.next()
	}
	return nil
}

func (dec *Decoder) expectFieldName() (string, error) {
	val, err := dec.expectValue()
	if err != nil {
		return "", err
	}

	if len(val) == 0 {
		return "", dec.error()
	}

	return strings.Title(string(val)), nil
}

func (dec *Decoder) expectValue() ([]byte, error) {
	if dec.token.typ == tokenString {
		return dec.token.val, nil
	}
	return nil, dec.error()
}

func (dec *Decoder) expectListStart() error {
	if dec.token.typ == tokenLeftBrace {
		return nil
	}
	return dec.error()
}

func (dec *Decoder) expectNil() error {
	if dec.token.typ == tokenString && string(dec.token.val) == "nil" {
		return nil
	}
	return dec.error()
}

func (dec *Decoder) expectNilOrListStart() (isNil bool, _ error) {
	if err := dec.expectNil(); err == nil {
		return true, nil
	}
	return false, dec.expectListStart()
}

func (dec *Decoder) expectListEndOrElem() (isElem bool, _ error) {
	if dec.token.typ == tokenRightBrace {
		return false, nil
	} else {
		return true, nil
	}
	return false, dec.error()
}

func (dec *Decoder) expectElemSepOrListEnd() error {
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
