package sqld

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

func PgPrepare(query string, args []driver.Value) string {
	for i := 1; i <= len(args); i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}

	return query
}
