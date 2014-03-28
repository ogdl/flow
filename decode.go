// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
	"reflect"
)

type Decoder struct {
	*parser
	refSetter
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		newParser(r),
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
	if err := dec.ParseAny(rv); err != nil {
		return err
	}
	dec.setAllRef()
	return nil
}

func (dec *Decoder) ParseAny(v reflect.Value) (err error) {
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
	for _, match := range matchFuncs {
		if encoding, ok := match(v); ok && encoding.Decode != nil {
			return encoding.Decode(dec)
		}
	}
	return fmt.Errorf("no decoding method defined for type: %v", v.Type())
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
