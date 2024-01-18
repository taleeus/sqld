package sqld

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

// Block builds a callback that just returns the provided strings.
// Use it for the "static" parts of your query, like SELECT and JOIN statements.
//
//	sqld.Block(`
//		SELECT
//			name,
//			pizzas
//		FROM Table`,
//	),
func Block(block string) SqldFn {
	return func() (string, []driver.Value, error) {
		return block, nil, nil
	}
}

// Select builds a callback that returns a SELECT statement with a concatenation of
// the provided columns.
//
//	sqld.Select(
//		"name",
//		"pizzas",
//	)
func Select(columns ...string) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(columns) == 0 {
			return "", nil, fmt.Errorf("select: %w", ErrNoColumns)
		}

		return "SELECT\n\t" + strings.Join(columns, ",\n\t"), nil, nil
	}
}

// Not negates the provided operator.
func Not(op SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		s, vals, err := op()
		if err != nil {
			return "", nil, fmt.Errorf("not: %w", err)
		}

		if s == "" {
			return "", nil, nil
		}

		return "NOT(" + s + ")", vals, nil
	}
}

// Eq builds a callback that compares a column with the provided value.
//
//	sqld.Eq("name", filters.Name)
func Eq[T driver.Value](columnExpr string, val *T) SqldFn {
	return func() (string, []driver.Value, error) {
		if val == nil {
			return "", nil, fmt.Errorf("eq (%s): %w", columnExpr, ErrNilVal)
		}

		return columnExpr + " = ?", []driver.Value{val}, nil
	}
}

// Eq builds a callback that checks if a column is NULL.
//
//	sqld.Null("name")
func Null(columnExpr string) SqldFn {
	return func() (string, []driver.Value, error) {
		return columnExpr + " IS NULL", nil, nil
	}
}

// In builds a callback that checks if a column value is contained in the provided slice of values.
//
//	sqld.In("pizzas", filters.Pizzas)
func In[T driver.Value](columnExpr string, vals []T) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(vals) == 0 {
			return "", nil, fmt.Errorf("in (%s): %w", columnExpr, ErrEmptySlice)
		}

		return columnExpr + " IN (" + strings.Repeat(", ?", len(vals))[1:] + ")", mapSlice(vals), nil
	}
}

func boolCond(cond string, ops ...SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(ops) == 0 {
			return "", nil, fmt.Errorf("%s: %w", strings.ToLower(cond), ErrNoOps)
		}

		var sb strings.Builder
		vals := make([]driver.Value, 0, len(ops))
		var errs error

		atLeastOne := false
		for _, fn := range ops {
			s, fnVals, err := fn()
			if err != nil {
				errs = errors.Join(errs, err)
			}

			if errs != nil || s == "" {
				continue
			}

			if atLeastOne {
				sb.WriteString(cond + " ")
			}
			sb.WriteString(s)
			sb.WriteRune('\n')

			vals = append(vals, fnVals)
			atLeastOne = true
		}

		if errs != nil {
			return "", nil, fmt.Errorf("%s: %w", cond, errs)
		}

		return sb.String(), vals, nil
	}
}

// And builds a callback combining all the operators with AND conditions.
//
//	sqld.And(
//		sqld.IfNotNil(filters.Name,
//			sqld.Eq("name", filters.Name),
//		),
//		sqld.IfNotEmpty(filters.Pizzas,
//			sqld.In("pizzas", filters.Pizzas),
//		),
//	)
func And(ops ...SqldFn) SqldFn {
	return boolCond("AND", ops...)
}

// Or builds a callback combining all the operators with OR conditions.
//
//	sqld.Or(
//		sqld.IfNotNil(filters.Name,
//			sqld.Eq("name", filters.Name),
//		),
//		sqld.IfNotEmpty(filters.Pizzas,
//			sqld.In("pizzas", filters.Pizzas),
//		),
//	)
func Or(ops ...SqldFn) SqldFn {
	return boolCond("OR", ops...)
}

