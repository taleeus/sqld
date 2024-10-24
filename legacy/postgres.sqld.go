package sqld_legacy

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// PgPrepare swaps all ? placeholders with postgres ones ($1, $2...)
func PgPrepare(query string, args []driver.Value) string {
	for i := 1; i <= len(args); i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}

	return query
}

// PgPrepareOp applies PgPrepare() to the resulting query in the operator.
// Use this as the last operator!
func PgPrepareOp(op SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		query, args, err := op()
		if err != nil {
			return "", nil, err
		}

		return PgPrepare(query, args), args, nil
	}
}
