package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/schemahub/backend/internal/pkg/testdb"
	schemadomain "github.com/schemahub/backend/internal/schema/domain"
)

func TestIntrospection_RealDatabase(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()

	// Seed tables, an index, an FK, and an enum in the shared public schema.
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS it_customers (
		id uuid PRIMARY KEY,
		name text NOT NULL,
		email varchar(320) UNIQUE
	)`)
	requireNoErr(t, err, "create it_customers")
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS it_orders (
		id uuid PRIMARY KEY,
		customer_id uuid REFERENCES it_customers(id),
		total numeric
	)`)
	requireNoErr(t, err, "create it_orders")
	_, err = pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS it_orders_total_idx ON it_orders (total)")
	requireNoErr(t, err, "create index")
	_, err = pool.Exec(ctx, "DO $$ BEGIN CREATE TYPE it_mood AS ENUM ('happy','sad'); EXCEPTION WHEN duplicate_object THEN NULL; END $$")
	requireNoErr(t, err, "create enum")

	t.Cleanup(func() {
		cctx := context.Background()
		_, _ = pool.Exec(cctx, "DROP TABLE IF EXISTS it_orders")
		_, _ = pool.Exec(cctx, "DROP TABLE IF EXISTS it_customers")
		_, _ = pool.Exec(cctx, "DROP TYPE IF EXISTS it_mood")
	})

	schemadomain.SetConnector(func(ctx context.Context, connStr string) (schemadomain.DBPool, error) {
		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			return nil, err
		}
		return &introPoolWrapper{pool: pool}, nil
	})
	svc := schemadomain.NewIntrospectionService()
	raw, err := svc.Introspect(ctx, testdb.URL(), []string{"public"})
	requireNoErr(t, err, "Introspect")

	var meta schemadomain.SchemaMetadata
	requireNoErr(t, json.Unmarshal(raw, &meta), "unmarshal metadata")

	var customers *schemadomain.TableInfo
	for i := range meta.Tables {
		if meta.Tables[i].Name == "it_customers" {
			customers = &meta.Tables[i]
			break
		}
	}
	if customers == nil {
		t.Fatalf("it_customers not found in introspection; tables = %d", len(meta.Tables))
	}

	colNames := map[string]bool{}
	for _, c := range customers.Columns {
		colNames[c.Name] = true
	}
	for _, want := range []string{"id", "name", "email"} {
		if !colNames[want] {
			t.Fatalf("column %s missing from it_customers: %+v", want, customers.Columns)
		}
	}

	if len(customers.Constr.PrimaryKey) != 1 || customers.Constr.PrimaryKey[0] != "id" {
		t.Fatalf("PK = %+v, want [id]", customers.Constr.PrimaryKey)
	}
	foundUnique := false
	for _, u := range customers.Constr.Uniques {
		if u == "email" {
			foundUnique = true
		}
	}
	if !foundUnique {
		t.Fatalf("unique constraint on email missing: %+v", customers.Constr.Uniques)
	}

	var orders *schemadomain.TableInfo
	for i := range meta.Tables {
		if meta.Tables[i].Name == "it_orders" {
			orders = &meta.Tables[i]
			break
		}
	}
	if orders == nil {
		t.Fatal("it_orders not found in introspection")
	}
	if len(orders.Constr.ForeignKeys) != 1 ||
		orders.Constr.ForeignKeys[0].RefTable != "it_customers" ||
		orders.Constr.ForeignKeys[0].RefColumn != "id" {
		t.Fatalf("FK = %+v, want ref it_customers.id", orders.Constr.ForeignKeys)
	}
	hasIndex := false
	for _, ix := range orders.Indexes {
		if ix.Name == "it_orders_total_idx" {
			hasIndex = true
		}
	}
	if !hasIndex {
		t.Fatalf("index it_orders_total_idx missing: %+v", orders.Indexes)
	}

	foundEnum := false
	for _, e := range meta.Enums {
		if e.Name == "it_mood" {
			foundEnum = true
			if len(e.Values) != 2 || e.Values[0] != "happy" || e.Values[1] != "sad" {
				t.Fatalf("enum values = %+v", e.Values)
			}
		}
	}
	if !foundEnum {
		t.Fatalf("enum it_mood missing: %+v", meta.Enums)
	}
}

type introPoolWrapper struct {
	pool *pgxpool.Pool
}

func (w *introPoolWrapper) Query(ctx context.Context, sql string, args ...any) (schemadomain.Rows, error) {
	return w.pool.Query(ctx, sql, args...)
}

func (w *introPoolWrapper) Close() {
	w.pool.Close()
}
