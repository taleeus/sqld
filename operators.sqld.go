package sqld

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

// Just returns a callback that just returns the provided string
func Just(s string) SqldFn {
	return func() (string, []driver.Value, error) {
		return s, nil, nil
	}
}

// Columns builds a callback that returns a list of columns, comma-separated
func Columns(columns ...string) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(columns) == 0 {
			return "", nil, fmt.Errorf("columns: %w", ErrNoColumns)
		}

		return strings.Join(columns, ",\n\t"), nil, nil
	}
}

// Select builds a callback that returns a SELECT statement with a concatenation of
// the provided operators.
func Select(ops ...SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		if len(ops) == 0 {
			return "", nil, fmt.Errorf("select: %w", ErrNoOps)
		}

		columns, vals := "", make([]driver.Value, 0)
		for _, op := range ops {
			s, subVals, err := op()
			if err != nil {
				return "", nil, fmt.Errorf("select: %w", err)
			}

			columns = columns + ",\n\t" + s
			vals = append(vals, subVals)
		}

		return "SELECT\n\t" + columns, vals, nil
	}
}

// Count builds a callback that returns a COUNT function with the given argument
func Count(op SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		s, vals, err := op()
		if err != nil {
			return "", nil, fmt.Errorf("count: %w", err)
		}

		return "COUNT(" + s + ")", vals, nil
	}
}

// AllWildcard builds a callback that just returns a "*" string
func AllWildcard() SqldFn {
	return func() (string, []driver.Value, error) {
		return "*", nil, nil
	}
}

// From builds a callback that just returns a FROM statement with the provided table
func From(op SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		s, vals, err := op()
		if err != nil {
			return "", nil, fmt.Errorf("from: %w", err)
		}

		return "FROM " + s, vals, nil
	}
}

type JoinType string

const (
	LEFT_JOIN        JoinType = "LEFT"
	RIGHT_JOIN       JoinType = "RIGHT"
	INNER_JOIN       JoinType = "INNER"
	CROSS_JOIN       JoinType = "CROSS"
	FULL_JOIN        JoinType = "FULL"
	LEFT_OUTER_JOIN  JoinType = "LEFT OUTER"
	RIGHT_OUTER_JOIN JoinType = "RIGHT OUTER"
	INNER_OUTER_JOIN JoinType = "INNER OUTER"
	CROSS_OUTER_JOIN JoinType = "CROSS OUTER"
	FULL_OUTER_JOIN  JoinType = "FULL OUTER"
)

// Join builds a callback that returns a JOIN statement of the provided type
// with the desired subject, with a condition callback
func Join(joinType JoinType, subject SqldFn, op SqldFn) SqldFn {
	return func() (string, []driver.Value, error) {
		subj, subjVals, err := subject()
		if err != nil {
			return "", nil, fmt.Errorf("%s join: %w", joinType, err)
		}

		cond, condVals, err := op()
		if err != nil {
			return "", nil, fmt.Errorf("%s join: %w", joinType, err)
		}

		return string(joinType) + " JOIN " + subj + " ON " + cond, append(subjVals, condVals...), nil
	}
}

// As builds a callback that returns an alias
func As(op SqldFn, aliasName string) SqldFn {
	return func() (string, []driver.Value, error) {
		s, vals, err := op()
		if err != nil {
			return "", nil, fmt.Errorf("as: %w", err)
		}

		return s + " AS " + aliasName, vals, nil
	}
}

// SubQuery builds a callback that returns a subquery
func SubQuery(op SqldFn, aliasName string) SqldFn {
	return func() (string, []driver.Value, error) {
		s, vals, err := op()
		if err != nil {
			return "", nil, fmt.Errorf("as: %w", err)
		}

		return fmt.Sprintf("(\n%s\n) AS %s", s, aliasName), vals, nil
	}
}

// LeftJoin is a shortcut for `Join()` with `LEFT_JOIN` type
func LeftJoin(subject SqldFn, op SqldFn) SqldFn {
	return Join(LEFT_JOIN, subject, op)
}

// RightJoin is a shortcut for `Join()` with `RIGHT_JOIN` type
func RightJoin(subject SqldFn, op SqldFn) SqldFn {
	return Join(RIGHT_JOIN, subject, op)
}

// ColumnEq builds a callback that returns a comparison statement between two columns
func ColumnEq(firstColumn string, secondColumn string) SqldFn {
	return func() (string, []driver.Value, error) {
		return firstColumn + " = " + secondColumn, nil, nil
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
			return "", nil, nil
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

func Limit(count *uint) SqldFn {
	return func() (string, []driver.Value, error) {
		if count == nil {
			return "", nil, nil
		}

		return "LIMIT ?", []driver.Value{*count}, nil
	}
}

func Offset(skip *uint) SqldFn {
	return func() (string, []driver.Value, error) {
		if skip == nil {
			return "", nil, nil
		}

		return "OFFSET ?", []driver.Value{*skip}, nil
	}
}
