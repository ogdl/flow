// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"io"
)

const minBufSize = 512

type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenComment
	tokenString
	tokenLeftBrace
	tokenRightBrace
	tokenComma
)

func (t tokenType) String() string {
	switch t {
	case tokenError:
		return "tokenError"
	case tokenEOF:
		return "tokenEOF"
	case tokenComment:
		return "tokenComment"
	case tokenString:
		return "tokenString"
	case tokenLeftBrace:
		return "tokenLeftBrace"
	case tokenRightBrace:
		return "tokenRightBrace"
	case tokenComma:
		return "tokenComma"
	}
	return "token unkown"
}

type token struct {
	typ tokenType
	pos int
	val []byte
}

type scanner struct {
	r     io.Reader
	token *token
	err   error
	*fsm
}

func newScanner(r io.Reader) *scanner {
	return &scanner{
		r: r,
	}
}

func (s *scanner) scan() bool {
	if s.fsm == nil {
		s.fsm = newFSM(scanner_start)
	}
	s.token, s.err = s.next(s.r)
	return s.err == nil
}
