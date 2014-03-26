// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
)

type parser struct {
	*scanner
}

func newParser(r io.Reader) *parser {
	t := &parser{scanner: newScanner(r)}
	return t
}

func (t *parser) firstChild() *parser {
	return nil
}

func (t *parser) next() error {
	for t.scan() {
		if t.token.typ == tokenComment {
			continue
		} else {
			break
		}
	}
	return t.err
}

func (t *parser) isRef() bool {
	return t.isValue() && len(t.token.val) > 0 && t.token.val[0] == '^'
}

func (t *parser) isListStart() bool {
	return t.token.typ == tokenLeftBrace
}

func (t *parser) isNil() bool {
	val, err := t.getValue()
	if err != nil {
		return false
	}
	return string(val) == "nil"
}

func (t *parser) isListEnd() bool {
	return t.token.typ == tokenRightBrace
}

func (t *parser) isSepOrListEnd() bool {
	return t.isSep() || t.isListEnd()
}

func (t *parser) isSep() bool {
	return t.token.typ == tokenComma
}

func (t *parser) isValue() bool {
	return t.token.typ == tokenString
}

func (t *parser) getValue() ([]byte, error) {
	if !t.isValue() {
		return nil, t.error()
	}
	return t.token.val, nil
}

func (t *parser) error() error {
	return fmt.Errorf("unexpected token: %v, %s", t.token.typ,
		string(t.token.val))
}
