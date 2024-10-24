package sqld_legacy

import (
	"database/sql/driver"
)

func NoOp() (string, []driver.Value, error) {
	return "", nil, nil
}

func IfElse(pred func() bool, trueFn SqldFn, falseFn SqldFn) SqldFn {
	if pred() {
		return trueFn
	}

	return falseFn
}

func If(pred func() bool, op SqldFn) SqldFn {
	return IfElse(pred, op, NoOp)
}

func IfNilElse[T driver.Value](val *T, trueFn SqldFn, falseFn SqldFn) SqldFn {
	return IfElse(func() bool { return val == nil }, trueFn, falseFn)
}

func IfNotNilElse[T driver.Value](val *T, trueFn SqldFn, falseFn SqldFn) SqldFn {
	return IfElse(func() bool { return val != nil }, trueFn, falseFn)
}

func IfEmptyElse[T driver.Value](vals []T, trueFn SqldFn, falseFn SqldFn) SqldFn {
	return IfElse(func() bool { return len(mapSlice(vals)) == 0 }, trueFn, falseFn)
}

func IfNotEmptyElse[T driver.Value](vals []T, trueFn SqldFn, falseFn SqldFn) SqldFn {
	return IfElse(func() bool { return len(mapSlice(vals)) != 0 }, trueFn, falseFn)
}

func IfNil[T driver.Value](val *T, op SqldFn) SqldFn {
	return IfNilElse(val, op, NoOp)
}

func IfNotNil[T driver.Value](val *T, op SqldFn) SqldFn {
	return IfNotNilElse(val, op, NoOp)
}

func IfEmpty[T driver.Value](vals []T, op SqldFn) SqldFn {
	return IfEmptyElse(vals, op, NoOp)
}

func IfNotEmpty[T driver.Value](vals []T, op SqldFn) SqldFn {
	return IfNotEmptyElse(vals, op, NoOp)
}

func IfStringEmptyElse(val string, trueFn, falseFn SqldFn) SqldFn {
	return IfElse(func() bool { return val == "" }, trueFn, falseFn)
}

func IfStringNotEmptyElse(val string, trueFn, falseFn SqldFn) SqldFn {
	return IfElse(func() bool { return val != "" }, trueFn, falseFn)
}

func IfStringEmpty(val string, op SqldFn) SqldFn {
	return IfStringEmptyElse(val, op, NoOp)
}

func IfStringNotEmpty(val string, op SqldFn) SqldFn {
	return IfStringNotEmptyElse(val, op, NoOp)
}
