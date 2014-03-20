// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
    "fmt"
)

%%{
    machine scanner;
    include "scanner.rl";

    access   m.;
    variable p   m.p;
    variable pe  len(m.data);
    variable eof m.eof();
}%%

func (m *fsm) exec() (tok *token, err error) {
    %% write exec;
    if m.cs == scanner_error {
        return nil, fmt.Errorf("parse error: %s", string(m.data[m.te:]))
    }
    return
}
