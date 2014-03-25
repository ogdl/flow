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
)

type Decoder struct {
	*scanner
	refSetter
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		newScanner(r),
		newRefSetter(),
	}
}

func (dec *Decoder) Decode(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	if err := dec.next(); err != nil {
		return err
	}
	if err := dec.decode(rv); err != nil {
		return err
	}
	dec.setAllRef()
	return nil
}

func (dec *Decoder) decode(v reflect.Value) (err error) {
	if !v.CanSet() && v.Kind() != reflect.Ptr { // TODO: interface should also be allowed.
		return fmt.Errorf("unsetable nonpointer value: %v", v)
	}
	if v.Kind() == reflect.Ptr {
		v = deref(v).Elem()
	}
	if dec.isRef() {
		id := string(dec.token.val[1:])
		if err := dec.next(); err != nil {
			return err
		}
		if dec.isSepOrListEnd() {
			dec.addDstRef(id, v)
			return nil
		} else {
			dec.addSrcRef(id, v)
		}
	}

	defer func() {
		if e := dec.next(); e != nil && e != io.EOF {
			err = e
		}
	}()
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
	if dec.isNil() {
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
	if dec.isNil() {
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
	if !dec.isListStart() {
		return dec.error()
	}
	if err := dec.next(); err != nil {
		return err
	}
	for {
		if dec.isListEnd() {
			break
		}
		if err := decodeElem(); err != nil {
			return err
		}
		if dec.isSep() {
			if err := dec.next(); err != nil {
				return err
			}
		} else if !dec.isListEnd() {
			return dec.error()
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

	return string(val), nil
}

func (dec *Decoder) expectValue() ([]byte, error) {
	if dec.token.typ == tokenString {
		return dec.token.val, nil
	}
	return nil, dec.error()
}

func (dec *Decoder) isRef() bool {
	return dec.token.typ == tokenString && len(dec.token.val) > 0 && dec.token.val[0] == '^'
}

func (dec *Decoder) isListStart() bool {
	return dec.token.typ == tokenLeftBrace
}

func (dec *Decoder) isNil() bool {
	return dec.token.typ == tokenString && string(dec.token.val) == "nil"
}

func (dec *Decoder) isListEnd() bool {
	return dec.token.typ == tokenRightBrace
}

func (dec *Decoder) isSepOrListEnd() bool {
	return dec.isSep() || dec.isListEnd()
}

func (dec *Decoder) isSep() bool {
	return dec.token.typ == tokenComma
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

type refValue struct {
	src reflect.Value
	dst []reflect.Value
}

type refSetter struct {
	m map[string]refValue
}

func newRefSetter() refSetter {
	return refSetter{make(map[string]refValue)}
}

func (s *refSetter) setAllRef() {
	for _, refVal := range s.m {
		for _, dst := range refVal.dst {
			dst.Set(refVal.src)
		}
	}
}

func (s *refSetter) addSrcRef(id string, v reflect.Value) {
	refVal := s.m[id]
	refVal.src = v
	s.m[id] = refVal
}

func (s *refSetter) addDstRef(id string, v reflect.Value) {
	refVal := s.m[id]
	refVal.dst = append(refVal.dst, v)
	s.m[id] = refVal
}
