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

func Column[M Model](column string) SqldFn {
	return func() (string, []driver.Value, error) {
		if !slices.Contains(Columns[M](), column) {
			return "", nil, fmt.Errorf("column %s not present in model %T", column, *new(M))
		}

		return TableName[M]() + "." + column, nil, nil
	}
}
