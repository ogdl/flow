// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

type indenter struct {
	indentMode bool
	prefix     string
	indent     string
	depth      int
}

func (t *indenter) start(w writer, prefix, indent string) {
	t.indentMode = true
	t.prefix = prefix
	t.indent = indent
	t.depth = 0
	w.WriteString(prefix)
}

func (t *indenter) stop() {
	t.indentMode = false
}

func (t *indenter) newLine(w writer) {
	w.WriteByte('\n')
	w.WriteString(t.prefix)
	for i := 0; i < t.depth; i++ {
		w.WriteString(t.indent)
	}
}

func (t *indenter) listStart(w writer, count int) {
	w.WriteByte('{')
	if t.indentMode {
		if count > 0 {
			t.depth++
			t.newLine(w)
		}
	}
}

func (t *indenter) listSep(w writer) {
	if t.indentMode {
		w.WriteByte(',')
		t.newLine(w)
	} else {
		w.WriteString(", ")
	}
}

func (t *indenter) listEnd(w writer, count int) {
	if t.indentMode {
		if count > 0 {
			t.depth--
			w.WriteByte(',')
			t.newLine(w)
		}
	}
	w.WriteByte('}')
}
