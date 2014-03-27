// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"reflect"
)

type MatchFunc func(v reflect.Value) (*Encoding, bool)

type Encoding struct {
	Encode func(v reflect.Value, c Composer) error
	Decode func(parser Parser, v reflect.Value) error
}

var matchFuncs = []MatchFunc{
	matchStruct,
	matchSlice,
	matchMap,
	matchArray,
}

func matchArray(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Array {
		return nil, false
	}
	return &Encoding{encodeArray, decodeArray}, true
}

func encodeArray(v reflect.Value, c Composer) error {
	return c.ComposeList(v.Len(), func(i int) error {
		return c.ComposeAny(v.Index(i))
	})
}

func decodeArray(parser Parser, v reflect.Value) error {
	return parser.ParseList(func(i int) error {
		elem := reflect.Value{}
		if i < v.Len() {
			elem = v.Index(i)
		}
		if err := parser.ParseAny(elem); err != nil {
			return err
		}
		return nil
	})
}

func matchSlice(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Slice {
		return nil, false
	}
	return &Encoding{encodeSlice, decodeSlice}, true
}

func encodeSlice(v reflect.Value, c Composer) error {
	if v.IsNil() {
		return composeNil(c)
	}
	return encodeArray(v, c)
}

func decodeSlice(parser Parser, v reflect.Value) error {
	if isNil(parser) {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	return parser.ParseList(func(i int) error {
		if i == v.Len() {
			v.Set(reflect.Append(v, reflect.New(v.Type().Elem()).Elem()))
		}
		elem := v.Index(i)
		if err := parser.ParseAny(elem); err != nil {
			return err
		}
		return nil
	})
}

func matchStruct(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	return &Encoding{nil, decodeStruct}, true
}

func decodeStruct(parser Parser, v reflect.Value) error {
	return parser.ParseList(func(int) error {
		var fieldName string
		if err := parser.ParseAny(reflect.ValueOf(&fieldName)); err != nil {
			return err
		}
		elem := reflect.Value{}
		if field := v.FieldByName(string(fieldName)); field.CanSet() {
			elem = field
		}
		if err := parser.GoToOnlyChild(); err != nil {
			return err
		}
		if err := skipColon(parser); err != nil {
			return err
		}
		if err := parser.ParseAny(elem); err != nil {
			return err
		}
		return nil
	})
}

func matchMap(v reflect.Value) (*Encoding, bool) {
	if v.Kind() != reflect.Map {
		return nil, false
	}
	return &Encoding{nil, decodeMap}, true
}

func decodeMap(parser Parser, v reflect.Value) error {
	if isNil(parser) {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	return parser.ParseList(func(int) error {
		key := reflect.New(v.Type().Key()).Elem()
		if err := parser.ParseAny(key); err != nil {
			return err
		}
		elem := reflect.New(v.Type().Elem()).Elem()
		if err := parser.GoToOnlyChild(); err != nil {
			return err
		}
		if err := skipColon(parser); err != nil {
			return err
		}
		if err := parser.ParseAny(elem); err != nil {
			return err
		}
		v.SetMapIndex(key, elem)
		return nil
	})
}

func composeNil(c Composer) error {
	return c.ComposeValue("nil")
}

func isNil(parser Parser) bool {
	if val, ok := parser.Value(); ok {
		return string(val) == "nil"
	}
	return false
}

func skipColon(parser Parser) error {
	if val, ok := parser.Value(); ok && string(val) == ":" {
		return parser.GoToOnlyChild()
	}
	return nil
}
