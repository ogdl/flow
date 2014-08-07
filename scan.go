// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"io"

	"github.com/hailiang/gombi/scan"
)

type scanner struct {
	*scan.Scanner
}

func (s *scanner) Scan() bool {
	for s.Scanner.Scan() {
		if s.Token().Type != tokenSpace {
			return true
		}
	}
	return false
}

func newScanner(r io.Reader) *scanner {
	var (
		char   = scan.Char
		pat    = scan.Pat
		merge  = scan.Merge
		or     = scan.Or
		con    = scan.Con
		Tokens = scan.Tokens

		nonctrl = char(`[:cntrl:]`).Negate()
		indent  = char(`\t `)
		lbreak  = char(`\n\r`)
		space   = merge(indent, lbreak)
		inline  = merge(nonctrl, indent)
		any     = merge(nonctrl, space)
		invalid = any.Negate()
		delim   = char(`,{}`)
		empty   = pat(``)

		newline        = or(lbreak, pat(`\r\n`))
		inlineComment  = con(pat(`//`), inline.ZeroOrMore(), or(newline, empty))
		quoted         = or(inline.Exclude(char(`"`)), pat(`\\"`))
		quotedString   = con(pat(`"`), quoted.ZeroOrMore(), pat(`"`))
		unquoted       = nonctrl.Exclude(merge(delim, char(` `)))
		unquotedString = unquoted.OneOrMore()
		generalString  = or(quotedString, unquotedString)

		tokens = Tokens(
			invalid,
			inlineComment,
			char(`{`),
			char(`}`),
			char(`,`),
			generalString,
			space.OneOrMore(),
		)
	)
	s, err := scan.NewScanner(tokens.String(), r)
	if err != nil {
		panic(err)
	}
	return &scanner{s}
}

const (
	tokenError = iota
	tokenComment
	tokenLeftBrace
	tokenRightBrace
	tokenComma
	tokenString
	tokenSpace
	tokenEOF
)

type tokenType int

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
