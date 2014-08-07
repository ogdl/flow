
// line 1 "exec.rl"
// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
    "fmt"
)


// line 15 "../scan_auto.go"
const scanner_start int = 2
const scanner_error int = -1

const scanner_en_main int = 2


// line 19 "exec.rl"


func (m *fsm) exec() (tok *token, err error) {
    
// line 27 "../scan_auto.go"
	{
	if (   m.p) == (  len(m.data)) {
		goto _test_eof
	}
	switch    m.cs {
	case 2:
		goto st_case_2
	case 3:
		goto st_case_3
	case 4:
		goto st_case_4
	case 5:
		goto st_case_5
	case 0:
		goto st_case_0
	case 1:
		goto st_case_1
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 8:
		goto st_case_8
	case 9:
		goto st_case_9
	case 10:
		goto st_case_10
	case 11:
		goto st_case_11
	}
	goto st_out
tr0:
// line 24 "action.rl"

(   m.p) = (   m.te) - 1
{
        tok = m.newToken(tokenString)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr2:
// line 24 "action.rl"

   m.te = (   m.p)+1
{
        tok = m.newToken(tokenString)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr5:
// line 4 "action.rl"

   m.te = (   m.p)+1
{
        err = fmt.Errorf("Invalid character.")
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr9:
// line 19 "action.rl"

   m.te = (   m.p)+1
{
        tok = m.newToken(tokenComma)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr11:
// line 9 "action.rl"

   m.te = (   m.p)+1
{
        tok = m.newToken(tokenLeftBrace)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr12:
// line 14 "action.rl"

   m.te = (   m.p)+1
{
        tok = m.newToken(tokenRightBrace)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr13:
// line 38 "scanner.rl"

   m.te = (   m.p)
(   m.p)--

	goto st2
tr14:
// line 24 "action.rl"

   m.te = (   m.p)
(   m.p)--
{
        tok = m.newToken(tokenString)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr17:
// line 29 "action.rl"

   m.te = (   m.p)
(   m.p)--
{
        tok = m.newToken(tokenComment)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
tr19:
// line 29 "action.rl"

   m.te = (   m.p)+1
{
        tok = m.newToken(tokenComment)
        {(   m.p)++;    m.cs = 2; goto _out }
    }
	goto st2
	st2:
// line 1 "NONE"

   m.ts = 0

		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof2
		}
	st_case_2:
// line 1 "NONE"

   m.ts = (   m.p)

// line 162 "../scan_auto.go"
		switch    m.data[(   m.p)] {
		case 13:
			goto st3
		case 32:
			goto st3
		case 34:
			goto tr8
		case 44:
			goto tr9
		case 47:
			goto st8
		case 123:
			goto tr11
		case 125:
			goto tr12
		}
		switch {
		case    m.data[(   m.p)] < 9:
			if    m.data[(   m.p)] <= 8 {
				goto tr5
			}
		case    m.data[(   m.p)] > 10:
			if 11 <=    m.data[(   m.p)] &&    m.data[(   m.p)] <= 31 {
				goto tr5
			}
		default:
			goto st3
		}
		goto st4
	st3:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof3
		}
	st_case_3:
		switch    m.data[(   m.p)] {
		case 13:
			goto st3
		case 32:
			goto st3
		}
		if 9 <=    m.data[(   m.p)] &&    m.data[(   m.p)] <= 10 {
			goto st3
		}
		goto tr13
	st4:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof4
		}
	st_case_4:
		switch    m.data[(   m.p)] {
		case 44:
			goto tr14
		case 123:
			goto tr14
		case 125:
			goto tr14
		}
		if    m.data[(   m.p)] <= 32 {
			goto tr14
		}
		goto st4
tr8:
// line 1 "NONE"

   m.te = (   m.p)+1

	goto st5
	st5:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof5
		}
	st_case_5:
// line 235 "../scan_auto.go"
		switch    m.data[(   m.p)] {
		case 9:
			goto st0
		case 32:
			goto st0
		case 34:
			goto st4
		case 44:
			goto st0
		case 92:
			goto tr15
		case 123:
			goto st0
		case 125:
			goto st0
		}
		if    m.data[(   m.p)] <= 31 {
			goto tr14
		}
		goto tr8
	st0:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof0
		}
	st_case_0:
		switch    m.data[(   m.p)] {
		case 34:
			goto tr2
		case 92:
			goto st1
		}
		switch {
		case    m.data[(   m.p)] > 8:
			if 10 <=    m.data[(   m.p)] &&    m.data[(   m.p)] <= 31 {
				goto tr0
			}
		default:
			goto tr0
		}
		goto st0
	st1:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof1
		}
	st_case_1:
		switch    m.data[(   m.p)] {
		case 34:
			goto tr4
		case 92:
			goto st1
		}
		switch {
		case    m.data[(   m.p)] > 8:
			if 10 <=    m.data[(   m.p)] &&    m.data[(   m.p)] <= 31 {
				goto tr0
			}
		default:
			goto tr0
		}
		goto st0
tr4:
// line 1 "NONE"

   m.te = (   m.p)+1

	goto st6
	st6:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof6
		}
	st_case_6:
// line 307 "../scan_auto.go"
		switch    m.data[(   m.p)] {
		case 34:
			goto tr2
		case 92:
			goto st1
		}
		switch {
		case    m.data[(   m.p)] > 8:
			if 10 <=    m.data[(   m.p)] &&    m.data[(   m.p)] <= 31 {
				goto tr14
			}
		default:
			goto tr14
		}
		goto st0
tr15:
// line 1 "NONE"

   m.te = (   m.p)+1

	goto st7
	st7:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof7
		}
	st_case_7:
// line 334 "../scan_auto.go"
		switch    m.data[(   m.p)] {
		case 9:
			goto st0
		case 32:
			goto st0
		case 44:
			goto st0
		case 92:
			goto tr15
		case 123:
			goto st0
		case 125:
			goto st0
		}
		if    m.data[(   m.p)] <= 31 {
			goto tr14
		}
		goto tr8
	st8:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof8
		}
	st_case_8:
		switch    m.data[(   m.p)] {
		case 44:
			goto tr14
		case 47:
			goto st9
		case 123:
			goto tr14
		case 125:
			goto tr14
		}
		if    m.data[(   m.p)] <= 32 {
			goto tr14
		}
		goto st4
	st9:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof9
		}
	st_case_9:
		switch    m.data[(   m.p)] {
		case 9:
			goto st10
		case 10:
			goto tr19
		case 13:
			goto st11
		case 32:
			goto st10
		case 44:
			goto st10
		case 123:
			goto st10
		case 125:
			goto st10
		}
		if    m.data[(   m.p)] <= 31 {
			goto tr17
		}
		goto st9
	st10:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof10
		}
	st_case_10:
		switch    m.data[(   m.p)] {
		case 10:
			goto tr19
		case 13:
			goto st11
		}
		switch {
		case    m.data[(   m.p)] > 8:
			if 11 <=    m.data[(   m.p)] &&    m.data[(   m.p)] <= 31 {
				goto tr17
			}
		default:
			goto tr17
		}
		goto st10
	st11:
		if (   m.p)++; (   m.p) == (  len(m.data)) {
			goto _test_eof11
		}
	st_case_11:
		if    m.data[(   m.p)] == 10 {
			goto tr19
		}
		goto tr17
	st_out:
	_test_eof2:    m.cs = 2; goto _test_eof
	_test_eof3:    m.cs = 3; goto _test_eof
	_test_eof4:    m.cs = 4; goto _test_eof
	_test_eof5:    m.cs = 5; goto _test_eof
	_test_eof0:    m.cs = 0; goto _test_eof
	_test_eof1:    m.cs = 1; goto _test_eof
	_test_eof6:    m.cs = 6; goto _test_eof
	_test_eof7:    m.cs = 7; goto _test_eof
	_test_eof8:    m.cs = 8; goto _test_eof
	_test_eof9:    m.cs = 9; goto _test_eof
	_test_eof10:    m.cs = 10; goto _test_eof
	_test_eof11:    m.cs = 11; goto _test_eof

	_test_eof: {}
	if (   m.p) == ( m.eof()) {
		switch    m.cs {
		case 3:
			goto tr13
		case 4:
			goto tr14
		case 5:
			goto tr14
		case 0:
			goto tr0
		case 1:
			goto tr0
		case 6:
			goto tr14
		case 7:
			goto tr14
		case 8:
			goto tr14
		case 9:
			goto tr17
		case 10:
			goto tr17
		case 11:
			goto tr17
		}
	}

	_out: {}
	}

// line 23 "exec.rl"
    if m.cs == scanner_error {
        return nil, fmt.Errorf("parse error: %s", string(m.data[m.te:]))
    }
    return
}
