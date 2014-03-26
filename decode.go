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
	tokenVal, err := dec.getValue()
	if err != nil {
		return err
	}
	if encoding, ok := typeToValueEncoding[v.Type()]; ok {
		return encoding.Decode(tokenVal, v)
	} else {
		// TODO: unexpected type
	}
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
		fieldName, err := dec.getValue()
		if err != nil {
			return err
		}
		elem := reflect.Value{}
		if field := v.FieldByName(string(fieldName)); field.CanSet() {
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

func (dec *Decoder) skipColon() error {
	if dec.token.typ == tokenString && string(dec.token.val) == ":" {
		return dec.next()
	}
	return nil
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
