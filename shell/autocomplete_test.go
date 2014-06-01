// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shell_test

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/gwenn/gosqlite"
	. "github.com/gwenn/gosqlite/shell"
)

func TestPragmaNames(t *testing.T) {
	pragmas := CompletePragma("fo")
	assert.Equalf(t, 3, len(pragmas), "got %d pragmas; expected %d", len(pragmas), 3)
	assert.Equal(t, []string{"foreign_key_check", "foreign_key_list(", "foreign_keys"}, pragmas, "unexpected pragmas")
}
func TestFuncNames(t *testing.T) {
	funcs := CompleteFunc("su")
	assert.Equal(t, 2, len(funcs), "got %d functions; expected %d", len(funcs), 2)
	assert.Equal(t, []string{"substr(", "sum("}, funcs, "unexpected functions")
}
func TestCmdNames(t *testing.T) {
	cmds := CompleteCmd(".h")
	assert.Equal(t, 2, len(cmds), "got %d commands; expected %d", len(cmds), 2)
	assert.Equal(t, []string{".headers", ".help"}, cmds, "unexpected commands")
}
func TestCache(t *testing.T) {
	db, err := sqlite.Open(":memory:")
	assert.Tf(t, err == nil, "%v", err)
	defer db.Close()
	cc := CreateCache(db)
	err = cc.Update(db)
	assert.Tf(t, err == nil, "%v", err)
}