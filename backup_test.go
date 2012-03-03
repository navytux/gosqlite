package sqlite_test

import (
	. "github.com/gwenn/gosqlite"
	"testing"
)

func TestBackup(t *testing.T) {
	dst := open(t)
	defer dst.Close()
	src := open(t)
	defer src.Close()
	fill(src, 1000)

	bck, err := NewBackup(dst, "main", src, "main")
	checkNoError(t, err, "couldn't init backup: %#v")

	cbs := make(chan BackupStatus)
	go func() {
		for {
			s := <-cbs
			t.Logf("Backup progress %#v\n", s)
		}
	}()
	err = bck.Run(10, 0, cbs)
	checkNoError(t, err, "couldn't do backup: %#v")
}