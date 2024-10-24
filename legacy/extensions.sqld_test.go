package sqld_legacy

import (
	"slices"
	"testing"
)

type testModel struct {
	Hi    string
	Named string `db:"nameddd"`
}

func (testModel) TableName() string {
	return "TestModel"
}

func TestColumns(t *testing.T) {
	columns := TableColumns[testModel]()
	if len(columns) != 2 || !slices.Contains(columns, "TestModel.Hi") || !slices.Contains(columns, "TestModel.nameddd") {
		t.Fatal("wrong columns extracted")
	}
}
