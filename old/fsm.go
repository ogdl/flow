// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"io"
)

// fsm is a wrapper for Ragel FSM with buffer management.
type fsm struct {
	data []byte
	cs   int
	p    int

	ts, te int
	act    int
	offset int
	err    error
}

func newFSM(initialState int) *fsm {
	return &fsm{
		data: make([]byte, 0, minBufSize),
		cs:   initialState,
	}
}

func (m *fsm) next(r io.Reader) (tok *token, _ error) {
	for {
		if m.done() {
			return nil, m.err
		}
		tok, m.err = m.exec()
		if m.err != nil {
			return nil, m.err
		} else if tok != nil {
			return tok, nil
		}
		m.err = m.read(r)
	}
}

func (m *fsm) done() bool {
	return m.err == io.EOF && m.te == len(m.data) || // ended successfully
		m.err != io.EOF && m.err != nil // ended with an error
}

func (m *fsm) eof() int {
	if m.err != io.EOF {
		return -1
	}
	return len(m.data)
}

func (m *fsm) newToken(typ tokenType) *token {
	if m.te <= m.ts {
		return nil
	}
	return &token{typ, m.offset + m.ts, append([]byte(nil), m.data[m.ts:m.te]...)}
}

func (m *fsm) read(r io.Reader) error {
	m.prepareBuffer()
	buf := m.data
	n, err := r.Read(buf[len(buf):cap(buf)])
	m.data = buf[:len(buf)+n]
	return err
}

func (m *fsm) prepareBuffer() {
	buf := m.data
	if m.ts > 0 {
		copy(buf, buf[m.ts:])
		buf = buf[:len(buf)-m.ts]
		m.p -= m.ts
		m.te -= m.ts
		m.offset += m.ts
		m.ts = 0
	}
	if len(buf) == cap(buf) {
		buf = append(buf, 0)[:len(buf)]
	}
	m.data = buf
}
