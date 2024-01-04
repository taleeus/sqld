# sqld
A pure-Go library to build and manage dynamic queries. Slightly inspired by [Drizzle ORM](https://orm.drizzle.team/)

## Usage
```go
query := sqld.New(
	sqld.Block(`
		SELECT
			name,
			pizzas
		FROM Table`,
	),
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

## Customize
You can provide your own operators: every function (anonymous or not) that has [this signature](./sqld.go#L17) can be used by the library.

Check [here](./sqld.go#L10) for the built-in errors.

## Out of the box
- Statements
	- [Block()](./operators.sqld.go#L19)
	- [Where()](./operators.sqld.go#L155)
	- [Having()](./operators.sqld.go#L277)
	- [OrderBy()](./operators.sqld.go#L201)
		- [Asc()](./operators.sqld.go#L44)
		- [Desc()](./operators.sqld.go#L255)
- Boolean operations
	- [Not()](./operators.sqld.go#L26)
	- [Eq()](./operators.sqld.go#L44)
	- [Null()](./operators.sqld.go#L57)
	- [In()](./operators.sqld.go#L66)
	- [And()](./operators.sqld.go#L125)
	- [Or()](./operators.sqld.go#L139)
- Conditionals
	- [NoOp()](./conditionals.sqld.go#L7)
	- [IfElse()](./conditionals.sqld.go#L11)
	- [If()](./conditionals.sqld.go#L19)
	- [IfNilElse()](./conditionals.sqld.go#L23)
	- [IfNotNilElse()](./conditionals.sqld.go#27)
	- [IfEmptyElse()](./conditionals.sqld.go#L31)
	- [IfNotEmptyElse()](./conditionals.sqld.go#L35)
	- [IfNil()](./conditionals.sqld.go#L39)
	- [IfNotNil()](./conditionals.sqld.go#L43)
	- [IfEmpty()](./conditionals.sqld.go#L47)
	- [IfNotEmpty()](./conditionals.sqld.go#L51)

## Glossary
- _Operators_: the callbacks that respect [this signature](./sqld.go#L17), used to build various parts of the query
- _Statements_: major "blocks" of the query, like a whole WHERE statement with all its conditions
- _Conditionals_: functions that return different operators depending on the inputs (usually boolean checks). Check [this file](./conditionals.sqld.go)
