package sqld

import (
	"fmt"
	"reflect"
	"slices"
)

type Model interface {
	TableName() string
}

// TableColumns extracts a list of columns from a `Model`, using sqlx `db` tags
// and falling back on field names
func TableColumns[M Model]() []string {
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

// TableColumn returns a combination of `Model.TableName()` and the provided column.
// Panics if the column is not present in the model
func TableColumn[M Model](column string) string {
	fullColumn, err := TableColumnErr[M](column)
	if err != nil {
		panic(err)
	}

	return fullColumn
}

// TableColumnErr returns a combination of `Model.TableName()` and the provided column.
// Returns error if the column is not present in the model
func TableColumnErr[M Model](column string) (string, error) {
	if !slices.Contains(TableColumns[M](), column) {
		return "", fmt.Errorf("column %s not present in model %T", column, *new(M))
	}

	return TableName[M]() + "." + column, nil
}
