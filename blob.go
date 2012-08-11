// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite

/*
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Reader adapter to BLOB
type BlobReader struct {
	c          *Conn
	bl         *C.sqlite3_blob
	ReadOffset int
}

// ReadWriter adapter to BLOB
type BlobReadWriter struct {
	BlobReader
	WriteOffset int
}

// Zeroblobs are used to reserve space for a BLOB that is later written.
type ZeroBlobLength int

// NewBlobReader opens a BLOB for incremental I/O
//
// (See http://sqlite.org/c3ref/blob_open.html)
func (c *Conn) NewBlobReader(db, table, column string, row int64) (*BlobReader, error) {
	bl, err := c.blob_open(db, table, column, row, false)
	if err != nil {
		return nil, err
	}
	return &BlobReader{c, bl, 0}, nil
}

// NewBlobReadWriter open a BLOB for incremental I/O
// (See http://sqlite.org/c3ref/blob_open.html)
func (c *Conn) NewBlobReadWriter(db, table, column string, row int64) (*BlobReadWriter, error) {
	bl, err := c.blob_open(db, table, column, row, true)
	if err != nil {
		return nil, err
	}
	return &BlobReadWriter{BlobReader{c, bl, 0}, 0}, nil
}

func (c *Conn) blob_open(db, table, column string, row int64, write bool) (*C.sqlite3_blob, error) {
	zDb := C.CString(db)
	defer C.free(unsafe.Pointer(zDb))
	zTable := C.CString(table)
	defer C.free(unsafe.Pointer(zTable))
	zColumn := C.CString(column)
	defer C.free(unsafe.Pointer(zColumn))
	var bl *C.sqlite3_blob
	rv := C.sqlite3_blob_open(c.db, zDb, zTable, zColumn, C.sqlite3_int64(row), btocint(write), &bl)
	if rv != C.SQLITE_OK {
		if bl != nil {
			C.sqlite3_blob_close(bl)
		}
		return nil, c.error(rv)
	}
	if bl == nil {
		return nil, errors.New("sqlite succeeded without returning a blob")
	}
	return bl, nil
}

// Close closes a BLOB handle
// (See http://sqlite.org/c3ref/blob_close.html)
func (r *BlobReader) Close() error {
	if r == nil {
		return errors.New("nil sqlite blob reader")
	}
	rv := C.sqlite3_blob_close(r.bl)
	if rv != C.SQLITE_OK {
		return r.c.error(rv)
	}
	r.bl = nil
	return nil
}

// Read reads data from a BLOB incrementally
// (See http://sqlite.org/c3ref/blob_read.html)
func (r *BlobReader) Read(v []byte) (int, error) {
	var p *byte
	if len(v) > 0 {
		p = &v[0]
	}
	rv := C.sqlite3_blob_read(r.bl, unsafe.Pointer(p), C.int(len(v)), C.int(r.ReadOffset))
	if rv != C.SQLITE_OK {
		return 0, r.c.error(rv)
	}
	r.ReadOffset += len(v)
	return len(v), nil
}

// Size returns the size of an opened BLOB
// (See http://sqlite.org/c3ref/blob_bytes.html)
func (r *BlobReader) Size() (int, error) {
	s := C.sqlite3_blob_bytes(r.bl)
	return int(s), nil
}

// Write writes data into a BLOB incrementally
// (See http://sqlite.org/c3ref/blob_write.html)
func (w *BlobReadWriter) Write(v []byte) (int, error) {
	var p *byte
	if len(v) > 0 {
		p = &v[0]
	}
	rv := C.sqlite3_blob_write(w.bl, unsafe.Pointer(p), C.int(len(v)), C.int(w.WriteOffset))
	if rv != C.SQLITE_OK {
		return 0, w.c.error(rv)
	}
	w.WriteOffset += len(v)
	return len(v), nil
}

// Reopen moves a BLOB handle to a new row
// (See http://sqlite.org/c3ref/blob_reopen.html)
func (r *BlobReader) Reopen(rowid int64) error {
	rv := C.sqlite3_blob_reopen(r.bl, C.sqlite3_int64(rowid))
	if rv != C.SQLITE_OK {
		return r.c.error(rv)
	}
	r.ReadOffset = 0
	return nil
}