// Where builds a callback combining all the operators in a WHERE statement.
//
//	sqld.Where(
//		sqld.And(
//			sqld.IfNotNil(filters.Name,
//				sqld.Eq("name", filters.Name),
//			),
//			sqld.IfNotEmpty(filters.Pizzas,
//				sqld.In("pizzas", filters.Pizzas),
//			),
//		),
//	)
func Where(ops ...SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(ops) == 0 {
			return "", nil, fmt.Errorf("where: %w", ErrNoOps)
		}

		var sb strings.Builder
		vals := make([]driver.Value, 0, len(ops))
		var errs error

		for _, fn := range ops {
			s, fnVals, err := fn()
			if err != nil {
				errs = errors.Join(errs, err)
			}

			if errs != nil || s == "" {
				continue
			}

			sb.WriteString("\t" + s)
			sb.WriteRune('\n')

			vals = append(vals, fnVals)
		}

		if errs != nil {
			return "", nil, fmt.Errorf("where:\n\t\t%w", errs)
		}

		s := sb.String()
		if s == "" {
			return "", nil, nil
		}

		return "WHERE\n" + sb.String(), vals, nil
	}
}

// OrderBy builds a callback combining all the operators in a ORDER BY statement.
//
//	sqld.OrderBy(
//		sqld.IfNotNil(filters.OrderBy,
//			sqld.Desc(filters.OrderBy),
//		),
//	)
func OrderBy(ops ...SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(ops) == 0 {
			return "", nil, fmt.Errorf("orderBy: %w", ErrNoOps)
		}

		var sb strings.Builder
		vals := make([]driver.Value, 0, len(ops))
		var errs error

		for i, fn := range ops {
			s, fnVals, err := fn()
			if err != nil {
				errs = errors.Join(errs, err)
			}

			if errs != nil || s == "" {
				continue
			}

			sb.WriteString(s)
			if i != len(ops)-1 {
				sb.WriteRune(',')
			}
			sb.WriteRune('\n')

			vals = append(vals, fnVals)
		}

		if errs != nil {
			return "", nil, fmt.Errorf("orderBy:\n\t\t%w", errs)
		}

		s := sb.String()
		if s == "" {
			return "", nil, nil
		}

		return "ORDER BY\n" + sb.String(), vals, nil
	}
}

// Asc builds a callback used to specify the sorting in `OrderBy()`.
func Asc(columnExpr *string) SqldFn {
	return func() (string, []driver.Value, error) {
		if columnExpr == nil {
			return "", nil, fmt.Errorf("asc: %w", ErrNilColumnExpr)
		}

		return *columnExpr + " ASC", nil, nil
	}
}

// Desc builds a callback used to specify the sorting in `OrderBy()`.
func Desc(columnExpr *string) SqldFn {
	return func() (string, []driver.Value, error) {
		if columnExpr == nil {
			return "", nil, fmt.Errorf("asc: %w", ErrNilColumnExpr)
		}

		return *columnExpr + " DESC", nil, nil
	}
}

// Having builds a callback combining all the operators in a HAVING statement.
//
//	sqld.Having(
//		sqld.And(
//			sqld.IfNotNil(filters.Name,
//				sqld.Eq("name", filters.Name),
//			),
//			sqld.IfNotEmpty(filters.Pizzas,
//				sqld.In("pizzas", filters.Pizzas),
//			),
//		),
//	)
func Having(ops ...SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(ops) == 0 {
			return "", nil, fmt.Errorf("having: %w", ErrNoOps)
		}

		var sb strings.Builder
		vals := make([]driver.Value, 0, len(ops))
		var errs error

		for _, fn := range ops {
			s, fnVals, err := fn()
			if err != nil {
				errs = errors.Join(errs, err)
			}

			if errs != nil || s == "" {
				continue
			}

			sb.WriteString("\t" + s)
			sb.WriteRune('\n')

			vals = append(vals, fnVals)
		}

		if errs != nil {
			return "", nil, fmt.Errorf("having:\n\t\t%w", errs)
		}

		s := sb.String()
		if s == "" {
			return "", nil, nil
		}

		return "HAVING\n" + s, vals, nil
	}
}
