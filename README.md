# sqld
A Go library to build and manage dynamic queries.
It tries to remain as simple as possible, leveraging [sqlx](https://github.com/jmoiron/sqlx) parameter annotation.

## Legazy version
See the `legacy` module and the corresponding [README](legacy/README.md).

## Scope of the project
The scope of `sqld` is to provide an easy way to organize your dynamic queries, not to validate said queries.
I suggest to use other tools like [SQLParser](https://github.com/blastrain/vitess-sqlparser) and a lot of e2e tests!

# Usage
```go
func testQuery(db sqlx.ExtContext, args QueryArgs) ([]TestModel, error) {
	dynquery := `
	SELECT *
	FROM table
	`

	params := make(slqd.Params)	// it's just a map
	dynquery += sqld.Where(	// returns empty string if all predicates don't evaluate
		sqld.IfNotNil(args.Name, params, sqld.Contains("table.name"))
	)
	dynquery += sqld.IfNotNil(args.Limit, params, sqld.Limit)
	dynquery += sqld.IfNotNil(args.Offset, params, sqld.Offset)

	query, args, err := sqlx.Named(dynquery, params)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, db.Rebind(query), args...)	// use Rebind() with postgres
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TestModel
	for rows.Next() {
		var item TestModel
		if err := rows.StructScan(&item); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}
```
