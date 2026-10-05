package db

import "testing"

func TestSchemaMigrationsAreContiguousAndEmbedded(t *testing.T) {
	migrations, err := loadMigrations(migrationFiles)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) == 0 {
		t.Fatal("no migrations discovered")
	}
	for index := 1; index < len(migrations); index++ {
		if migrations[index-1].version >= migrations[index].version {
			t.Fatalf("migrations are not strictly ordered")
		}
	}
}
