package sqlds

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data/sqlutil"

	_ "github.com/go-sql-driver/mysql"
)

type testArgs struct {
	MySQLURL            string
	RunIntegrationTests bool
}

func testEnvArgs(t *testing.T) testArgs {
	t.Helper()
	var args testArgs
	if val, ok := os.LookupEnv("MYSQL_URL"); ok {
		args.MySQLURL = val
	} else {
		args.MySQLURL = "mysql:mysql@/mysql"
	}

	if _, ok := os.LookupEnv("INTEGRATION_TESTS"); ok {
		args.RunIntegrationTests = true
	}

	return args
}

func TestQuery_MySQL(t *testing.T) {
	var (
		args = testEnvArgs(t)
		ctx  = context.Background()
	)

	if !args.RunIntegrationTests {
		t.SkipNow()
	}

	db, err := sql.Open("mysql", args.MySQLURL)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	t.Run("The query should return a context.Canceled if it exceeds the timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		q := &Query{
			RawSQL: "SELECT SLEEP(5)",
		}

		_, err := query(ctx, db, []sqlutil.Converter{}, nil, q)
		if err == nil {
			t.Fatal("expected an error but received none")
		}
		if !(errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context deadline exceeded")) {
			t.Fatal("expected a context.Canceled error but received:", err)
		}
	})
}
