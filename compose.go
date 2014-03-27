// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"bytes"
	"io"
	"reflect"
)

type Composer interface {
	io.Writer
	ComposeList(length int, composeElem func(i int) error) error
	ComposeAny(v reflect.Value) error
	Indented() bool
}

type composer struct {
	bytes.Buffer
	indented bool
	prefix   string
	indent   string
	depth    int
}

func (t *composer) Indented() bool {
	return t.indented
}

func (t *composer) ComposeList(length int, composeElem func(i int) error) error {
	t.listStart(length)
	for i := 0; i < length; i++ {
		if i > 0 {
			t.listSep()
		}
		if err := composeElem(i); err != nil {
			return err
		}
	}
	t.listEnd(length)
	return nil
}

func (t *composer) start(prefix, indent string) {
	t.indented = true
	t.prefix = prefix
	t.indent = indent
	t.depth = 0
	t.WriteString(prefix)
}

func (t *composer) stop() {
	t.indented = false
}

func (t *composer) newLine() {
	t.WriteString("\n")
	t.WriteString(t.prefix)
	for i := 0; i < t.depth; i++ {
		t.WriteString(t.indent)
	}
}

func (t *composer) listStart(count int) {
	t.WriteString("{")
	if t.indented {
		if count > 0 {
			t.depth++
			t.newLine()
		}
	}
}

func (t *composer) listSep() {
	if t.indented {
		t.WriteString(",")
		t.newLine()
	} else {
		t.WriteString(", ")
	}
}

func (t *composer) listEnd(count int) {
	if t.indented {
		if count > 0 {
			t.depth--
			t.WriteString(",")
			t.newLine()
		}
	}
	t.WriteString("}")
}

func (t *composer) encodeNil() {
	t.WriteString("nil")
}
