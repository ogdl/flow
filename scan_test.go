// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hailiang/gspec/core"
	"github.com/hailiang/gspec/errors"
	exp "github.com/hailiang/gspec/expectation"
	"github.com/hailiang/gspec/suite"
)

type testCase struct {
	text   string
	tokens []*token
}

var testCases = []testCase{
	{
		"a",
		[]*token{
			{tokenString, 0, []byte("a")},
		},
	},
	{
		"a b",
		[]*token{
			{tokenString, 0, []byte("a")},
			{tokenString, 2, []byte("b")},
		},
	},
	{
		"{}",
		[]*token{
			{tokenLeftBrace, 0, []byte("{")},
			{tokenRightBrace, 1, []byte("}")},
		},
	},
	{
		"{a}",
		[]*token{
			{tokenLeftBrace, 0, []byte("{")},
			{tokenString, 1, []byte("a")},
			{tokenRightBrace, 2, []byte("}")},
		},
	},
	{
		"{a, b}",
		[]*token{
			{tokenLeftBrace, 0, []byte("{")},
			{tokenString, 1, []byte("a")},
			{tokenComma, 2, []byte(",")},
			{tokenString, 4, []byte("b")},
			{tokenRightBrace, 5, []byte("}")},
		},
	},
	{
		"{a,}",
		[]*token{
			{tokenLeftBrace, 0, []byte("{")},
			{tokenString, 1, []byte("a")},
			{tokenComma, 2, []byte(",")},
			{tokenRightBrace, 3, []byte("}")},
		},
	},
	{
		"a{b}",
		[]*token{
			{tokenString, 0, []byte("a")},
			{tokenLeftBrace, 1, []byte("{")},
			{tokenString, 2, []byte("b")},
			{tokenRightBrace, 3, []byte("}")},
		},
	},
	{
		`"a"`,
		[]*token{
			{tokenString, 0, []byte(`"a"`)},
		},
	},
	{
		"a:",
		[]*token{
			{tokenString, 0, []byte("a")},
			{tokenString, 1, []byte(":")},
		},
	},
	{
		`"a":`,
		[]*token{
			{tokenString, 0, []byte(`"a"`)},
			{tokenString, 3, []byte(`:`)},
		},
	},
	{
		`{a}:`,
		[]*token{
			{tokenLeftBrace, 0, []byte("{")},
			{tokenString, 1, []byte("a")},
			{tokenRightBrace, 2, []byte("}")},
			{tokenString, 3, []byte(":")},
		},
	},
	{
		"/usr/bin",
		[]*token{
			{tokenString, 0, []byte("/usr/bin")},
		},
	},
	{
		"a //0123",
		[]*token{
			{tokenString, 0, []byte("a")},
			{tokenComment, 2, []byte("//0123")},
		},
	},
}

var _ = suite.Add(func(s core.S) {
	describe, testcase := s.Alias("describe"), s.Alias("testcase")
	expect := exp.Alias(s.Fail)

	describe("flow.scanner", func() {
		for _, tc := range testCases {
			testcase(fmt.Sprint(tc), func() {
				s := newScanner(strings.NewReader(tc.text))
				tokens := s.scanAll()
				expect(tokens).Equal(tc.tokens)
			})
		}
	})
})

func (s *scanner) scanAll() (tokens []*token) {
	for s.scan() {
		tokens = append(tokens, s.token)
	}
	return
}

func TestAll(t *testing.T) {
	suite.Run(t, false)
}

func p(v ...interface{}) {
	fmt.Println(v...)
}

func init() {
	errors.Sprint = flowPrint
	//	errors.Sprint = jsonPrint
	//errors.Sprint = dumpPrint
}

func flowPrint(v interface{}) string {
	buf, _ := MarshalIndent(v, "    ", "    ")
	return "\n" + string(buf) + "\n"
}

func dumpPrint(v interface{}) string {
	spew.Config.Indent = "    "
	return "\n" + spew.Sdump(v)
}

func jsonPrint(v interface{}) string {
	buf, _ := json.MarshalIndent(v, "    ", "  ")
	return "\n    " + string(buf) + "\n"
}
