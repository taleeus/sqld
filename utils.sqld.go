package sqld

import "database/sql/driver"

func mapSlice[T driver.Value](vals []T) []driver.Value {
	mappedVals := make([]driver.Value, 0, len(vals))
	for _, val := range vals {
		mappedVals = append(mappedVals, val)
	}

	return mappedVals
}
