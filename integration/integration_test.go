package integration

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/taleeus/sqld"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed schema.sql
var schemaFile embed.FS

var ctx = context.Background()
var db *pgxpool.Pool

func TestMain(m *testing.M) {
	// init
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pgContainer := Must(postgres.Run(ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	))

	connStr := Must(pgContainer.ConnectionString(ctx, "sslmode=disable"))
	db = Must(pgxpool.New(context.Background(), connStr))

	// init schema
	schema := Must(schemaFile.ReadFile("schema.sql"))
	Must(db.Exec(ctx, string(schema)))

	// run tests
	exitVal := m.Run()

	// cleanup
	db.Close()
	NoErr(pgContainer.Terminate(ctx))

	os.Exit(exitVal)
}

func FuzzFilters(f *testing.F) {
	f.Fuzz(func(t *testing.T, id int, name string, createdAtMs int64, count int, shouldOrder bool) {
		params := make(sqld.Params)
		dynquery := fmt.Sprintf(`
			SELECT
				COUNT(id),
				name,
				created_at
			FROM model
			%s
			GROUP BY
				name,
				created_at
			%s`,
			sqld.Where(sqld.And(
				sqld.IfNotZero(id, &params, sqld.Eq("id")),
				sqld.Or(
					sqld.IfNotZero(sqld.FmtContains(name), &params, sqld.ILike("name")),
					sqld.IfNotZero(time.UnixMilli(createdAtMs), &params, sqld.Gte("created_at")),
				),
			)),
			sqld.Having(sqld.IfNotZero(count, &params, sqld.Gte("COUNT(*)"))),
		)
		if shouldOrder {
			dynquery += sqld.OrderBy(
				sqld.Asc("name"),
				sqld.Desc("created_at"),
			)
		}

		query, args, err := sqlx.Named(dynquery, params)
		if err != nil {
			t.Fatalf("sqlx named query generation failed\nerr: %s\ndynquery: %s\nparams: %v", err.Error(), dynquery, params)
		}
		query = sqlx.Rebind(sqlx.DOLLAR, query)

		rows, err := db.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("fuzzed query failed\nerr: %s\nquery: %s\nargs: %v", err.Error(), query, args)
		}
		defer rows.Close()
	})
}
