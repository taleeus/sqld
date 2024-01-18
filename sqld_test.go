package sqld

import (
	"testing"
)

type testFilters struct {
	Name    *string
	Pizzas  []string
	OrderBy *string
}

func buildTestQuery(filters testFilters) SqldFn {
	return New(
		Select(
			"name",
			"pizzas",
		),
		Block("FROM Table"),
		Where(
			Not(
				And(
					IfNotNil(filters.Name,
						Eq("name", filters.Name),
					),
					IfNotEmpty(filters.Pizzas,
						In("pizzas", filters.Pizzas),
					),
				),
			),
		),
		OrderBy(
			IfNotNil(filters.OrderBy,
				Desc(filters.OrderBy),
			),
		),
	)
}

func TestSqld(t *testing.T) {
	name := "test"
	orderBy := "name"
	filters := testFilters{
		Name:    &name,
		Pizzas:  []string{"margherita", "diavola", "4 stagioni"},
		OrderBy: &orderBy,
	}

	query := buildTestQuery(filters)
	s, _, err := query()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)

	filters.Name = nil
	query = buildTestQuery(filters)
	s, _, err = query()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)

	filters.Pizzas = nil
	query = buildTestQuery(filters)
	s, _, err = query()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)

	filters.OrderBy = nil
	query = buildTestQuery(filters)
	s, _, err = query()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}
