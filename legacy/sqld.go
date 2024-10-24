package sqld_legacy

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

var ErrNoColumns = errors.New("no columns in statement")
var ErrNilVal = errors.New("value is nil")
var ErrNilColumnExpr = errors.New("column expression is nil")
var ErrArgNotSlice = errors.New("argument is not a slice")
var ErrEmptySlice = errors.New("slice is empty")
var ErrNoOps = errors.New("operations slice is empty")

// SqldFn is the type describing all callbacks used in the library.
type SqldFn func() (string, []driver.Value, error)

// New builds a `SqldFn` callback combining the provided operators.
//
// Example usage:
//
//	const query := sqld.New(
//		sqld.Select(
//			sqld.Columns(
//				"name",
//				"pizzas",
//			),
//		),
//		sqld.From(sqld.Just("Table")),
//		sqld.Where(
//			sqld.And(
//				sqld.IfNotNil(filters.Name,
//					sqld.Eq("name", filters.Name),
//				),
//				sqld.IfNotEmpty(filters.Pizzas,
//					sqld.In("pizzas", filters.Pizzas),
//				),
//			),
//		),
//		sqld.OrderBy(sqld.Desc(filters.OrderBy)),
//	)
func New(ops ...SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(ops) == 0 {
			return "", nil, fmt.Errorf("query: %w", ErrNoOps)
		}

		var sb strings.Builder
		vals := make([]driver.Value, 0)
		var errs error

		for _, fn := range ops {
			s, fnVals, err := fn()
			if err != nil {
				errs = errors.Join(errs, err)
			}

			if errs != nil {
				continue
			}

			sb.WriteString(s)
			sb.WriteRune('\n')

			if len(fnVals) != 0 {
				vals = append(vals, fnVals...)
			}
		}

		if errs != nil {
			return "", nil, fmt.Errorf("query:\n%w", errs)
		}

		return sb.String(), vals, nil
	}
}
