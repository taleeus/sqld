package sqld

import (
	"fmt"
	"strconv"
	"strings"
)

// Op is a boolean operator
type Op string

const (
	OR  Op = "OR"
	AND Op = "AND"
)

// Sect is a filtering section of the query
type Sect string

const (
	WHERE  Sect = "WHERE"
	HAVING Sect = "HAVING"
)

// Sorting is the direction in which you want to sort a query
type Sorting string

const (
	ASC  Sorting = "ASC"
	DESC Sorting = "DESC"
)

// Section builds a filtering section.
// If the condition is empty, the returned string is also empty.
func Section(sect Sect, cond string) string {
	if cond == "" {
		return ""
	}

	return "\n" + string(sect) + " " + cond
}

// Where builds a WHERE filtering section.
// If the condition is empty, the returned string is also empty.
func Where(pred string) string {
	return Section(WHERE, pred)
}

// Having builds an HAVING filtering section.
// If the condition is empty, the returned string is also empty.
func Having(pred string) string {
	return Section(HAVING, pred)
}

// Cond builds a condition concatenating the filters with the given operator.
// If the filters are all empty, the returned string is also empty.
func Cond(op Op, filters ...string) string {
	bldr := strings.Builder{}
	for _, filter := range filters {
		if filter == "" {
			continue
		}

		if bldr.Len() != 0 {
			bldr.WriteString(" " + string(op) + "\n\t")
		}

		bldr.WriteString(filter)
	}

	if bldr.Len() == 0 {
		return ""
	}

	return "(\n\t" + bldr.String() + "\n)"
}

// And builds a condition concatenating the filters with the AND operator.
// If the filters are all empty, the returned string is also empty.
func And(filters ...string) string {
	return Cond(AND, filters...)
}

// Or builds a condition concatenating the filters with the OR operator.
// If the filters are all empty, the returned string is also empty.
func Or(filters ...string) string {
	return Cond(OR, filters...)
}

// Not negates the given string.
// If the filter is empty, the returned string is also empty.
func Not(filter string) string {
	if filter == "" {
		return ""
	}

	return "NOT(" + filter + ")"
}

// Sort produces a sorting statement on the column, with the given direction
func Sort(column string, sorting Sorting) string {
	return column + " " + string(sorting)
}

// Asc produces an ASC sorting statement on the column, with the given direction
func Asc(column string) string {
	return Sort(column, ASC)
}

// Desc produces a DESC sorting statement on the column, with the given direction
func Desc(column string) string {
	return Sort(column, DESC)
}

// OrderBy builds an ORDER BY section.
// If the sortings are all empty, the returned string is also empty.
func OrderBy(sorts ...string) string {
	bldr := strings.Builder{}
	for _, sort := range sorts {
		if sort == "" {
			continue
		}

		if bldr.Len() > 0 {
			bldr.WriteString(",\n\t")
		}

		bldr.WriteString(sort)
	}

	if bldr.Len() == 0 {
		return ""
	}

	return "\nORDER BY " + bldr.String()
}

// Null produces a filter that checks if the target is NULL
func Null(target string) string {
	return target + " IS NULL"
}

// PrinterFn is a callback that applies a parameter to the given statement (usually a filter)
type PrinterFn func(string) string

// Eq produces a PrinterFn that equates the target with the given parameter
func Eq(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s = :%s", target, param)
	}
}

// Like produces a PrinterFn that checks if the target text respects the given pattern
func Like(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s LIKE :%s", target, param)
	}
}

// ILike produces a PrinterFn that checks if the target text respects the given pattern, ignoring the casing
func ILike(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(:%s)", target, param)
	}
}

// ILike produces a PrinterFn that checks if the target text respects the given pattern, ignoring the casing
//
// Instead of lowering the two strings, it uses Postgres ILIKE operand
func PgILike(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s ILIKE :%s", target, param)
	}
}

// In produces a PrinterFn that checks if the target is contained in the given parameter slice
func In(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s IN(:%s)", target, param)
	}
}

// Gt produces a PrinterFn that checks if the target is greater than the given parameter
func Gt(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s > :%s", target, param)
	}
}

// Gte produces a PrinterFn that checks if the target is greater or equal the given parameter
func Gte(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s >= :%s", target, param)
	}
}

// Lt produces a PrinterFn that checks if the target is smaller than the given parameter
func Lt(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s < :%s", target, param)
	}
}

// Lte produces a PrinterFn that checks if the target is smaller or equal the given parameter
func Lte(target string) PrinterFn {
	return func(param string) string {
		return fmt.Sprintf("%s <= :%s", target, param)
	}
}

// FmtStartsWith maps the parameter with the desired pattern.
// Skips the mapping if the value is empty or nil
func FmtStartsWith[S string | *string](val S) S {
	if cast, ok := any(val).(string); ok {
		if cast == "" {
			return val
		}

		return any(cast + "%").(S)
	} else if cast, ok := any(val).(*string); ok {
		if cast == nil {
			return val
		}

		str := *cast + "%"
		return any(&str).(S)
	} else {
		panic("unreachable")
	}
}

// FmtEndsWith maps the parameter with the desired pattern.
// Skips the mapping if the value is empty or nil
func FmtEndsWith[S string | *string](val S) S {
	if cast, ok := any(val).(string); ok {
		if cast == "" {
			return val
		}

		return any("%" + cast).(S)
	} else if cast, ok := any(val).(*string); ok {
		if cast == nil {
			return val
		}

		str := "%" + *cast
		return any(&str).(S)
	} else {
		panic("unreachable")
	}
}

// FmtContains maps the parameter with the desired pattern.
// Skips the mapping if the value is empty or nil
func FmtContains[S string | *string](val S) S {
	if cast, ok := any(val).(string); ok {
		if cast == "" {
			return val
		}

		return any("%" + cast + "%").(S)
	} else if cast, ok := any(val).(*string); ok {
		if cast == nil {
			return val
		}

		str := "%" + *cast + "%"
		return any(&str).(S)
	} else {
		panic("unreachable")
	}
}

// Params is just an alias for a map containing the query parameters
type Params map[string]any

// Predicate is a callback that validates a condition on a value
type PredicateFn[T any] func(T) bool

// If is used to build the query dynamically, based on runtime conditions.
//
// If the predicate is true, the value is pushed in the parameter map and the printed filter is returned.
// If the predicate is false, the parameter map is untouched, and an empty string is returned.
func If[T any](pred PredicateFn[T], val T, params *Params, printer PrinterFn) string {
	if !pred(val) {
		return ""
	}

	argName := "arg" + strconv.Itoa(len(*params))
	(*params)[argName] = val

	return printer(argName)
}

// IfNotNil is a proxy for If with a predicate that checks if the pointer is not nil
func IfNotNil[T any](val *T, params *Params, printer PrinterFn) string {
	return If(func(t *T) bool {
		return t != nil
	}, val, params, printer)
}

// IfNotZero is a proxy for If with a predicate that checks if the value is not equal to the zero value of its type
func IfNotZero[T comparable](val T, params *Params, printer PrinterFn) string {
	return If(func(t T) bool {
		var zero T
		return t != zero
	}, val, params, printer)
}

// IfNotEmpty is a proxy for If with a predicate that checks if the slice is not empty
func IfNotEmpty[T any](val []T, params *Params, printer PrinterFn) string {
	return If(func(t []T) bool {
		return len(t) > 0
	}, val, params, printer)
}
