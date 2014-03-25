// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
)

type ValueEncoding struct {
	Encode func(v reflect.Value, w writer) error
	Decode func(val []byte, v reflect.Value) error
}

var typeToValueEncoding = map[reflect.Type]ValueEncoding{
	reflect.TypeOf(bool(false)):   ValueEncoding{encodeBool, decodeBool},
	reflect.TypeOf(int8(0)):       ValueEncoding{encodeInt, decodeInt},
	reflect.TypeOf(int16(0)):      ValueEncoding{encodeInt, decodeInt},
	reflect.TypeOf(int32(0)):      ValueEncoding{encodeInt, decodeInt},
	reflect.TypeOf(int64(0)):      ValueEncoding{encodeInt, decodeInt},
	reflect.TypeOf(int(0)):        ValueEncoding{encodeInt, decodeInt},
	reflect.TypeOf(uint8(0)):      ValueEncoding{encodeUint, decodeUint},
	reflect.TypeOf(uint16(0)):     ValueEncoding{encodeUint, decodeUint},
	reflect.TypeOf(uint32(0)):     ValueEncoding{encodeUint, decodeUint},
	reflect.TypeOf(uint64(0)):     ValueEncoding{encodeUint, decodeUint},
	reflect.TypeOf(uint(0)):       ValueEncoding{encodeUint, decodeUint},
	reflect.TypeOf(uintptr(0)):    ValueEncoding{encodeUint, decodeUint},
	reflect.TypeOf(float32(0)):    ValueEncoding{encodeFloat32, decodeFloat32},
	reflect.TypeOf(float64(0)):    ValueEncoding{encodeFloat64, decodeFloat64},
	reflect.TypeOf(complex64(0)):  ValueEncoding{encodeComplex64, decodeComplex},
	reflect.TypeOf(complex128(0)): ValueEncoding{encodeComplex128, decodeComplex},
	reflect.TypeOf(string("")):    ValueEncoding{encodeString, decodeString},
}

func encodeBool(v reflect.Value, w writer) error {
	return w.WriteString(strconv.FormatBool(v.Bool()))
}

func decodeBool(val []byte, v reflect.Value) error {
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

func encodeInt(v reflect.Value, w writer) error {
	return w.WriteString(strconv.FormatInt(v.Int(), 10))
}

func decodeInt(val []byte, v reflect.Value) error {
	i := big.NewInt(0)
	i, ok := i.SetString(string(val), 10)
	if !ok {
		return fmt.Errorf("unexpected int value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetInt(i.Int64())
	return nil
}

func encodeUint(v reflect.Value, w writer) error {
	return w.WriteString(strconv.FormatUint(v.Uint(), 10))
}

func decodeUint(val []byte, v reflect.Value) error {
	i := big.NewInt(0)
	i, ok := i.SetString(string(val), 10)
	if !ok {
		return fmt.Errorf("unexpected int value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetUint(i.Uint64())
	return nil
}

func encodeFloat32(v reflect.Value, w writer) error {
	return encodeFloat(v, w, 32)
}

func decodeFloat32(val []byte, v reflect.Value) error {
	return decodeFloat(val, v, 32)
}

func encodeFloat64(v reflect.Value, w writer) error {
	return encodeFloat(v, w, 32)
}

func decodeFloat64(val []byte, v reflect.Value) error {
	return decodeFloat(val, v, 64)
}

func encodeFloat(v reflect.Value, w writer, bit int) error {
	return w.WriteString(strconv.FormatFloat(v.Float(), 'g', -1, bit))
}

func decodeFloat(val []byte, v reflect.Value, bit int) error {
	f, err := strconv.ParseFloat(string(val), bit)
	if err != nil {
		return fmt.Errorf("unexpected float value: %s", strconv.Quote(string(val)))
	}
	// TODO: handle overflow
	v.SetFloat(f)
	return nil
}

func encodeComplex64(v reflect.Value, w writer) error {
	return encodeComplex(v, w, 32)
}

func encodeComplex128(v reflect.Value, w writer) error {
	return encodeComplex(v, w, 64)
}

func encodeComplex(v reflect.Value, w writer, bitSize int) error {
	c := v.Complex()
	r, i := real(c), imag(c)
	if err := w.WriteString(strconv.FormatFloat(r, 'g', -1, bitSize)); err != nil {
		return err
	}
	if i >= 0 {
		if err := w.WriteByte('+'); err != nil {
			return err
		}
	}
	if err := w.WriteString(strconv.FormatFloat(i, 'g', -1, bitSize)); err != nil {
		return err
	}
	if err := w.WriteByte('i'); err != nil {
		return err
	}
	return nil
}

func decodeComplex(val []byte, v reflect.Value) error {
	var c complex128
	if _, err := fmt.Sscan(string(val), &c); err != nil {
		return err
	}
	// TODO: handle overflow
	v.SetComplex(c)
	return nil
}

func encodeString(v reflect.Value, w writer) error {
	return w.WriteString(strconv.Quote(v.String()))
}

func decodeString(val []byte, v reflect.Value) error {
	s, err := strconv.Unquote(string(val))
	if err != nil {
		s = string(val)
	}
	v.SetString(s)
	return nil
}
