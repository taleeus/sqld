package sqld_legacy

import (
	"database/sql/driver"
	"testing"
)

func TestPgPrepare(t *testing.T) {
	str := "?,?,?,?"
	args := []driver.Value{0, 0, 0, 0}
	if PgPrepare(str, args) != "$1,$2,$3,$4" {
		t.Fatal("Prepare failed")
	}
}
