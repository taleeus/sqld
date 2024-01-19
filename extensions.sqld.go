package sqld

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type Model interface {
	TableName() string
}

// Columns extracts a list of columns from a `Model`, using sqlx `db` tags
// and falling back on field names
func Columns[M Model]() []string {
	var model M
	columns := make([]string, 0)

	typ := reflect.TypeOf(model)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		column := field.Tag.Get("db")
		if column == "" {
			column = field.Name
		}

		columns = append(columns, model.TableName()+"."+column)
	}

	return columns
}

// TableName is a generic proxy for `Model.TableName()`
func TableName[M Model]() string {
	var model M
	return model.TableName()
}

// TableNameOp is a generic proxy for `Model.TableName()`, formatted as a callback
func TableNameOp[M Model]() SqldFn {
	return func() (string, []driver.Value, error) {
		return TableName[M](), nil, nil
	}
}

// Column returns a combination of `Model.TableName()` and the provided column.
// Panics if the column is not present in the model
func Column[M Model](column string) string {
	fullColumn, err := ColumnErr[M](column)
	if err != nil {
		panic(err)
	}

	return fullColumn
}

// ColumnErr returns a combination of `Model.TableName()` and the provided column.
// Returns error if the column is not present in the model
func ColumnErr[M Model](column string) (string, error) {
	if !slices.Contains(Columns[M](), column) {
		return "", fmt.Errorf("column %s not present in model %T", column, *new(M))
	}

	return TableName[M]() + "." + column, nil
}
