// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"testing"

	"github.com/hailiang/gspec/errors"
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
		"`a`",
		[]*token{
			{tokenString, 0, []byte("`a`")},
		},
	},
	{
		"a:",
		[]*token{
			{tokenString, 0, []byte("a:")},
		},
	},
	{
		"`a`:",
		[]*token{
			{tokenString, 0, []byte("`a`:")},
		},
	},
	{
		`"a":`,
		[]*token{
			{tokenString, 0, []byte(`"a":`)},
		},
	},
	{
		`{a}:`,
		[]*token{
			{tokenLeftBrace, 0, []byte("{")},
			{tokenString, 1, []byte("a")},
			{tokenRightBrace, 2, []byte("}:")},
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

func TestAll(t *testing.T) {
	suite.Run(t, false)
}

func p(v ...interface{}) {
	fmt.Println(v...)
}

func init() {
	errors.Sprint = dumpPrint
}

func dumpPrint(v interface{}) string {
	spew.Config.Indent = "    "
	return "\n" + spew.Sdump(v)
}
