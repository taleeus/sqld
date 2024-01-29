# sqld
A Go library to build and manage dynamic queries. No external dependencies. Slightly inspired by [Drizzle ORM](https://orm.drizzle.team/)

# Scope of the project
The scope of `sqld` is to provide an easy way to organize your dynamic queries in re-usable components, not to validate said queries.
I suggest to use other tools like [SQLParser](https://github.com/blastrain/vitess-sqlparser) and a lot of e2e tests!

## Usage
```go
query := sqld.New(
	sqld.Select(
		sqld.Columns(
			"name",
			"pizzas",
		),
	),
	sqld.From(sqld.Just("Table")),
	sqld.Where(
		sqld.And( 
			sqld.IfNotNil(filters.Name,
				sqld.Eq("name", filters.Name),
			),
			sqld.IfNotEmpty(filters.Pizzas,
				sqld.In("pizzas", filters.Pizzas),
			),
		),
	),
	sqld.OrderBy(
		sqld.IfNotNil(filters.OrderBy,
			sqld.Desc(filters.OrderBy),
		),
	),
)

s, args, err := query()
```

## Glossary
- _Operators_: callbacks that have [this signature](./sqld.go#L17), used to build various parts of the query
- _Statements_: major "blocks" of the query, like a whole WHERE statement with all its conditions
- _Conditionals_: functions that return different operators depending on the inputs (usually boolean checks). Check [this file](./conditionals.sqld.go)

## Customize
You can provide your own operators: every function (anonymous or not) that has [this signature](./sqld.go#L17) can be used by the library.

Check [here](./sqld.go#L10) for the built-in errors.
