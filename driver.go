// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"log"
	"os"
)

func init() {
	sql.Register("sqlite3", &Driver{})
	if os.Getenv("SQLITE_LOG") != "" {
		ConfigLog(func(d interface{}, err error, msg string) {
			log.Printf("%s: %s, %s\n", d, err, msg)
		}, "SQLITE")
	}
}

// Adapter to database/sql/driver
type Driver struct {
}
type connImpl struct {
	c *Conn
}
type stmtImpl struct {
	s *Stmt
}
type rowsImpl struct {
	s           *Stmt
	columnNames []string // cache
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	c, err := Open(name)
	if err != nil {
		return nil, err
	}
	c.BusyTimeout(500)
	return &connImpl{c}, nil
}

// PRAGMA schema_version may be used to detect when the database schema is altered

func (c *connImpl) Exec(query string, args []driver.Value) (driver.Result, error) {
	// http://code.google.com/p/go-wiki/wiki/InterfaceSlice
	tmp := make([]interface{}, len(args))
	for i, arg := range args {
		tmp[i] = arg
	}
	if err := c.c.Exec(query, tmp...); err != nil {
		return nil, err
	}
	return c, nil // FIXME RowAffected/noRows
}

// TODO How to know that the last Stmt has done an INSERT? An authorizer?
func (c *connImpl) LastInsertId() (int64, error) {
	return c.c.LastInsertRowid(), nil
}

// TODO How to know that the last Stmt has done a DELETE/INSERT/UPDATE? An authorizer?
func (c *connImpl) RowsAffected() (int64, error) {
	return int64(c.c.Changes()), nil
}

func (c *connImpl) Prepare(query string) (driver.Stmt, error) {
	s, err := c.c.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &stmtImpl{s}, nil
}

func (c *connImpl) Close() error {
	return c.c.Close()
}

func (c *connImpl) Begin() (driver.Tx, error) {
	if err := c.c.Begin(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *connImpl) Commit() error {
	return c.c.Commit()
}
func (c *connImpl) Rollback() error {
	return c.c.Rollback()
}

func (s *stmtImpl) Close() error {
	return s.s.Finalize()
}

func (s *stmtImpl) NumInput() int {
	return s.s.BindParameterCount()
}

func (s *stmtImpl) Exec(args []driver.Value) (driver.Result, error) {
	if err := s.bind(args); err != nil {
		return nil, err
	}
	if err := s.s.exec(); err != nil {
		return nil, err
	}
	return s, nil // FIXME RowAffected/noRows
}

// TODO How to know that this Stmt has done an INSERT? An authorizer?
func (s *stmtImpl) LastInsertId() (int64, error) {
	return s.s.c.LastInsertRowid(), nil
}

// TODO How to know that this Stmt has done a DELETE/INSERT/UPDATE? An authorizer?
func (s *stmtImpl) RowsAffected() (int64, error) {
	return int64(s.s.c.Changes()), nil
}

func (s *stmtImpl) Query(args []driver.Value) (driver.Rows, error) {
	if err := s.bind(args); err != nil {
		return nil, err
	}
	return &rowsImpl{s.s, nil}, nil
}

func (s *stmtImpl) bind(args []driver.Value) error {
	for i, v := range args {
		if err := s.s.BindByIndex(i+1, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *rowsImpl) Columns() []string {
	if r.columnNames == nil {
		r.columnNames = r.s.ColumnNames()
	}
	return r.columnNames
}

func (r *rowsImpl) Next(dest []driver.Value) error {
	ok, err := r.s.Next()
	if err != nil {
		return err
	}
	if !ok {
		return io.EOF
	}
	for i := range dest {
		dest[i] = r.s.ScanValue(i)
	}
	return nil
}

func (r *rowsImpl) Close() error {
	return r.s.Reset()
}
