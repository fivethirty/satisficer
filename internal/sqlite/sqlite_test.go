package sqlite_test

import (
	"testing"

	"github.com/fivethirty/static/internal/sqlite"
)

func TestNew(t *testing.T) {
	sqlite, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}

	var count int
	row := sqlite.DB.QueryRow("SELECT COUNT(*) FROM content")
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows, got %d", count)
	}
}
